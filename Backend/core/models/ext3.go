package models

// SuperBlockExt3 representa el superbloque para EXT3
// Tamaño fijo: 112 bytes (11×int32[44] + 9×int64[72])
type SuperBlockExt3 struct {
	// Información básica del filesystem
	FsType       int32 // 2=EXT2, 3=EXT3
	InodesCount  int32 // Cantidad total de inodos (n)
	BlocksCount  int32 // Cantidad total de bloques (3n)
	FreeBlocks   int32 // Bloques libres
	FreeInodes   int32 // Inodos libres

	// Timestamps
	MTime  int64 // Último montaje (unix timestamp)
	UMTime int64 // Último desmontaje (unix timestamp)

	// Metadatos
	MntCount  int32 // Contador de montajes
	Magic     int32 // 0xEF53
	InodeSize int32 // Tamaño de cada inodo (100 bytes típicamente)
	BlockSize int32 // Tamaño de cada bloque (64 bytes)

	// Offsets absolutos (en bytes desde el inicio del disco)
	FirstInode   int64 // Primer inodo disponible
	FirstBlock   int64 // Primer bloque disponible
	BmInodeStart int64 // Inicio del bitmap de inodos
	BmBlockStart int64 // Inicio del bitmap de bloques
	InodeStart   int64 // Inicio de la tabla de inodos
	BlockStart   int64 // Inicio del área de bloques

	// Journal (solo EXT3)
	JournalStart int64 // Offset absoluto del inicio del journal
	JournalCount int32 // Cantidad de entradas de journal (50 por defecto)

	// Contexto (no se serializa)
	PartStart int64 `binary:"-"` // Inicio de la partición (para cálculos)
}

// Constantes para EXT3
const (
	JournalMaxEntries = 50
	Ext3FsType        = 3
	Ext2FsType        = 2
	// FSMagic ya está definido en types.go
)

// IsExt3 verifica si el tipo de filesystem es EXT3
func IsExt3(fsType int32) bool {
	return fsType == Ext3FsType
}
