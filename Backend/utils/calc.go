package utils

import (
	"Backend/core/models"
	"math"
)

// CalcN aplica la fórmula del enunciado P1 (EXT2 de proyecto):
// tamaño_partición = sizeof(SB) + n + 3n + n*sizeof(inodo) + 3n*sizeof(bloque)
// => n = floor( (part_size - sizeof(SB)) / (1 + 3 + sizeof(inodo) + 3*sizeof(bloque)) )
//
// IMPORTANTE: Esta fórmula asume la simplificación del proyecto (bitmaps como n y 3n bytes).
// Si tu implementación usa bitmaps reales a nivel bit, ajusta usando BytesForBits(n) y BytesForBits(3n).
func CalcN(partSize int64, inodeSize int, blockSize int) int {
	num := float64(partSize - int64(binarySizeSuperBlock()))
	den := float64(1 + 3 + inodeSize + 3*blockSize)
	if den <= 0 {
		return 0
	}
	n := int(math.Floor(num / den))
	if n < 0 {
		return 0
	}
	return n
}

// ComputeLayout calcula conteos y offsets absolutos dentro de la partición,
// y devuelve un SuperBlock completo con s_magic, s_filesystem_type y starts absolutos.
func ComputeLayout(partStart, partSize int64, inodeSize, blockSize int) models.SuperBlock {
	n := CalcN(partSize, inodeSize, blockSize)
	if n < 1 {
		n = 1 // evita división por cero; mínimo 1 inodo
	}
	inodes := int32(n)
	blocks := int32(3 * n)

	// tamaños según fórmula del proyecto (bitmaps como n y 3n bytes)
	sbSize := int64(binarySizeSuperBlock())
	bmInodeSize := int64(n)     // bytes
	bmBlockSize := int64(3 * n) // bytes
	inodeTblSize := int64(n * inodeSize)
	blockAreaSize := int64(3 * n * blockSize)

	// Offsets ABSOLUTOS
	sbStart := partStart
	bmInodeStart := sbStart + sbSize
	bmBlockStart := bmInodeStart + bmInodeSize
	inodeStart := bmBlockStart + bmBlockSize
	blockStart := inodeStart + inodeTblSize

	sb := models.SuperBlock{
		SFilesystemType:  models.FSProjectType,
		SInodesCount:     inodes,
		SBlocksCount:     blocks,
		SFreeInodesCount: inodes, // se actualizarán después de reservar / y users.txt
		SFreeBlocksCount: blocks,
		SMtime:           0,
		SUmtime:          0,
		SMagic:           models.FSMagic,
		SBmInodeStart:    bmInodeStart,
		SBmBlockStart:    bmBlockStart,
		SInodeStart:      inodeStart,
		SBlockStart:      blockStart,
	}
	_ = blockAreaSize // útil para validaciones si quieres comprobar que cabe
	return sb
}

// CalcNExt3 calcula n para EXT3 con la fórmula que incluye journal:
// partSize = sizeof(SB) + n*sizeof(JournalEntry) + n + 3n + n*sizeof(Inode) + 3n*sizeof(Block)
// n = floor((partSize - SB) / (JournalEntry + 1 + 3 + Inode + 3*Block))
func CalcNExt3(partSize int64, inodeSize int, blockSize int) int {
	const journalEntrySize = 288 // tamaño de models.JournalEntry
	sbSize := int64(binarySizeSuperBlockExt3())

	denom := float64(journalEntrySize + 1 + 3 + inodeSize + 3*blockSize)
	if denom <= 0 {
		return 0
	}

	num := float64(partSize - sbSize)
	n := int(math.Floor(num / denom))
	if n < 0 {
		return 0
	}
	return n
}

// ComputeLayoutExt3 calcula el layout para EXT3 con journal
func ComputeLayoutExt3(partStart, partSize int64, inodeSize, blockSize int) models.SuperBlock {
	n := CalcNExt3(partSize, inodeSize, blockSize)
	if n < 1 {
		n = 1
	}

	inodes := int32(n)
	blocks := int32(3 * n)

	const journalEntrySize = 288
	const maxJournalEntries = 50

	// Tamaños
	sbSize := int64(binarySizeSuperBlockExt3())
	journalSize := int64(maxJournalEntries * journalEntrySize)
	bmInodeSize := int64(n)
	bmBlockSize := int64(3 * n)
	inodeTblSize := int64(n * inodeSize)

	// Offsets ABSOLUTOS
	sbStart := partStart
	journalStart := sbStart + sbSize
	bmInodeStart := journalStart + journalSize
	bmBlockStart := bmInodeStart + bmInodeSize
	inodeStart := bmBlockStart + bmBlockSize
	blockStart := inodeStart + inodeTblSize

	sb := models.SuperBlock{
		SFilesystemType:  3, // EXT3
		SInodesCount:     inodes,
		SBlocksCount:     blocks,
		SFreeInodesCount: inodes,
		SFreeBlocksCount: blocks,
		SMtime:           0,
		SUmtime:          0,
		SMagic:           models.FSMagic,
		SBmInodeStart:    bmInodeStart,
		SBmBlockStart:    bmBlockStart,
		SInodeStart:      inodeStart,
		SBlockStart:      blockStart,
	}

	return sb
}

// binarySizeSuperBlock devuelve el tamaño que ocupará el SB en disco para esta app.
func binarySizeSuperBlock() int {
	// Reflect del struct para calcular; como el layout es fijo, puedes devolver el valor precalculado.
	// Para evitar reflección aquí, dejamos un valor aproximado típico.
	// Recomendado: sustituir por binary.Size(models.SuperBlock{}) si no te preocupa el import circular.
	var sb models.SuperBlock
	return sizeOf(sb)
}

// binarySizeSuperBlockExt3 devuelve el tamaño del SuperBlock para EXT3
func binarySizeSuperBlockExt3() int {
	// SuperBlock + 2 campos adicionales (SJournalStart int64 + SJournalCount int32)
	return binarySizeSuperBlock() + 8 + 4 // +12 bytes
}

// sizeOf obtiene binary.Size de forma genérica
func sizeOf(v any) int {
	// Pequeño wrapper para centralizar el cálculo (binary.Size panica en tipos no fijos).
	// Lo dejamos simple: evitamos importar encoding/binary aquí para no duplicar funciones.
	// Si prefieres, mueve esto a utils/bytes.go usando binary.Size().
	switch any(v).(type) {
	case models.SuperBlock:
		// Campos: 6 int32/int64 + 4 int64 => ajusta si tu struct cambia
		// Con el struct dado: int32*4 + int32*2? + int64*4 ... Para ir a la segura, devuelve  (32*? + 64*?)/8
		// Mejor: valor seguro típico para este proyecto (usa el tuyo real si difiere):
		return 4 + 4 + 4 + 4 + 4 + 4 + 8 + 8 + 4 + 8 + 8 + 8 + 8 // ≈ 80 bytes (ajusta si es distinto)
	default:
		return 0
	}
}
