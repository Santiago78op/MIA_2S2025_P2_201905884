package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ReportService genera un reporte y devuelve, idealmente, la ruta final del archivo creado.
type ReportService interface {
	// name: mbr, disk, inode, block, bm_inode, bm_block, tree, sb, file, ls
	// id:   id de mount (ej. "841A")
	// out:  ruta absoluta/relativa donde guardar (el servicio debe crear carpetas si no existen)
	// extra: parámetro adicional (p.ej. path_file_ls para 'ls' o 'file')
	Generate(name, id, out, extra string) (string, error)
}

type ReportsController struct {
	svc         ReportService
	reportsPath string
}

func NewReportsController(svc ReportService, reportsPath string) *ReportsController {
	return &ReportsController{svc: svc, reportsPath: reportsPath}
}

type reportReq struct {
	Name  string `json:"name"  binding:"required"`
	ID    string `json:"id"    binding:"required"`
	Out   string `json:"out"   binding:"required"`
	Extra string `json:"extra"` // opcional: path_file_ls, path_file, etc.
}

type reportRes struct {
	Path string `json:"path"` // ruta del archivo generado (para que el front lo consuma)
	Note string `json:"note"` // msg opcional
}

func (r *ReportsController) Generate(ctx *gin.Context) {
	var req reportReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}
	name := strings.TrimSpace(strings.ToLower(req.Name))
	if name == "" || req.ID == "" || strings.TrimSpace(req.Out) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "parámetros inválidos"})
		return
	}

	path, err := r.svc.Generate(name, req.ID, req.Out, req.Extra)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, reportRes{
		Path: path,
		Note: "reporte generado",
	})
}

// FileInfo representa un archivo en la carpeta Reports
type FileInfo struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
	IsImage  bool      `json:"is_image"`
	Ext      string    `json:"ext"`
	URL      string    `json:"url"`
}

// List devuelve todos los archivos en la carpeta Reports
func (r *ReportsController) List(ctx *gin.Context) {
	if r.reportsPath == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "reportsPath no configurado"})
		return
	}

	entries, err := os.ReadDir(r.reportsPath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "error leyendo carpeta Reports: " + err.Error()})
		return
	}

	var files []FileInfo
	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".svg": true, ".webp": true}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		isImage := imageExts[ext]

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsImage: isImage,
			Ext:     ext,
			URL:     "/reports/static/" + entry.Name(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"files": files,
		"count": len(files),
	})
}
