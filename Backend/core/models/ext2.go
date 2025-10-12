package models

// SuperBlock define offsets y contadores del FS
type SuperBlock struct {
	SFilesystemType  int32
	SInodesCount     int32
	SBlocksCount     int32
	SFreeInodesCount int32
	SFreeBlocksCount int32
	SMtime           int64 // último montaje (unix)
	SUmtime          int64 // último desmontaje (unix)
	SMagic           int32 // 0xEF53
	SBmInodeStart    int64 // offset absoluto bitmap inodos
	SBmBlockStart    int64 // offset absoluto bitmap bloques
	SInodeStart      int64 // offset absoluto tabla inodos
	SBlockStart      int64 // offset absoluto área de bloques
}

// Inode con 15 apuntadores (12 directos, 3 indirectos)
type Inode struct {
	IUid   int32
	IGid   int32
	ISize  int32
	IAtime int64
	ICtime int64
	IMtime int64
	IType  byte                 // 0 carpeta, 1 archivo
	IPerm  [3]byte              // UGO (ej: {'6','6','4'})
	IBlock [InodePointers]int32 // -1 si libre
}

// Directorios: 4 entradas por bloque
type DirEntry struct {
	BName  [12]byte // nombre base (sin '/')
	BInodo int32    // índice de inodo; -1 si libre
}
type FolderBlock struct {
	Content [DirEntriesPerBlk]DirEntry
}

// Archivos: 64 bytes por bloque
type FileBlock struct {
	Content [BlockSizeFile]byte
}

// Apuntadores: 16 enteros por bloque
type PointerBlock struct {
	Pointers [PtrPerBlk]int32
}
