package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ViewerController maneja endpoints REST para el visualizador UI
type ViewerController struct {
}

// NewViewerController crea una nueva instancia del controller
func NewViewerController() *ViewerController {
	return &ViewerController{}
}

// ListDisks devuelve la lista de discos disponibles
// GET /api/disks
func (vc *ViewerController) ListDisks(ctx *gin.Context) {
	// TODO: Implementar GetAllMounts() en mounts
	ctx.JSON(http.StatusOK, gin.H{
		"disks": []interface{}{},
	})
}

// ListPartitions devuelve las particiones de un disco
// GET /api/disks/:disk/partitions
func (vc *ViewerController) ListPartitions(ctx *gin.Context) {
	diskPath := ctx.Param("disk")

	// TODO: Implementar GetAllMounts() en mounts
	ctx.JSON(http.StatusOK, gin.H{
		"disk":       diskPath,
		"partitions": []interface{}{},
	})
}

// GetTree devuelve el árbol de directorios de una partición
// GET /api/fs/:id/tree?path=/ruta
func (vc *ViewerController) GetTree(ctx *gin.Context) {
	mountID := ctx.Param("id")
	path := ctx.Query("path")

	if path == "" {
		path = "/"
	}

	// TODO: Verificar que la partición está montada
	// mount, err := vc.mountStore.GetMount(mountID)

	// TODO: Implementar lectura de directorio
	// Por ahora devolvemos ejemplo
	ctx.JSON(http.StatusOK, gin.H{
		"mount_id": mountID,
		"path":     path,
		"entries": []gin.H{
			{
				"name":  "users.txt",
				"type":  "file",
				"size":  128,
				"perm":  "664",
				"owner": "root",
				"group": "root",
			},
		},
	})
}

// GetFile devuelve el contenido de un archivo
// GET /api/fs/:id/file?path=/archivo.txt
func (vc *ViewerController) GetFile(ctx *gin.Context) {
	_ = ctx.Param("id")
	path := ctx.Query("path")

	if path == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "path parameter is required",
		})
		return
	}

	// TODO: Verificar que la partición está montada
	// TODO: Leer archivo usando Cat

	ctx.JSON(http.StatusOK, gin.H{
		"path":    path,
		"content": "Contenido del archivo (stub)",
	})
}

// GetJournal devuelve las entradas del journal
// GET /api/journal/:id
func (vc *ViewerController) GetJournal(ctx *gin.Context) {
	mountID := ctx.Param("id")

	// TODO: Verificar que la partición está montada
	// TODO: Leer journal

	ctx.JSON(http.StatusOK, gin.H{
		"mount_id": mountID,
		"entries":  []interface{}{},
	})
}

// Login maneja la autenticación de usuarios
// POST /api/auth/login
// Body: {"user": "usuario", "pass": "contraseña", "id": "mount_id"}
func (vc *ViewerController) Login(ctx *gin.Context) {
	var req struct {
		User string `json:"user" binding:"required"`
		Pass string `json:"pass" binding:"required"`
		ID   string `json:"id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// TODO: Verificar que la partición está montada
	// TODO: Verificar credenciales en users.txt
	// TODO: Crear sesión

	ctx.JSON(http.StatusOK, gin.H{
		"token":    req.ID, // Por simplicidad, usamos el mount ID como token
		"user":     req.User,
		"mount_id": req.ID,
	})
}

// Logout cierra la sesión de un usuario
// POST /api/auth/logout
// Body: {"id": "mount_id"}
func (vc *ViewerController) Logout(ctx *gin.Context) {
	var req struct {
		ID string `json:"id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// TODO: Cerrar sesión
	ctx.Status(http.StatusNoContent)
}
