package ports

import "backend/core/models"

// Región (inicio y tamaño absolutos) de la partición formateada
type Region struct {
	Start int64
	Size  int64
}

type FsRepository interface {
	// ---- MKFS ----
	// Crea EXT2 en la partición (calcula n, offsets, inicializa SB/bitmaps/tablas)
	Mkfs(id string) error

	// ---- CRUD de directorios/archivos ----
	Mkdir(id string, absPath []string, parents bool) error
	Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool) error
	Cat(id string, files [][]string) (string, error)

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
}
