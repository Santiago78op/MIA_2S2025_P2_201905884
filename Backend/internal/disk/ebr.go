package disk

import (
	"fmt"
	"os"
)

// CreateEBR crea un nuevo EBR al inicio de una partición extendida
func CreateEBR(f *os.File, extStart int64, fit byte) error {
	ebr := EBR{
		Status: 0, // Vacío inicialmente
		Fit:    fit,
		Start:  extStart,
		Size:   0,
		Next:   -1, // No hay siguiente
	}
	return writeStruct(f, extStart, &ebr)
}

// ListEBRs lee todos los EBRs en la cadena dentro de una partición extendida
func ListEBRs(f *os.File, extStart, extEnd int64) ([]EBR, error) {
	var ebrs []EBR
	currentOffset := extStart

	for currentOffset != -1 && currentOffset < extEnd {
		var ebr EBR
		if err := readStruct(f, currentOffset, &ebr); err != nil {
			return nil, fmt.Errorf("error al leer EBR en offset %d: %v", currentOffset, err)
		}

		// Agregar EBR a la lista (incluso si está vacío, para análisis)
		ebrs = append(ebrs, ebr)

		// Si el EBR está vacío (primer EBR sin usar), terminamos
		if ebr.Status == 0 && ebr.Size == 0 {
			break
		}

		// Siguiente EBR
		currentOffset = ebr.Next
	}

	return ebrs, nil
}

// FindEBRByName busca un EBR por nombre de partición
func FindEBRByName(f *os.File, extStart, extEnd int64, partName string) (*EBR, int64, error) {
	currentOffset := extStart

	for currentOffset != -1 && currentOffset < extEnd {
		var ebr EBR
		if err := readStruct(f, currentOffset, &ebr); err != nil {
			return nil, 0, fmt.Errorf("error al leer EBR: %v", err)
		}

		// Comparar nombre
		if ebr.Status == PartStatusUsed && trimEBRName(ebr.Name) == partName {
			return &ebr, currentOffset, nil
		}

		currentOffset = ebr.Next
	}

	return nil, 0, fmt.Errorf("partición lógica %s no encontrada", partName)
}

// AddLogicalPartition agrega una nueva partición lógica en la extendida
func AddLogicalPartition(f *os.File, extStart, extEnd int64, partName string, sizeBytes int64, fit byte) error {
	// 1. Listar EBRs existentes
	ebrs, err := ListEBRs(f, extStart, extEnd)
	if err != nil {
		return fmt.Errorf("error al listar EBRs: %v", err)
	}

	// 2. Calcular espacios libres dentro de la extendida
	freeSpaces := calculateFreeSpacesInExtended(ebrs, extStart, extEnd)
	if len(freeSpaces) == 0 {
		return ErrNoSpace
	}

	// 3. Elegir espacio según el algoritmo de fit
	var chosenSpace *FreeSpaceEBR
	switch fit {
	case FitFF: // First Fit
		for i := range freeSpaces {
			if freeSpaces[i].Size >= sizeBytes {
				chosenSpace = &freeSpaces[i]
				break
			}
		}
	case FitBF: // Best Fit
		var bestIdx = -1
		var bestSize int64 = extEnd - extStart + 1
		for i := range freeSpaces {
			if freeSpaces[i].Size >= sizeBytes && freeSpaces[i].Size < bestSize {
				bestSize = freeSpaces[i].Size
				bestIdx = i
			}
		}
		if bestIdx != -1 {
			chosenSpace = &freeSpaces[bestIdx]
		}
	case FitWF: // Worst Fit
		var worstIdx = -1
		var worstSize int64 = -1
		for i := range freeSpaces {
			if freeSpaces[i].Size >= sizeBytes && freeSpaces[i].Size > worstSize {
				worstSize = freeSpaces[i].Size
				worstIdx = i
			}
		}
		if worstIdx != -1 {
			chosenSpace = &freeSpaces[worstIdx]
		}
	}

	if chosenSpace == nil {
		return ErrNoSpace
	}

	// 4. Crear nuevo EBR
	name, err := fmtName(partName)
	if err != nil {
		return err
	}

	newEBR := EBR{
		Status: PartStatusUsed,
		Fit:    fit,
		Start:  chosenSpace.Start,
		Size:   sizeBytes,
		Next:   -1,
		Name:   name,
	}

	// 5. Escribir el nuevo EBR
	if err := writeStruct(f, chosenSpace.Start, &newEBR); err != nil {
		return fmt.Errorf("error al escribir EBR: %v", err)
	}

	// 6. Actualizar el EBR anterior para que apunte al nuevo
	if chosenSpace.PrevOffset != -1 {
		var prevEBR EBR
		if err := readStruct(f, chosenSpace.PrevOffset, &prevEBR); err != nil {
			return fmt.Errorf("error al leer EBR previo: %v", err)
		}
		prevEBR.Next = chosenSpace.Start
		if err := writeStruct(f, chosenSpace.PrevOffset, &prevEBR); err != nil {
			return fmt.Errorf("error al actualizar EBR previo: %v", err)
		}
	}

	return nil
}

