// Package: storage/diskio
// File   : partition_ops.go
// Goal   : Operaciones avanzadas sobre particiones MBR/EBR para P2
//          (FindPartition, ResizePartition, DeletePartition, etc.)

package diskio

import (
	"Backend/core/models"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ========================= BÚSQUEDA DE PARTICIONES =========================

// FindPartition busca una partición por nombre en el MBR/EBR
func (r *FileFsRepository) FindPartition(path string, name string) (PartitionRef, error) {
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

	// Buscar en lógicas (dentro de la extendida)
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

	return PartitionRef{}, fmt.Errorf("partición '%s' no encontrada", name)
}

// ========================= SIGUIENTE PARTICIÓN =========================

// NextPartitionAfter devuelve la partición inmediata después de ref (por offset)
func (r *FileFsRepository) NextPartitionAfter(path string, ref PartitionRef) (PartitionRef, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		return PartitionRef{}, err
	}
	defer f.Close()

	var m models.MBR
	if err := readMBR(f, &m); err != nil {
		return PartitionRef{}, err
	}

	// Recopilar todas las particiones con sus offsets
	type part struct {
		ref   PartitionRef
		start int64
	}
	var parts []part

	// Primarias y extendida
	for i := 0; i < 4; i++ {
		p := m.Partitions[i]
		if p.Status == models.PartStatusUsed && p.Size > 0 {
			ptype := TypePrimary
			if p.Type == models.PartTypeExtend {
				ptype = TypeExtended
			}
			parts = append(parts, part{
				ref: PartitionRef{
					Type:  ptype,
					Index: i,
					Start: p.Start,
					Size:  p.Size,
				},
				start: p.Start,
			})
		}
	}

	// Lógicas
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
			if e.Status == models.PartStatusUsed && e.Size > 0 {
				parts = append(parts, part{
					ref: PartitionRef{
						Type:   TypeLogical,
						Index:  -1,
						Start:  e.Start,
						Size:   e.Size,
						EBROff: ebrOff,
					},
					start: e.Start,
				})
			}
			ebrOff = e.Next
		}
	}

	// Buscar la siguiente partición después de ref
	minNext := int64(-1)
	var nextRef PartitionRef
	for _, p := range parts {
		if p.start > ref.Start {
			if minNext == -1 || p.start < minNext {
				minNext = p.start
				nextRef = p.ref
			}
		}
	}

	if minNext == -1 {
		return PartitionRef{}, errors.New("no hay partición siguiente")
	}

	return nextRef, nil
}

// ========================= LECTURA DE PARTICIÓN =========================

// ReadPartition devuelve el start y size de una partición
func (r *FileFsRepository) ReadPartition(path string, ref PartitionRef) (int64, int64, error) {
	return ref.Start, ref.Size, nil
}

// ========================= RESIZE DE PARTICIÓN =========================

// ResizePartition actualiza el tamaño de una partición
func (r *FileFsRepository) ResizePartition(path string, ref PartitionRef, newSize int64) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if ref.Type == TypeLogical {
		// Actualizar EBR
		var e models.EBR
		if err := readEBR(f, &e, ref.EBROff); err != nil {
			return err
		}
		e.Size = newSize
		return writeEBR(f, &e, ref.EBROff)
	}

	// Actualizar MBR (primaria o extendida)
	var m models.MBR
	if err := readMBR(f, &m); err != nil {
		return err
	}

	if ref.Index < 0 || ref.Index >= 4 {
		return errors.New("índice de partición inválido")
	}

	m.Partitions[ref.Index].Size = newSize
	return writeMBR(f, &m)
}

// ========================= DELETE DE PARTICIÓN =========================

// DeletePartitionFast limpia la entrada de la partición (fast)
func (r *FileFsRepository) DeletePartitionFast(path string, ref PartitionRef) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if ref.Type == TypeLogical {
		// Limpiar EBR y reenlazar cadena
		var e models.EBR
		if err := readEBR(f, &e, ref.EBROff); err != nil {
			return err
		}

		// Encontrar EBR anterior para reenlazar
		var m models.MBR
		if err := readMBR(f, &m); err != nil {
			return err
		}

		extIdx := -1
		for i := 0; i < 4; i++ {
			if m.Partitions[i].Type == models.PartTypeExtend && m.Partitions[i].Status == models.PartStatusUsed {
				extIdx = i
				break
			}
		}

		if extIdx < 0 {
			return errors.New("no se encontró partición extendida")
		}

		extStart := m.Partitions[extIdx].Start
		prevOff := int64(-1)
		ebrOff := extStart

		for ebrOff != -1 && ebrOff != 0 {
			if ebrOff == ref.EBROff {
				// Encontrado, reenlazar
				if prevOff == -1 {
					// Es el primero, no hay anterior
					// Solo limpiamos este EBR
				} else {
					// Reenlazar el anterior al siguiente
					var prevEBR models.EBR
					if err := readEBR(f, &prevEBR, prevOff); err != nil {
						return err
					}
					prevEBR.Next = e.Next
					if err := writeEBR(f, &prevEBR, prevOff); err != nil {
						return err
					}
				}
				break
			}
			prevOff = ebrOff
			var tempEBR models.EBR
			if err := readEBR(f, &tempEBR, ebrOff); err != nil {
				return err
			}
			ebrOff = tempEBR.Next
		}

		// Limpiar el EBR actual
		e.Status = models.PartStatusFree
		e.Size = 0
		e.Name = [16]byte{}
		return writeEBR(f, &e, ref.EBROff)
	}

	// Limpiar entrada MBR (primaria o extendida)
	var m models.MBR
	if err := readMBR(f, &m); err != nil {
		return err
	}

	if ref.Index < 0 || ref.Index >= 4 {
		return errors.New("índice de partición inválido")
	}

	// Si es extendida, limpiar toda la cadena de EBRs
	if ref.Type == TypeExtended {
		ebrOff := ref.Start
		for ebrOff != -1 && ebrOff != 0 {
			var e models.EBR
			if err := readEBR(f, &e, ebrOff); err != nil {
				break
			}
			nextOff := e.Next
			e.Status = models.PartStatusFree
			e.Size = 0
			e.Name = [16]byte{}
			_ = writeEBR(f, &e, ebrOff)
			ebrOff = nextOff
		}
	}

	// Limpiar entrada MBR
	m.Partitions[ref.Index] = models.Partition{}
	return writeMBR(f, &m)
}

// DeletePartitionFull elimina y rellena con ceros la región
func (r *FileFsRepository) DeletePartitionFull(path string, ref PartitionRef) error {
	if err := r.DeletePartitionFast(path, ref); err != nil {
		return err
	}

	// Rellenar con ceros
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	return zeroRegion(f, ref.Start, ref.Size)
}

// ========================= INFORMACIÓN EXTENDIDA =========================

// GetExtendedBounds devuelve inicio y tamaño de la partición extendida
func (r *FileFsRepository) GetExtendedBounds(path string) (int64, int64, bool, error) {
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

// GetPartitionUsedBytes devuelve bytes usados (placeholder, retorna -1)
func (r *FileFsRepository) GetPartitionUsedBytes(path string, ref PartitionRef) (int64, error) {
	// TODO: leer superbloque si la partición está formateada
	return -1, nil
}

// ========================= HELPERS =========================

func trimPartName(b []byte) string {
	return strings.Trim(string(b), "\x00 ")
}
