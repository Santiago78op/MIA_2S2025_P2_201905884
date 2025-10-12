package models

// Partition: slot de partición primaria o extendida dentro del MBR.
type Partition struct {
	Status byte     // 0 libre, 1 usada
	Type   byte     // 'P' | 'E'
	Fit    byte     // 'F' | 'B' | 'W'
	Start  int64    // offset absoluto en bytes dentro del .mia
	Size   int64    // tamaño en bytes
	Name   [16]byte // nombre (terminado en 0 si sobra)
}

// MBR principal del disco .mia
type MBR struct {
	Size       int64 // tamaño total del disco .mia
	Timestamp  int64 // unix time
	Signature  int64 // firma única (para IDs de mount)
	Fit        byte  // fit del disco (FF/BF/WF)
	Partitions [4]Partition
}

// EBR para particiones lógicas (encadenadas) dentro de la extendida
type EBR struct {
	Status byte
	Fit    byte
	Start  int64 // offset absoluto del EBR
	Size   int64 // tamaño de la partición lógica
	Next   int64 // offset absoluto del siguiente EBR (o -1 si no hay)
	Name   [16]byte
}
