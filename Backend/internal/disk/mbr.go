package disk

import "time"

// Constantes de tipo/estado/fit (usa minúsculas para estricto binario)
const (
	PartTypePrimary  byte = 'P'
	PartTypeExtended byte = 'E'
	PartTypeLogical  byte = 'L'

	PartStatusFree byte = 0
	PartStatusUsed byte = 1

	FitFF byte = 'F' // first
	FitBF byte = 'B' // best
	FitWF byte = 'W' // worst
)

// Tamaños fijos y límites
const (
	MaxPrimaries = 4
	NameLen      = 16 // nombre de partición fijo
)

// MBR fijo (sin timestamps variables en binario; lo derivamos en runtime si quieres)
type MBR struct {
	SizeBytes int64
	CreatedAt int64 // unix seconds (para binary.Size fijo)
	DiskSig   int32
	Fit       byte
	_         [7]byte // padding para alinear a 8 bytes
	Parts     [MaxPrimaries]Partition
}

// Partition (primaria o extendida) — tamaño fijo en disco.
type Partition struct {
	Status byte  // 0 libre, 1 usada
	Type   byte  // 'P'|'E'
	Fit    byte  // 'F'|'B'|'W'
	Start  int64 // inicio en bytes
	Size   int64 // tamaño en bytes
	Name   [NameLen]byte
	_      [8]byte // padding
}

// EBR (para lógicas) — en lista enlazada dentro del espacio de la extendida.
type EBR struct {
	Status byte
	Fit    byte
	Start  int64 // inicio del EBR (posición propia)
	Size   int64 // tamaño del bloque de datos lógico
	Next   int64 // offset del siguiente EBR o -1
	Name   [NameLen]byte
	_      [8]byte
}

// Helpers runtime (no forman parte del binario en disco)
func NewMBR(total int64, fit byte, sig int32) MBR {
	return MBR{
		SizeBytes: total,
		CreatedAt: time.Now().Unix(),
		DiskSig:   sig,
		Fit:       fit,
	}
}
