package controllers

import (
	"net/http"
	"strings"

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
	svc ReportService
}

func NewReportsController(svc ReportService) *ReportsController {
	return &ReportsController{svc: svc}
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