// DeleteLogicalPartition elimina una partición lógica
func DeleteLogicalPartition(f *os.File, extStart, extEnd int64, partName string, fullDelete bool) error {
	// Buscar el EBR
	ebr, ebrOffset, err := FindEBRByName(f, extStart, extEnd, partName)
	if err != nil {
		return err
	}

	// Si fullDelete, limpiar el área con ceros
	if fullDelete {
		if err := zeroRange(f, ebr.Start, ebr.Size); err != nil {
			return fmt.Errorf("error al limpiar partición: %v", err)
		}
	}

	// Marcar el EBR como libre
	ebr.Status = PartStatusFree
	ebr.Size = 0
	ebr.Name = [NameLen]byte{}

	// Escribir el EBR actualizado
	return writeStruct(f, ebrOffset, ebr)
}

// FreeSpaceEBR representa un espacio libre en la partición extendida
type FreeSpaceEBR struct {
	Start      int64 // Inicio del espacio libre
	Size       int64 // Tamaño del espacio
	PrevOffset int64 // Offset del EBR previo (-1 si es el primero)
}

// calculateFreeSpacesInExtended calcula los espacios libres dentro de una extendida
func calculateFreeSpacesInExtended(ebrs []EBR, extStart, extEnd int64) []FreeSpaceEBR {
	var freeSpaces []FreeSpaceEBR
	const ebrSize = 64 // Tamaño aproximado del EBR

	// Si no hay EBRs, todo el espacio está libre
	if len(ebrs) == 0 {
		freeSpaces = append(freeSpaces, FreeSpaceEBR{
			Start:      extStart,
			Size:       extEnd - extStart,
			PrevOffset: -1,
		})
		return freeSpaces
	}

	// Ordenar EBRs por posición de inicio
	sortedEBRs := make([]struct {
		ebr    EBR
		offset int64
	}, 0)

	currentOffset := extStart
	for i := 0; i < len(ebrs) && currentOffset != -1; i++ {
		sortedEBRs = append(sortedEBRs, struct {
			ebr    EBR
			offset int64
		}{ebrs[i], currentOffset})
		currentOffset = ebrs[i].Next
	}

	// Buscar espacios libres entre EBRs
	for i := 0; i < len(sortedEBRs); i++ {
		ebrStart := sortedEBRs[i].offset
		ebrEnd := ebrStart + ebrSize + sortedEBRs[i].ebr.Size

		var nextStart int64
		if i+1 < len(sortedEBRs) {
			nextStart = sortedEBRs[i+1].offset
		} else {
			nextStart = extEnd
		}

		// Si hay espacio entre este EBR y el siguiente
		if ebrEnd < nextStart {
			gap := nextStart - ebrEnd
			if gap > ebrSize { // Debe caber al menos un EBR
				freeSpaces = append(freeSpaces, FreeSpaceEBR{
					Start:      ebrEnd,
					Size:       gap,
					PrevOffset: ebrStart,
				})
			}
		}
	}

	// Verificar si hay espacio al inicio (antes del primer EBR)
	if len(sortedEBRs) > 0 && sortedEBRs[0].offset > extStart+ebrSize {
		gap := sortedEBRs[0].offset - extStart
		freeSpaces = append([]FreeSpaceEBR{{
			Start:      extStart,
			Size:       gap,
			PrevOffset: -1,
		}}, freeSpaces...)
	}

	return freeSpaces
}

// trimEBRName convierte [16]byte a string limpio
func trimEBRName(n [NameLen]byte) string {
	for i, b := range n {
		if b == 0 {
			return string(n[:i])
		}
	}
	return string(n[:])
}
