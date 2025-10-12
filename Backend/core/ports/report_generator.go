package ports

// Generación de reportes (usa Graphviz/lecturas del repo de FS/Disk)
type ReportGenerator interface {
	MBR(id, out string) (string, error)
	Disk(id, out string) (string, error)
	Inode(id, out string) (string, error)
	Block(id, out string) (string, error)
	BmInode(id, out string) (string, error)
	BmBlock(id, out string) (string, error)
	Tree(id, out string) (string, error)
	SuperBlock(id, out string) (string, error)
	File(id, out, filePath string) (string, error)
	LS(id, out, pathForLs string) (string, error)
}
