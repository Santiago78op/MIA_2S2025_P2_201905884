package ports

import "Backend/core/models"

// Región (inicio y tamaño absolutos) de la partición formateada
type Region struct {
	Start int64
	Size  int64
}

type FsRepository interface {
	// ---- MKFS ----
	// Crea EXT2 o EXT3 en la partición (calcula n, offsets, inicializa SB/bitmaps/tablas/journal)
	// fsType: 2 para EXT2, 3 para EXT3
	// full: true para formateo completo (único soportado actualmente)
	Mkfs(id string) error
	MkfsWithType(id string, fsType int32, full bool) error

	// ---- CRUD de directorios/archivos ----
	Mkdir(id string, absPath []string, parents bool) error
	Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool) error
	Cat(id string, files [][]string) (string, error)

	// ---- Nuevas operaciones P2 ----
	Remove(id string, path []string) error
	Edit(id string, path []string, contentHostPath string) error
	Rename(id string, path []string, newName string) error
	Copy(id string, srcPath []string, destPath []string) error
	Move(id string, srcPath []string, destPath []string) error
	Find(id string, basePath []string, pattern string) ([]string, error)

	// ---- Permisos y propietarios ----
	Chmod(id string, path []string, ugo [3]byte, recursive bool) error
	Chown(id string, path []string, user string, recursive bool) error

	// ---- Lecturas/escrituras de estructuras (si tu caso de uso lo requiere directo) ----
	ReadSuper(id string) (models.SuperBlock, Region, error)
	WriteSuper(id string, sb models.SuperBlock) error

	ReadBitmapInode(id string) ([]byte, error)
	ReadBitmapBlock(id string) ([]byte, error)
	WriteBitmapInode(id string, data []byte) error
	WriteBitmapBlock(id string, data []byte) error

	ReadInode(id string, index int) (models.Inode, error)
	WriteInode(id string, index int, ino models.Inode) error

	ReadBlock(id string, index int) ([]byte, error)
	WriteBlock(id string, index int, data []byte) error

	// ---- Journal (solo para EXT3) ----
	JournalAppend(id string, entry models.Journal) error
	JournalList(id string) ([]models.Journal, error)
	JournalClear(id string) error

	// ---- Recovery y Loss ----
	WipeDataAreas(id string) error // Loss: limpia bitmaps, inodos y bloques
	Recovery(id string) error       // Recovery: restaura desde journal
}
