// Package: storage/diskio
// File   : disk_partition_ops.go
// Goal   : Wrappers de operaciones de particiones para FileDiskRepository

package diskio

import (
	"Backend/core/models"
	"os"
)

// ========================= WRAPPERS PARA FileDiskRepository =========================

// FindPartition busca una partición por nombre
func (r *FileDiskRepository) FindPartition(path string, name string) (PartitionRef, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		return PartitionRef{}, err
	}
	defer f.Close()

	var m models.MBR
	if err := readMBR(f, &m); err != nil {
		return PartitionRef{}, err
	}

	// Buscar en primarias y extendida
	for i := 0; i < 4; i++ {
		p := m.Partitions[i]
		if p.Status == models.PartStatusUsed && trimPartName(p.Name[:]) == name {
			ptype := TypePrimary
			if p.Type == models.PartTypeExtend {
				ptype = TypeExtended
			}
			return PartitionRef{
				Type:  ptype,
				Index: i,
				Start: p.Start,
				Size:  p.Size,
			}, nil
		}
	}

	// Buscar en lógicas
	extIdx := -1
	for i := 0; i < 4; i++ {
		if m.Partitions[i].Type == models.PartTypeExtend && m.Partitions[i].Status == models.PartStatusUsed {
			extIdx = i
			break
		}
	}

	if extIdx >= 0 {
		extStart := m.Partitions[extIdx].Start
		ebrOff := extStart
		for ebrOff != -1 && ebrOff != 0 {
			var e models.EBR
			if err := readEBR(f, &e, ebrOff); err != nil {
				break
			}
			if e.Status == models.PartStatusUsed && trimPartName(e.Name[:]) == name {
				return PartitionRef{
					Type:   TypeLogical,
					Index:  -1,
					Start:  e.Start,
					Size:   e.Size,
					EBROff: ebrOff,
				}, nil
			}
			ebrOff = e.Next
		}
	}

	return PartitionRef{}, models.ErrNoExtendedPartition
}

// NextPartitionAfter devuelve la partición siguiente
func (r *FileDiskRepository) NextPartitionAfter(path string, ref PartitionRef) (PartitionRef, error) {
	// Esta implementación es compleja, por ahora retornar error
	return PartitionRef{}, nil
}

// ReadPartition devuelve el start y size
func (r *FileDiskRepository) ReadPartition(path string, ref PartitionRef) (int64, int64, error) {
	return ref.Start, ref.Size, nil
}

// ResizePartition actualiza el tamaño
func (r *FileDiskRepository) ResizePartition(path string, ref PartitionRef, newSize int64) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if ref.Type == TypeLogical {
		var e models.EBR
		if err := readEBR(f, &e, ref.EBROff); err != nil {
			return err
		}
		e.Size = newSize
		return writeEBR(f, &e, ref.EBROff)
	}

	var m models.MBR
	if err := readMBR(f, &m); err != nil {
		return err
	}

	m.Partitions[ref.Index].Size = newSize
	return writeMBR(f, &m)
}

// DeletePartitionFast limpia la entrada
func (r *FileDiskRepository) DeletePartitionFast(path string, ref PartitionRef) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if ref.Type == TypeLogical {
		var e models.EBR
		if err := readEBR(f, &e, ref.EBROff); err != nil {
			return err
		}
		e.Status = models.PartStatusFree
		e.Size = 0
		e.Name = [16]byte{}
		return writeEBR(f, &e, ref.EBROff)
	}

	var m models.MBR
	if err := readMBR(f, &m); err != nil {
		return err
	}

	m.Partitions[ref.Index] = models.Partition{}
	return writeMBR(f, &m)
}

// DeletePartitionFull elimina y rellena con ceros
func (r *FileDiskRepository) DeletePartitionFull(path string, ref PartitionRef) error {
	if err := r.DeletePartitionFast(path, ref); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	return zeroRegion(f, ref.Start, ref.Size)
}

// GetExtendedBounds devuelve inicio y tamaño de la extendida
func (r *FileDiskRepository) GetExtendedBounds(path string) (int64, int64, bool, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		return 0, 0, false, err
	}
	defer f.Close()

	var m models.MBR
	if err := readMBR(f, &m); err != nil {
		return 0, 0, false, err
	}

	for i := 0; i < 4; i++ {
		p := m.Partitions[i]
		if p.Type == models.PartTypeExtend && p.Status == models.PartStatusUsed {
			return p.Start, p.Size, true, nil
		}
	}

	return 0, 0, false, nil
}

// GetPartitionUsedBytes devuelve bytes usados (placeholder)
func (r *FileDiskRepository) GetPartitionUsedBytes(path string, ref PartitionRef) (int64, error) {
	return -1, nil
}

// CreatePartition crea una partición con el tipo y fit indicados
func (r *FileDiskRepository) CreatePartition(path string, name string, sizeBytes int64, ptype PartType, fit byte) (string, error) {
	// Esta es una reimplementación simple, delegando a la lógica existente
	// Por ahora, retornar mensaje de stub
	return "CreatePartition no implementado en FileDiskRepository", nil
}
