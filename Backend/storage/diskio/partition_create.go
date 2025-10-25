// Package: storage/diskio
// File   : partition_create.go
// Goal   : Crear particiones (P|E|L) con FIRST/BEST/WORST FIT sobre MBR/EBR

package diskio

import (
	"Backend/core/models"
	"errors"
	"fmt"
	"os"
	"sort"
)

// Constantes de ajuste
const (
	mbrReservedArea = int64(160) // Área reservada después del MBR
	ebrHeaderSize   = int64(42)  // Tamaño estimado del header EBR
)

// Códigos de Fit
const (
	FitFirst = 'F'
	FitBest  = 'B'
	FitWorst = 'W'
)

// ====== API PÚBLICA ======

// CreatePartition crea una partición primaria/extendida o lógica
func (r *FileFsRepository) CreatePartition(path string, name string, sizeBytes int64, ptype PartType, fit byte) (string, error) {
	if sizeBytes <= 0 {
		return "", errors.New("size debe ser > 0")
	}

	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Validar nombre único
	if _, err := r.FindPartition(path, name); err == nil {
		return "", fmt.Errorf("ya existe una partición llamada '%s'", name)
	}

	var m models.MBR
	if err := readMBR(f, &m); err != nil {
		return "", err
	}

	switch ptype {
	case TypePrimary, TypeExtended:
		return r.createPrimaryOrExtended(f, &m, name, sizeBytes, ptype, fit)
	case TypeLogical:
		return r.createLogical(f, &m, name, sizeBytes, fit)
	default:
		return "", errors.New("tipo inválido (P|E|L)")
	}
}

// ====== IMPLEMENTACIÓN: PRIMARIA / EXTENDIDA ======

type gap struct {
	start int64
	size  int64
}

