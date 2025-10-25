// Package: storage/diskio
// File   : partition_types.go
// Goal   : Tipos auxiliares para operaciones de particiones

package diskio

// PartType identifica el tipo de partición
type PartType int

const (
	TypePrimary PartType = iota
	TypeExtended
	TypeLogical
)

// PartitionRef es una referencia a una partición (primaria, extendida o lógica)
type PartitionRef struct {
	Type  PartType // Primaria, Extendida o Lógica
	Index int      // índice en MBR (0-3) para P/E, o -1 para lógicas
	Start int64    // offset absoluto de inicio
	Size  int64    // tamaño en bytes
	EBROff int64   // offset del EBR (solo para lógicas)
}

// String retorna representación textual del tipo de partición
func (t PartType) String() string {
	switch t {
	case TypePrimary:
		return "Primaria"
	case TypeExtended:
		return "Extendida"
	case TypeLogical:
		return "Lógica"
	default:
		return "Desconocida"
	}
}
