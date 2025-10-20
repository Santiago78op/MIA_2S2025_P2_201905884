package controllers

import (
	"net/http"
	"strings"
	"time"

	"Backend/core/ports"

	"github.com/gin-gonic/gin"
)

// ViewerController maneja endpoints REST para el visualizador UI
type ViewerController struct {
	fs     ports.FsRepository
	mounts ports.MountStore
	sess   ports.SessionStore
}

// NewViewerController crea una nueva instancia del controller
func NewViewerController(fs ports.FsRepository, mounts ports.MountStore, sess ports.SessionStore) *ViewerController {
	return &ViewerController{
		fs:     fs,
		mounts: mounts,
		sess:   sess,
	}
}

// ListDisks devuelve la lista de discos disponibles
// GET /api/disks
func (vc *ViewerController) ListDisks(ctx *gin.Context) {
	mounts := vc.mounts.List()

	// Agrupar por disco (path)
	disksMap := make(map[string][]gin.H)

	for _, m := range mounts {
		if _, exists := disksMap[m.Path]; !exists {
			disksMap[m.Path] = []gin.H{}
		}

		disksMap[m.Path] = append(disksMap[m.Path], gin.H{
			"id":   m.ID,
			"name": m.Name,
		})
	}

	// Convertir a formato de respuesta
	disks := []gin.H{}
	for path, partitions := range disksMap {
		disks = append(disks, gin.H{
			"path":       path,
			"partitions": partitions,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"disks": disks,
	})
}

// ListPartitions devuelve las particiones de un disco
// GET /api/disks/:disk/partitions
func (vc *ViewerController) ListPartitions(ctx *gin.Context) {
	diskPath := ctx.Param("disk")

	mounts := vc.mounts.List()
	partitions := []gin.H{}

	for _, m := range mounts {
		if m.Path == diskPath {
			partitions = append(partitions, gin.H{
				"id":      m.ID,
				"name":    m.Name,
				"mounted": true,
			})
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"disk":       diskPath,
		"partitions": partitions,
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

	// Validar que la partición está montada
	_, err := vc.mounts.GetMount(mountID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "partición no montada",
		})
		return
	}

	// Convertir path a slice
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) == 1 && pathParts[0] == "" {
		pathParts = []string{}
	}

	// Listar el directorio
	entries, err := vc.fs.ListDirectory(mountID, pathParts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Convertir a formato JSON
	jsonEntries := make([]gin.H, 0, len(entries))
	for _, e := range entries {
		jsonEntries = append(jsonEntries, gin.H{
			"name":  e.Name,
			"type":  e.Type,
			"size":  e.Size,
			"perm":  e.Perm,
			"owner": e.Owner,
			"group": e.Group,
			"mtime": e.Mtime,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"mount_id": mountID,
		"path":     path,
		"entries":  jsonEntries,
	})
}

// GetFile devuelve el contenido de un archivo
// GET /api/fs/:id/file?path=/archivo.txt
func (vc *ViewerController) GetFile(ctx *gin.Context) {
	mountID := ctx.Param("id")
	path := ctx.Query("path")

	if path == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "path parameter is required",
		})
		return
	}

	// Validar que la partición está montada
	_, err := vc.mounts.GetMount(mountID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "partición no montada",
		})
		return
	}

	// Convertir path a slice de rutas
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// Leer el archivo usando Cat
	content, err := vc.fs.Cat(mountID, [][]string{pathParts})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"path":    path,
		"content": content,
	})
}

// ====== Journal DTOs ======

type journalEntryDTO struct {
	Op        string    `json:"op"`
	Path      string    `json:"path"`
	Content   string    `json:"content,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type journalRowDTO struct {
	Operacion string `json:"operacion"`
	Path      string `json:"path"`
	Contenido string `json:"contenido"`
	Fecha     string `json:"fecha"`
}

// GetJournal devuelve las entradas del journal en formato crudo
// GET /api/journal/:id
func (vc *ViewerController) GetJournal(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "se requiere id"})
		return
	}

	// Validar que el id existe en el store de montajes
	found := false
	for _, m := range vc.mounts.List() {
		if m.ID == id {
			found = true
			break
		}
	}
	if !found {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "partición no montada"})
		return
	}

	// Leer entradas del journal
	journals, err := vc.fs.JournalList(id)
	if err != nil {
		// Si es EXT2 o no soportado, devolver error específico
		if strings.Contains(err.Error(), "EXT2") || strings.Contains(err.Error(), "no soportado") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "journal no disponible (solo EXT3)"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convertir a DTOs
	entries := make([]journalEntryDTO, 0, len(journals))
	for _, j := range journals {
		entries = append(entries, journalEntryDTO{
			Op:        strings.TrimSpace(string(j.Content.Op[:])),
			Path:      strings.TrimSpace(string(j.Content.Path[:])),
			Content:   strings.TrimSpace(string(j.Content.Content[:])),
			Timestamp: time.Unix(int64(j.Content.Date), 0),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"mount_id": id,
		"entries":  entries,
	})
}

// GetJournalTable devuelve las entradas del journal en formato tabla
// GET /api/journal/:id/table
// Formato compatible con UI: { operacion, path, contenido, fecha }
func (vc *ViewerController) GetJournalTable(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "se requiere id"})
		return
	}

	// Validar que el id existe en el store de montajes
	found := false
	for _, m := range vc.mounts.List() {
		if m.ID == id {
			found = true
			break
		}
	}
	if !found {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "partición no montada"})
		return
	}

	// Leer entradas del journal
	journals, err := vc.fs.JournalList(id)
	if err != nil {
		// Si es EXT2 o no soportado, devolver error específico
		if strings.Contains(err.Error(), "EXT2") || strings.Contains(err.Error(), "no soportado") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "journal no disponible (solo EXT3)"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convertir a formato tabla
	rows := make([]journalRowDTO, 0, len(journals))
	for _, j := range journals {
		op := strings.TrimSpace(string(j.Content.Op[:]))
		path := strings.TrimSpace(string(j.Content.Path[:]))
		content := strings.TrimSpace(string(j.Content.Content[:]))
		timestamp := time.Unix(int64(j.Content.Date), 0)

		rows = append(rows, journalRowDTO{
			Operacion: op,
			Path:      path,
			Contenido: content,
			Fecha:     timestamp.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"mount_id": id,
		"rows":     rows,
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
