package reports

import "fmt"

// ReportGenerator es el adaptador que usa Graphviz/lecturas para materializar el reporte.
type ReportGenerator interface {
	MBR(id, out string) (string, error)
	Disk(id, out string) (string, error)
	Inode(id, out string) (string, error)
	Block(id, out string) (string, error)
	BmInode(id, out string) (string, error)
	BmBlock(id, out string) (string, error)
	Tree(id, out string) (string, error)
	SuperBlock(id, out string) (string, error)
	File(id, out, filePath string) (string, error) // filePath = path_file
	LS(id, out, pathForLs string) (string, error)  // pathForLs = path_file_ls
}

type ReportService struct {
	gen ReportGenerator
}

func NewReportService(gen ReportGenerator) *ReportService {
	return &ReportService{gen: gen}
}

// Generate selecciona el generador correcto según name.
func (r *ReportService) Generate(name, id, out, extra string) (string, error) {
	switch name {
	case "mbr":
		return r.gen.MBR(id, out)
	case "disk":
		return r.gen.Disk(id, out)
	case "inode":
		return r.gen.Inode(id, out)
	case "block":
		return r.gen.Block(id, out)
	case "bm_inode":
		return r.gen.BmInode(id, out)
	case "bm_block":
		return r.gen.BmBlock(id, out)
	case "tree":
		return r.gen.Tree(id, out)
	case "sb", "superblock":
		return r.gen.SuperBlock(id, out)
	case "file":
		if extra == "" {
			return "", fmt.Errorf("file requiere extra=path_file")
		}
		return r.gen.File(id, out, extra)
	case "ls":
		if extra == "" {
			return "", fmt.Errorf("ls requiere extra=path_file_ls")
		}
		return r.gen.LS(id, out, extra)
	default:
		return "", fmt.Errorf("reporte no soportado: %s", name)
	}
}
