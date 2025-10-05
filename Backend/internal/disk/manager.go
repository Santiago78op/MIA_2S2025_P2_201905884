package disk

import (
	"context"
	"crypto/rand"
	"encoding/binary"
)

type Manager interface {
	Mkdisk(ctx context.Context, path string, sizeBytes int64, fit string) error
	FdiskAdd(ctx context.Context, path, partName string, sizeBytes int64, ptype string, fit string) error
	FdiskDelete(ctx context.Context, path, partName string, mode string) error
	Mount(ctx context.Context, path, partName string) (PartitionRef, error)
	Unmount(ctx context.Context, ref PartitionRef) error
	ListMounted(ctx context.Context) ([]PartitionRef, error)
}

type FileManager struct {
	mounts *mountTable
}

func NewManager() *FileManager {
	return &FileManager{mounts: newMountTable()}
}

// -------------------- MKDISK --------------------
func (m *FileManager) Mkdisk(ctx context.Context, path string, sizeBytes int64, fit string) error {
	if sizeBytes <= 0 {
		return ErrInvalidParam
	}
	fitB := parseFit(fit)
	if err := ensureSize(path, sizeBytes); err != nil {
		return err
	}
	// Escribe MBR
	f, err := openRW(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sig := int32(randU32())
	mbr := NewMBR(sizeBytes, fitB, sig)
	if err := writeStruct(f, 0, &mbr); err != nil {
		return err
	}
	// Zero el área posterior (opcional: ya se truncó)
	return nil
}

// -------------------- FDISK ADD --------------------
func (m *FileManager) FdiskAdd(ctx context.Context, path, partName string, sizeBytes int64, ptype string, fit string) error {
	if sizeBytes <= 0 || partName == "" {
		return ErrInvalidParam
	}
	name, err := fmtName(partName)
	if err != nil {
		return err
	}

	f, err := openRW(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var mbr MBR
	if err := readStruct(f, 0, &mbr); err != nil {
		return err
	}

	switch normalizeType(ptype) {
	case PartTypePrimary, PartTypeExtended:
		// 1) hallar hueco libre entre primarias/extendida
		free := buildFreePrimaries(&mbr)
		if len(free) == 0 {
			return ErrNoSpace
		}
		seg, ok := pickByFit(free, sizeBytes, mbr.Fit)
		if !ok {
			return ErrNoSpace
		}
		// 2) tomar slot libre en Parts
		idx := -1
		for i := 0; i < MaxPrimaries; i++ {
			if mbr.Parts[i].Status != PartStatusUsed {
				idx = i
				break
			}
		}
		if idx == -1 {
			return ErrNoSpace
		}
		mbr.Parts[idx] = Partition{
			Status: PartStatusUsed,
			Type:   normalizeType(ptype),
			Fit:    parseFit(fit),
			Start:  seg.start,
			Size:   sizeBytes,
			Name:   name,
		}
		// 3) si es extendida, inicializa un EBR vacío al inicio
		if mbr.Parts[idx].Type == PartTypeExtended {
			ebr := EBR{Status: 0, Fit: FitFF, Start: seg.start, Size: 0, Next: -1}
			if err := writeStruct(f, seg.start, &ebr); err != nil {
				return err
			}
		}
		// 4) persistir MBR
		return writeStruct(f, 0, &mbr)

	case PartTypeLogical:
		// Requiere una extendida existente
		// 1) Buscar la partición extendida
		var extended *Partition
		for i := 0; i < MaxPrimaries; i++ {
			if mbr.Parts[i].Status == PartStatusUsed && mbr.Parts[i].Type == PartTypeExtended {
				extended = &mbr.Parts[i]
				break
			}
		}
		if extended == nil {
			return ErrNoExtended
		}

		// 2) Agregar partición lógica usando las funciones EBR
		if err := AddLogicalPartition(f, extended.Start, extended.Start+extended.Size, partName, sizeBytes, parseFit(fit)); err != nil {
			return err
		}

		return nil

	default:
		return ErrInvalidParam
	}
}

// -------------------- FDISK DELETE --------------------
func (m *FileManager) FdiskDelete(ctx context.Context, path, partName string, mode string) error {
	f, err := openRW(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var mbr MBR
	if err := readStruct(f, 0, &mbr); err != nil {
		return err
	}

	// buscar en primarias/extendida
	idx := -1
	for i := 0; i < MaxPrimaries; i++ {
		if mbr.Parts[i].Status == PartStatusUsed && trimName(mbr.Parts[i].Name) == partName {
			idx = i
			break
		}
	}
	if idx == -1 {
		// Buscar en particiones lógicas (EBRs)
		// Primero necesitamos encontrar la partición extendida
		var extended *Partition
		for i := 0; i < MaxPrimaries; i++ {
			if mbr.Parts[i].Status == PartStatusUsed && mbr.Parts[i].Type == PartTypeExtended {
				extended = &mbr.Parts[i]
				break
			}
		}

		if extended != nil {
			// Buscar y eliminar partición lógica
			fullDelete := mode == "full"
			if err := DeleteLogicalPartition(f, extended.Start, extended.Start+extended.Size, partName, fullDelete); err != nil {
				return err
			}
			return nil
		}

		return ErrNotFound
	}

	// limpiar rango si quieres (full/fast)
	if mode == "full" {
		if err := zeroRange(f, mbr.Parts[idx].Start, mbr.Parts[idx].Size); err != nil {
			return err
		}
	}
	mbr.Parts[idx] = Partition{} // marca libre
	return writeStruct(f, 0, &mbr)
}

// -------------------- MOUNT / UNMOUNT --------------------
func (m *FileManager) Mount(ctx context.Context, path, partName string) (PartitionRef, error) {
	f, err := openRW(path)
	if err != nil {
		return PartitionRef{}, err
	}
	defer f.Close()

	var mbr MBR
	if err := readStruct(f, 0, &mbr); err != nil {
		return PartitionRef{}, err
	}

	// primarias/extendida
	for i := 0; i < MaxPrimaries; i++ {
		p := mbr.Parts[i]
		if p.Status == PartStatusUsed && trimName(p.Name) == partName {
			ref := PartitionRef{DiskPath: path, PartitionID: partName}
			if err := m.mounts.put(path, partName, ref); err != nil {
				return PartitionRef{}, err
			}
			return ref, nil
		}
	}

	// Buscar en particiones lógicas (EBRs)
	// Primero encontrar la partición extendida
	var extended *Partition
	for i := 0; i < MaxPrimaries; i++ {
		if mbr.Parts[i].Status == PartStatusUsed && mbr.Parts[i].Type == PartTypeExtended {
			extended = &mbr.Parts[i]
			break
		}
	}

	if extended != nil {
		// Buscar en particiones lógicas
		_, ebrOffset, err := FindEBRByName(f, extended.Start, extended.Start+extended.Size, partName)
		if err == nil {
			// Encontrada partición lógica
			ref := PartitionRef{DiskPath: path, PartitionID: partName}
			if err := m.mounts.put(path, partName, ref); err != nil {
				return PartitionRef{}, err
			}
			return ref, nil
		}
		// Si no se encontró, continuar con error not found
		_ = ebrOffset // evitar warning de variable no usada
	}

	return PartitionRef{}, ErrNotFound
}

func (m *FileManager) Unmount(ctx context.Context, ref PartitionRef) error {
	_, ok := m.mounts.get(ref.DiskPath, ref.PartitionID)
	if !ok {
		return ErrNotMounted
	}
	m.mounts.del(ref.DiskPath, ref.PartitionID)
	return nil
}

func (m *FileManager) ListMounted(ctx context.Context) ([]PartitionRef, error) {
	return m.mounts.list(), nil
}

// -------------------- Helpers --------------------
func parseFit(s string) byte {
	switch stringsUpper(s) {
	case "FF":
		return FitFF
	case "BF":
		return FitBF
	case "WF":
		return FitWF
	default:
		return FitFF
	}
}
func normalizeType(s string) byte {
	switch stringsUpper(s) {
	case "P":
		return PartTypePrimary
	case "E":
		return PartTypeExtended
	case "L":
		return PartTypeLogical
	default:
		return 0
	}
}
func stringsUpper(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 32
		}
	}
	return string(b)
}
func trimName(n [NameLen]byte) string {
	i := 0
	for ; i < len(n); i++ {
		if n[i] == 0 {
			break
		}
	}
	return string(n[:i])
}

func randU32() uint32 {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return binary.LittleEndian.Uint32(b[:])
}