func (r *FileFsRepository) createPrimaryOrExtended(f *os.File, m *models.MBR, name string, size int64, ptype PartType, fit byte) (string, error) {
	// Validar que hay slots disponibles
	slotIdx := -1
	for i := 0; i < 4; i++ {
		if m.Partitions[i].Status == models.PartStatusFree || m.Partitions[i].Size == 0 {
			slotIdx = i
			break
		}
	}
	if slotIdx == -1 {
		return "", errors.New("no hay slots disponibles en MBR")
	}

	// Si es extendida, validar que no exista otra
	if ptype == TypeExtended {
		for i := 0; i < 4; i++ {
			if m.Partitions[i].Type == models.PartTypeExtend && m.Partitions[i].Status == models.PartStatusUsed {
				return "", errors.New("ya existe una partición extendida")
			}
		}
	}

	// Calcular gaps disponibles
	gaps, err := r.collectPrimaryGaps(m)
	if err != nil {
		return "", err
	}

	// Seleccionar gap según fit
	g, found := selectGap(gaps, size, fit)
	if !found {
		return "", fmt.Errorf("no hay espacio contiguo para %d bytes", size)
	}

	// Crear partición
	partType := byte('P')
	if ptype == TypeExtended {
		partType = models.PartTypeExtend
	}

	var nameArr [16]byte
	copy(nameArr[:], []byte(name))

	m.Partitions[slotIdx] = models.Partition{
		Status: models.PartStatusUsed,
		Type:   partType,
		Fit:    fit,
		Start:  g.start,
		Size:   size,
		Name:   nameArr,
	}

	if err := writeMBR(f, m); err != nil {
		return "", err
	}

	// Si es extendida, inicializar primer EBR vacío
	if ptype == TypeExtended {
		emptyEBR := models.EBR{
			Status: models.PartStatusFree,
			Start:  g.start,
			Size:   0,
			Next:   -1,
			Name:   [16]byte{},
		}
		if err := writeEBR(f, &emptyEBR, g.start); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("Partición %s '%s' creada: %d bytes en offset %d", ptype.String(), name, size, g.start), nil
}

func (r *FileFsRepository) collectPrimaryGaps(m *models.MBR) ([]gap, error) {
	type seg struct {
		s int64
		e int64
	}

	// Segmentos usados
	used := []seg{{0, mbrReservedArea}} // MBR
	for i := 0; i < 4; i++ {
		p := m.Partitions[i]
		if p.Size <= 0 || p.Status == models.PartStatusFree {
			continue
		}
		used = append(used, seg{p.Start, p.Start + p.Size})
	}

	// Ordenar y fusionar
	sort.Slice(used, func(i, j int) bool { return used[i].s < used[j].s })
	merged := make([]seg, 0, len(used))
	for _, s := range used {
		if len(merged) == 0 {
			merged = append(merged, s)
			continue
		}
		last := &merged[len(merged)-1]
		if s.s <= last.e {
			if s.e > last.e {
				last.e = s.e
			}
		} else {
			merged = append(merged, s)
		}
	}

	// Calcular gaps
	var gaps []gap
	for i := 0; i < len(merged)-1; i++ {
		gapStart := merged[i].e
		gapEnd := merged[i+1].s
		gapSize := gapEnd - gapStart
		if gapSize > 0 {
			gaps = append(gaps, gap{start: gapStart, size: gapSize})
		}
	}

	// Gap final (desde última partición hasta el final del disco)
	if len(merged) > 0 {
		lastEnd := merged[len(merged)-1].e
		if m.Size > lastEnd {
			gaps = append(gaps, gap{start: lastEnd, size: m.Size - lastEnd})
		}
	} else {
		// No hay particiones, todo el disco está disponible
		gaps = append(gaps, gap{start: mbrReservedArea, size: m.Size - mbrReservedArea})
	}

	return gaps, nil
}

// ====== IMPLEMENTACIÓN: LÓGICA ======

func (r *FileFsRepository) createLogical(f *os.File, m *models.MBR, name string, size int64, fit byte) (string, error) {
	// Buscar partición extendida
	extIdx := -1
	for i := 0; i < 4; i++ {
		if m.Partitions[i].Type == models.PartTypeExtend && m.Partitions[i].Status == models.PartStatusUsed {
			extIdx = i
			break
		}
	}
	if extIdx == -1 {
		return "", errors.New("no existe partición extendida")
	}

	extStart := m.Partitions[extIdx].Start
	extSize := m.Partitions[extIdx].Size

	// Calcular gaps dentro de la extendida
	gaps, err := r.collectLogicalGaps(f, extStart, extSize)
	if err != nil {
		return "", err
	}

	// Cada lógica necesita espacio para EBR + datos
	requiredSize := ebrHeaderSize + size

	// Seleccionar gap
	g, found := selectGap(gaps, requiredSize, fit)
	if !found {
		return "", fmt.Errorf("no hay espacio en la extendida para %d bytes", requiredSize)
	}

	// Crear nuevo EBR
	var nameArr [16]byte
	copy(nameArr[:], []byte(name))

	newEBR := models.EBR{
		Status: models.PartStatusUsed,
		Fit:    fit,
		Start:  g.start + ebrHeaderSize, // Datos después del header EBR
		Size:   size,
		Next:   -1,
		Name:   nameArr,
	}

	// Encontrar último EBR para enlazar
	lastEBROff := r.findLastEBR(f, extStart)
	if lastEBROff == extStart {
		// Es el primer EBR lógico
		if err := writeEBR(f, &newEBR, g.start); err != nil {
			return "", err
		}
	} else {
		// Enlazar al final de la cadena
		var lastEBR models.EBR
		if err := readEBR(f, &lastEBR, lastEBROff); err != nil {
			return "", err
		}
		lastEBR.Next = g.start
		if err := writeEBR(f, &lastEBR, lastEBROff); err != nil {
			return "", err
		}
		if err := writeEBR(f, &newEBR, g.start); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("Partición Lógica '%s' creada: %d bytes en offset %d", name, size, g.start+ebrHeaderSize), nil
}

func (r *FileFsRepository) collectLogicalGaps(f *os.File, extStart, extSize int64) ([]gap, error) {
	type seg struct {
		s int64
		e int64
	}

	used := []seg{}

	// Recorrer cadena de EBRs
	ebrOff := extStart
	for ebrOff != -1 && ebrOff != 0 {
		var e models.EBR
		if err := readEBR(f, &e, ebrOff); err != nil {
			break
		}

		// Área del EBR + datos
		ebrEnd := ebrOff + ebrHeaderSize
		if e.Status == models.PartStatusUsed && e.Size > 0 {
			dataEnd := e.Start + e.Size
			if dataEnd > ebrEnd {
				ebrEnd = dataEnd
			}
		}
		used = append(used, seg{ebrOff, ebrEnd})

		ebrOff = e.Next
	}

	// Ordenar y fusionar
	sort.Slice(used, func(i, j int) bool { return used[i].s < used[j].s })
	merged := make([]seg, 0, len(used))
	for _, s := range used {
		if len(merged) == 0 {
			merged = append(merged, s)
			continue
		}
		last := &merged[len(merged)-1]
		if s.s <= last.e {
			if s.e > last.e {
				last.e = s.e
			}
		} else {
			merged = append(merged, s)
		}
	}

	// Calcular gaps
	var gaps []gap
	extEnd := extStart + extSize

	if len(merged) == 0 {
		// No hay EBRs, toda la extendida disponible
		gaps = append(gaps, gap{start: extStart, size: extSize})
	} else {
		// Gap inicial
		if merged[0].s > extStart {
			gaps = append(gaps, gap{start: extStart, size: merged[0].s - extStart})
		}

		// Gaps intermedios
		for i := 0; i < len(merged)-1; i++ {
			gapStart := merged[i].e
			gapEnd := merged[i+1].s
			gapSize := gapEnd - gapStart
			if gapSize > 0 {
				gaps = append(gaps, gap{start: gapStart, size: gapSize})
			}
		}

		// Gap final
		if merged[len(merged)-1].e < extEnd {
			gaps = append(gaps, gap{start: merged[len(merged)-1].e, size: extEnd - merged[len(merged)-1].e})
		}
	}

	return gaps, nil
}

func (r *FileFsRepository) findLastEBR(f *os.File, extStart int64) int64 {
	lastOff := extStart
	ebrOff := extStart

	for ebrOff != -1 && ebrOff != 0 {
		var e models.EBR
		if err := readEBR(f, &e, ebrOff); err != nil {
			break
		}
		lastOff = ebrOff
		ebrOff = e.Next
	}

	return lastOff
}

// ====== SELECCIÓN DE GAP ======

func selectGap(gaps []gap, requiredSize int64, fit byte) (gap, bool) {
	// Filtrar gaps que cumplen el tamaño
	valid := []gap{}
	for _, g := range gaps {
		if g.size >= requiredSize {
			valid = append(valid, g)
		}
	}

	if len(valid) == 0 {
		return gap{}, false
	}

	switch fit {
	case FitFirst:
		// Primer gap válido
		return valid[0], true

	case FitBest:
		// Gap más pequeño que cumple
		best := valid[0]
		for _, g := range valid[1:] {
			if g.size < best.size {
				best = g
			}
		}
		return best, true

	case FitWorst:
		// Gap más grande
		worst := valid[0]
		for _, g := range valid[1:] {
			if g.size > worst.size {
				worst = g
			}
		}
		return worst, true

	default:
		// Default: Worst Fit
		worst := valid[0]
		for _, g := range valid[1:] {
			if g.size > worst.size {
				worst = g
			}
		}
		return worst, true
	}
}
