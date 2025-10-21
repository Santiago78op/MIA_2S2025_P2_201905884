package diskio

import (
	"Backend/core/models"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// =============================
// Config / Constantes
// =============================

type JournalMode int

const (
	JournalConst50 JournalMode = iota // journalCount=50
	JournalPerN                       // journalCount=n (modo alterno)
)

// Tamaños lógicos (estables, no dependen de sizeof de structs)
const (
	SzSuperExt3      = int64(112) // SuperBlockExt3: 112 bytes (11×int32[44] + 9×int64[72])
	SzJournalEntry   = int64(600) // Journal: 600 bytes por entrada
	SzBitmapUnit     = int64(1)   // 1 byte por unidad en bitmaps
	SzInode          = int64(100) // Inodo: 100 bytes
	SzBlock          = int64(64)  // Bloque: 64 bytes
	DefaultJournal50 = int64(50)
)

// =============================
// Layout EXT3
// =============================

type Ext3Layout struct {
	PartStart    int64
	PartSize     int64
	N            int64 // número de inodos
	NB           int64 // número de bloques (3*n)
	JournalCount int64
	SBStart      int64
	JournalStart int64
	BmInodeStart int64
	BmBlockStart int64
	InodeStart   int64
	BlockStart   int64
	EndOffset    int64 // donde termina el área ocupada
}

// CalcNExt3 calcula n desde la ecuación del enunciado
// part = SB + (J * SzJournal) + n + 3n + (n*SzInode) + (3n*SzBlock)
func CalcNExt3(partSize int64, journalCount int64) int64 {
	if partSize <= SzSuperExt3 {
		return 0
	}
	// Coeficiente por n en bytes:
	coef := SzBitmapUnit + 3*SzBitmapUnit + SzInode + 3*SzBlock
	base := SzSuperExt3 + journalCount*SzJournalEntry
	if coef <= 0 || partSize <= base {
		return 0
	}
	n := (partSize - base) / coef
	if n < 0 {
		return 0
	}
	return n
}

// ComputeLayoutExt3 calcula offsets siguiendo el diagrama del enunciado
func ComputeLayoutExt3(partStart, partSize int64, mode JournalMode) (Ext3Layout, error) {
	if partSize <= 0 {
		return Ext3Layout{}, errors.New("partSize inválido")
	}

	journalCount := DefaultJournal50
	if mode == JournalPerN {
		journalCount = 0
	}

	n := CalcNExt3(partSize, journalCount)
	if mode == JournalPerN {
		journalCount = n
		// Recalcular n con el nuevo journalCount
		n = CalcNExt3(partSize, journalCount)
	}

	// Offsets absolutos
	SB := partStart
	J := SB + SzSuperExt3
	BMI := J + journalCount*SzJournalEntry
	BMB := BMI + n*SzBitmapUnit
	IN := BMB + (3 * n * SzBitmapUnit)
	BL := IN + n*SzInode
	end := BL + (3 * n * SzBlock)

	// Si por redondeos nos pasamos, reducir n hasta que quepa
	for end-partStart > partSize && n > 0 {
		n--
		if mode == JournalPerN {
			journalCount = n
		}
		BMI = J + journalCount*SzJournalEntry
		BMB = BMI + n*SzBitmapUnit
		IN = BMB + (3 * n * SzBitmapUnit)
		BL = IN + n*SzInode
		end = BL + (3 * n * SzBlock)
	}

	layout := Ext3Layout{
		PartStart:    partStart,
		PartSize:     partSize,
		N:            n,
		NB:           3 * n,
		JournalCount: journalCount,
		SBStart:      SB,
		JournalStart: J,
		BmInodeStart: BMI,
		BmBlockStart: BMB,
		InodeStart:   IN,
		BlockStart:   BL,
		EndOffset:    end,
	}

	if layout.EndOffset > partStart+partSize {
		return layout, fmt.Errorf("layout no cabe: end=%d > limit=%d", layout.EndOffset, partStart+partSize)
	}
	return layout, nil
}

// =============================
// SuperBloque EXT3 - escritura/lectura
// =============================

func (r *FileFsRepository) writeSuperExt3(f *os.File, sb models.SuperBlockExt3) error {
	buf := make([]byte, SzSuperExt3)
	off := 0
	put32 := func(v int32) { binary.LittleEndian.PutUint32(buf[off:off+4], uint32(v)); off += 4 }
	put64 := func(v int64) { binary.LittleEndian.PutUint64(buf[off:off+8], uint64(v)); off += 8 }

	put32(sb.FsType)
	put32(sb.InodesCount)
	put32(sb.BlocksCount)
	put32(sb.FreeBlocks)
	put32(sb.FreeInodes)
	put64(sb.MTime)
	put64(sb.UMTime)
	put32(sb.MntCount)
	put32(sb.Magic)
	put32(sb.InodeSize)
	put32(sb.BlockSize)
	put64(sb.FirstInode)
	put64(sb.FirstBlock)
	put64(sb.BmInodeStart)
	put64(sb.BmBlockStart)
	put64(sb.InodeStart)
	put64(sb.BlockStart)
	put64(sb.JournalStart)
	put32(sb.JournalCount)
	// Padding si queda - el buffer ya es de 80 bytes

	if _, err := f.Seek(sb.PartStart, 0); err != nil {
		return err
	}
	_, err := f.Write(buf)
	return err
}

func (r *FileFsRepository) readSuperExt3(f *os.File, partStart int64) (models.SuperBlockExt3, error) {
	buf := make([]byte, SzSuperExt3)
	if _, err := f.Seek(partStart, 0); err != nil {
		return models.SuperBlockExt3{}, err
	}
	if _, err := f.Read(buf); err != nil {
		return models.SuperBlockExt3{}, err
	}

	off := 0
	get32 := func() int32 { v := int32(binary.LittleEndian.Uint32(buf[off : off+4])); off += 4; return v }
	get64 := func() int64 { v := int64(binary.LittleEndian.Uint64(buf[off : off+8])); off += 8; return v }

	sb := models.SuperBlockExt3{}
	sb.FsType = get32()
	sb.InodesCount = get32()
	sb.BlocksCount = get32()
	sb.FreeBlocks = get32()
	sb.FreeInodes = get32()
	sb.MTime = get64()
	sb.UMTime = get64()
	sb.MntCount = get32()
	sb.Magic = get32()
	sb.InodeSize = get32()
	sb.BlockSize = get32()
	sb.FirstInode = get64()
	sb.FirstBlock = get64()
	sb.BmInodeStart = get64()
	sb.BmBlockStart = get64()
	sb.InodeStart = get64()
	sb.BlockStart = get64()
	sb.JournalStart = get64()
	sb.JournalCount = get32()
	sb.PartStart = partStart
	return sb, nil
}

// =============================
// MKFS(3fs) - inicialización
// =============================

// getMountInfo busca la información de montaje y partición
func (r *FileFsRepository) getMountInfo(id string) (diskPath string, partStart int64, partSize int64, err error) {
	// Buscar en la lista de montajes
	for _, m := range r.mounts.List() {
		if m.ID == id {
			// Leer MBR para obtener Start y Size de la partición
			mbr, e := r.disk.ReadMBR(m.Path)
			if e != nil {
				return "", 0, 0, e
			}

			// Buscar en particiones primarias/extendidas
			for _, p := range mbr.Partitions {
				pName := strings.TrimRight(string(p.Name[:]), "\x00 ")
				if p.Status == models.PartStatusUsed && pName == m.Name {
					return m.Path, p.Start, p.Size, nil
				}
			}

			// Si no se encontró, buscar en particiones lógicas (EBRs)
			for _, p := range mbr.Partitions {
				if p.Type == models.PartTypeExtend && p.Status == models.PartStatusUsed {
					// Buscar en la cadena de EBRs
					ebrOffset := p.Start
					extStart := p.Start
					f, err := os.Open(m.Path)
					if err != nil {
						continue
					}
					defer f.Close()

					for ebrOffset != -1 {
						var ebr models.EBR
						if _, err := f.Seek(ebrOffset, 0); err != nil {
							break
						}
						if err := binary.Read(f, binary.LittleEndian, &ebr); err != nil {
							break
						}

						ebrName := strings.TrimRight(string(ebr.Name[:]), "\x00 ")
						if ebr.Status == models.PartStatusUsed && ebrName == m.Name {
							return m.Path, ebr.Start, ebr.Size, nil
						}

						if ebr.Next == -1 {
							break
						}
						ebrOffset = extStart + ebr.Next
					}
				}
			}

			return "", 0, 0, fmt.Errorf("partición %s no encontrada en MBR", m.Name)
		}
	}
	return "", 0, 0, fmt.Errorf("id no montado: %s", id)
}

// MkfsWithTypeExt3 formatea la partición con EXT3
func (r *FileFsRepository) MkfsWithTypeExt3(id string, mode JournalMode) error {
	diskPath, partStart, partSize, err := r.getMountInfo(id)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	layout, err := ComputeLayoutExt3(partStart, partSize, mode)
	if err != nil {
		return err
	}

	// 1) Super bloque
	sb := models.SuperBlockExt3{
		FsType:       3,
		InodesCount:  int32(layout.N),
		BlocksCount:  int32(layout.NB),
		FreeInodes:   int32(layout.N - 2), // Reservados: root + users.txt
		FreeBlocks:   int32(layout.NB - 2),
		MTime:        time.Now().Unix(),
		UMTime:       0,
		MntCount:     1,
		Magic:        models.FSMagic,
		InodeSize:    int32(SzInode),
		BlockSize:    int32(SzBlock),
		FirstInode:   layout.InodeStart,
		FirstBlock:   layout.BlockStart,
		BmInodeStart: layout.BmInodeStart,
		BmBlockStart: layout.BmBlockStart,
		InodeStart:   layout.InodeStart,
		BlockStart:   layout.BlockStart,
		JournalStart: layout.JournalStart,
		JournalCount: int32(layout.JournalCount),
		PartStart:    layout.PartStart,
	}

	if err := r.writeSuperExt3(f, sb); err != nil {
		return err
	}

	// 2) Inicializar Journal a vacío
	if err := zeroRegion(f, layout.JournalStart, layout.JournalCount*SzJournalEntry); err != nil {
		return err
	}

	// 3) Bitmaps (0 = libre)
	if err := zeroRegion(f, layout.BmInodeStart, layout.N*SzBitmapUnit); err != nil {
		return err
	}
	if err := zeroRegion(f, layout.BmBlockStart, 3*layout.N*SzBitmapUnit); err != nil {
		return err
	}

	// 4) Tablas
	if err := zeroRegion(f, layout.InodeStart, layout.N*SzInode); err != nil {
		return err
	}
	if err := zeroRegion(f, layout.BlockStart, layout.NB*SzBlock); err != nil {
		return err
	}

	// 5) TODO: Crear raíz y users.txt (usando helpers de P1)
	// if err := r.bootstrapRootAndUsers(f, sb); err != nil { return err }

	return nil
}

// zeroRegion escribe 0x00 en el rango especificado
func zeroRegion(f *os.File, start, size int64) error {
	if size <= 0 {
		return nil
	}
	if _, err := f.Seek(start, 0); err != nil {
		return err
	}
	buf := make([]byte, 4096)
	remaining := size
	for remaining > 0 {
		chunk := int64(len(buf))
		if chunk > remaining {
			chunk = remaining
		}
		if _, err := f.Write(buf[:chunk]); err != nil {
			return err
		}
		remaining -= chunk
	}
	return nil
}

// ============================================================
// JOURNAL - Implementación según enunciado
// ============================================================

// AppendJournal escribe la siguiente entrada disponible
func (r *FileFsRepository) AppendJournal(f *os.File, sb models.SuperBlockExt3, j models.Journal) error {
	// Buscar el primer slot vacío o el más viejo si está lleno
	idx := int64(-1)
	oldestIdx := int64(0)
	oldestTs := float64(time.Now().Unix() + 1)

	for i := int64(0); i < int64(sb.JournalCount); i++ {
		off := sb.JournalStart + i*models.JournalEntrySize
		buf := make([]byte, models.JournalEntrySize)
		if _, err := f.ReadAt(buf, off); err != nil {
			return err
		}
		entry, _ := models.UnmarshalJournal(buf)
		if isEmptyOp(entry.Content.Op[:]) {
			idx = i
			break
		}
		// Rastrea el más viejo (menor fecha) para política FIFO
		if entry.Content.Date < oldestTs {
			oldestTs = entry.Content.Date
			oldestIdx = i
		}
	}

	if idx < 0 {
		idx = oldestIdx
	}

	// Numerar Count automáticamente si viene en 0
	if j.Count == 0 {
		j.Count = int32(idx + 1)
	}

	buf := j.Marshal()
	_, err := f.WriteAt(buf, sb.JournalStart+idx*models.JournalEntrySize)
	return err
}

// ListJournal devuelve todas las entradas no vacías en orden de tiempo
func (r *FileFsRepository) ListJournal(f *os.File, sb models.SuperBlockExt3) ([]models.Journal, error) {
	out := make([]models.Journal, 0, sb.JournalCount)
	for i := int64(0); i < int64(sb.JournalCount); i++ {
		off := sb.JournalStart + i*models.JournalEntrySize
		buf := make([]byte, models.JournalEntrySize)
		if _, err := f.ReadAt(buf, off); err != nil {
			return nil, err
		}
		entry, _ := models.UnmarshalJournal(buf)
		if !isEmptyOp(entry.Content.Op[:]) {
			out = append(out, entry)
		}
	}

	// Ordenar por fecha
	sort.Slice(out, func(i, j int) bool {
		return out[i].Content.Date < out[j].Content.Date
	})

	return out, nil
}

// ClearJournal rellena con entradas vacías
func (r *FileFsRepository) ClearJournal(f *os.File, sb models.SuperBlockExt3) error {
	empty := models.Journal{}
	buf := empty.Marshal()
	for i := int64(0); i < int64(sb.JournalCount); i++ {
		off := sb.JournalStart + i*models.JournalEntrySize
		if _, err := f.WriteAt(buf, off); err != nil {
			return err
		}
	}
	return nil
}

func isEmptyOp(op []byte) bool {
	return len(strings.Trim(string(op), "\x00 ")) == 0
}
