package models

// Constantes de tamaños (en bytes)
const (
	BlockSizeFile    = 64 // tamaño de bloque de archivo
	DirEntriesPerBlk = 4  // entradas por bloque de carpeta
	PtrPerBlk        = 16 // punteros por bloque de apuntadores
	InodePointers    = 15 // 12 directos + 3 indirectos
)

// Tipos y banderas de partición
const (
	PartStatusFree  byte = 0
	PartStatusUsed  byte = 1
	PartTypePrimary byte = 'P'
	PartTypeExtend  byte = 'E'
	// (Lógicas se representan con EBR)
	PartFitFF byte = 'F' // First Fit
	PartFitBF byte = 'B' // Best Fit
	PartFitWF byte = 'W' // Worst Fit
)

// EXT2 “del proyecto”
const (
	FSMagic         int32 = 0xEF53
	FSProjectType   int32 = 2 // marca EXT2 de proyecto
	FileTypeFolder  byte  = 0 // inodo carpeta
	FileTypeRegular byte  = 1 // inodo archivo
	PermDefaultFile       = "664"
	PermRootBypass        = true // root ignora permisos
)
