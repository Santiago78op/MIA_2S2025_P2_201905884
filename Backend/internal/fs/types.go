package fs

type MkfsRequest struct {
	MountID string // id de partición montada
	FSKind  string // "2fs" | "3fs"
}

type MountRequest struct {
	DiskPath  string // ruta .mia
	Partition string // nombre
}

type FileStat struct {
	Size  int64
	Mode  uint16 // octal (ej. 0755)
	Owner string
	Group string
	IsDir bool
}

type TreeNode struct {
	Path     string
	IsDir    bool
	Mode     uint16
	Owner    string
	Group    string
	Children []TreeNode
}

type WriteFileRequest struct {
	Path    string
	Content []byte
	Append  bool
}

type MkdirRequest struct {
	Path string
	Deep bool // -p
}

type FindRequest struct {
	BasePath string
	Pattern  string
	Limit    int
}
