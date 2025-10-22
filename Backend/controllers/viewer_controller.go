package controllers

import (
	"encoding/binary"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"Backend/core/models"
	"Backend/core/ports"

	"github.com/gin-gonic/gin"
)

// ViewerController maneja endpoints REST para el visualizador UI
type ViewerController struct {
	fs        ports.FsRepository
	mounts    ports.MountStore
	sess      ports.SessionStore
	disksPath string
}

// NewViewerController crea una nueva instancia del controller
func NewViewerController(fs ports.FsRepository, mounts ports.MountStore, sess ports.SessionStore, disksPath string) *ViewerController {
	return &ViewerController{
		fs:        fs,
		mounts:    mounts,
		sess:      sess,
		disksPath: disksPath,
	}
}

// ListDisks devuelve la lista de discos disponibles
// GET /api/disks
func (vc *ViewerController) ListDisks(ctx *gin.Context) {
	// Leer todos los archivos .mia de la carpeta Discos
	files, err := filepath.Glob(filepath.Join(vc.disksPath, "*.mia"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error leyendo carpeta de discos: " + err.Error(),
		})
		return
	}

	// Obtener lista de montajes para saber qué particiones están montadas
	mounts := vc.mounts.List()
	mountsByDisk := make(map[string][]mountedPartition)
	for _, m := range mounts {
		mountsByDisk[m.Path] = append(mountsByDisk[m.Path], mountedPartition{
			ID:   m.ID,
			Name: m.Name,
		})
	}

	// Construir respuesta
	disks := []gin.H{}
	for _, filePath := range files {
		// Obtener info del archivo
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		// Extraer nombre del archivo
		fileName := filepath.Base(filePath)

		// Obtener particiones montadas de este disco (si las hay)
		mounted := mountsByDisk[filePath]
		if mounted == nil {
			mounted = []mountedPartition{}
		}

		// Leer Fit del MBR
		fit := readMBRFit(filePath)

		disks = append(disks, gin.H{
			"path":    filePath,
			"name":    fileName,
			"size":    formatBytes(fileInfo.Size()),
			"fit":     fit,
			"mounted": mounted,
		})
	}

	ctx.JSON(http.StatusOK, disks)
}

// readMBRFit lee el campo Fit del MBR de un disco
func readMBRFit(diskPath string) string {
	file, err := os.Open(diskPath)
	if err != nil {
		return "N/A"
	}
	defer file.Close()

	var mbr models.MBR
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		return "N/A"
	}

	// Convertir byte a string
	switch mbr.Fit {
	case 'F', 'f':
		return "FF"
	case 'B', 'b':
		return "BF"
	case 'W', 'w':
		return "WF"
	default:
		return "N/A"
	}
}

// logicalPartitionInfo representa la información de una partición lógica
type logicalPartitionInfo struct {
	Name string
	Size int64
	Fit  string
}

// readLogicalPartitions lee las particiones lógicas de una partición extendida
func readLogicalPartitions(file *os.File, extStart int64) []logicalPartitionInfo {
	var logicals []logicalPartitionInfo

	// Empezar desde el inicio de la partición extendida
	currentEBRPos := extStart

	for currentEBRPos != -1 && currentEBRPos != 0 {
		// Posicionarse en el EBR actual
		_, err := file.Seek(currentEBRPos, 0)
		if err != nil {
			break
		}

		// Leer el EBR
		var ebr models.EBR
		err = binary.Read(file, binary.LittleEndian, &ebr)
		if err != nil {
			break
		}

		// Si el EBR está vacío (primer EBR sin usar), terminar
		if ebr.Status == 0 && ebr.Size == 0 {
			break
		}

		// Si el EBR está en uso, agregar a la lista
		if ebr.Status == 1 || ebr.Status == '1' {
			name := strings.TrimRight(string(ebr.Name[:]), "\x00")

			// Convertir fit
			fit := "N/A"
			switch ebr.Fit {
			case 'F', 'f':
				fit = "FF"
			case 'B', 'b':
				fit = "BF"
			case 'W', 'w':
				fit = "WF"
			}

			logicals = append(logicals, logicalPartitionInfo{
				Name: name,
				Size: ebr.Size,
				Fit:  fit,
			})
		}

		// Ir al siguiente EBR
		currentEBRPos = ebr.Next
	}

	return logicals
}

// formatBytes convierte bytes a formato legible
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + " " + "KMGTPE"[exp:exp+1] + "B"
}

type diskInfo struct {
	Path    string
	Mounted []mountedPartition
}

type mountedPartition struct {
	ID   string
	Name string
}

// ListPartitions devuelve las particiones de un disco
// GET /api/disks/partitions?path=/ruta/disco.mia
func (vc *ViewerController) ListPartitions(ctx *gin.Context) {
	diskPath := ctx.Query("path")

	if diskPath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "path parameter required"})
		return
	}

	// Leer el MBR del disco
	file, err := os.Open(diskPath)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "disk not found"})
		return
	}
	defer file.Close()

	var mbr models.MBR
	err = binary.Read(file, binary.LittleEndian, &mbr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read MBR"})
		return
	}

	mounts := vc.mounts.List()
	partitions := []gin.H{}

	// Leer particiones primarias y extendidas
	for i := 0; i < 4; i++ {
		partition := mbr.Partitions[i]

		// Verificar si la partición existe (status != 0)
		if partition.Status == 0 {
			continue
		}

		// Obtener tipo de partición
		partType := "Primaria"
		if partition.Type == 'E' || partition.Type == 'e' {
			partType = "Extendida"
		}

		// Obtener fit
		fit := string(partition.Fit)
		if partition.Fit == 'F' || partition.Fit == 'f' {
			fit = "FF"
		} else if partition.Fit == 'B' || partition.Fit == 'b' {
			fit = "BF"
		} else if partition.Fit == 'W' || partition.Fit == 'w' {
			fit = "WF"
		}

		// Obtener nombre (limpiar bytes nulos)
		name := string(partition.Name[:])
		name = strings.TrimRight(name, "\x00")

		// Verificar si está montada
		var mountID string
		formatted := false
		for _, m := range mounts {
			if m.Path == diskPath && m.Name == name {
				mountID = m.ID
				formatted = true
				break
			}
		}

		partitions = append(partitions, gin.H{
			"id":        mountID,
			"name":      name,
			"type":      partType,
			"size":      formatBytes(int64(partition.Size)),
			"fit":       fit,
			"formatted": formatted,
		})

		// Si es extendida, leer particiones lógicas
		if partition.Type == 'E' || partition.Type == 'e' {
			logicals := readLogicalPartitions(file, partition.Start)
			for _, logical := range logicals {
				// Verificar si está montada
				var logicalMountID string
				logicalFormatted := false
				for _, m := range mounts {
					if m.Path == diskPath && m.Name == logical.Name {
						logicalMountID = m.ID
						logicalFormatted = true
						break
					}
				}

				partitions = append(partitions, gin.H{
					"id":        logicalMountID,
					"name":      logical.Name,
					"type":      "Lógica",
					"size":      formatBytes(logical.Size),
					"fit":       logical.Fit,
					"formatted": logicalFormatted,
				})
			}
		}
	}

	ctx.JSON(http.StatusOK, partitions)
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
		// Construir path absoluto para cada entrada
		fullPath := path
		if fullPath == "/" {
			fullPath = "/" + e.Name
		} else {
			fullPath = path + "/" + e.Name
		}

		jsonEntries = append(jsonEntries, gin.H{
			"name":  e.Name,
			"type":  e.Type,
			"path":  fullPath,
			"size":  e.Size,
			"perm":  e.Perm,
			"uid":   0, // TODO: Extraer del inodo si se requiere
			"gid":   0, // TODO: Extraer del inodo si se requiere
			"owner": e.Owner,
			"group": e.Group,
			"mtime": e.Mtime,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items": jsonEntries,
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
			"message": err.Error(),
		})
		return
	}

	// Verificar que la partición está montada
	_, err := vc.mounts.GetMount(req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "Partición no montada o ID inválido",
		})
		return
	}

	// Leer users.txt para validar credenciales
	// Usamos Cat directamente como lo hace UsersAdapter
	pathParts := [][]string{{"users.txt"}}
	content, err := vc.fs.Cat(req.ID, pathParts)
	if err != nil {
		// Mejorar mensaje de error para particiones no formateadas
		errMsg := err.Error()
		if strings.Contains(errMsg, "inodo 0") || strings.Contains(errMsg, "invalid argument") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "La partición no está formateada. Ejecuta 'mkfs' primero para crear el sistema de archivos.",
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error leyendo archivo de usuarios: " + errMsg,
			})
		}
		return
	}

	// Parsear users.txt
	lines := strings.Split(strings.TrimSpace(content), "\n")
	authenticated := false
	var uid, gid int
	var isRoot bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 5 {
			continue
		}

		// Formato: uid, tipo (U/G), grupo, usuario, contraseña
		typ := parts[1]
		if typ != "U" {
			continue // Solo usuarios
		}

		username := parts[3]
		password := parts[4]

		if username == req.User && password == req.Pass {
			// Convertir uid
			uidVal, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}
			uid = uidVal

			// El campo parts[2] es el NOMBRE del grupo, no el GID numérico
			// Necesitamos buscar el GID del grupo en las líneas de tipo "G"
			// Por ahora, usamos el mismo valor que el UID
			gid = uid

			isRoot = (req.User == "root")
			authenticated = true
			break
		}
	}

	if !authenticated {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Usuario o contraseña incorrectos",
		})
		return
	}

	// Crear sesión usando SessionStore
	vc.sess.Login(req.User, uid, gid, req.ID)

	ctx.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"user": req.User,
		"id":   req.ID,
		"uid":  uid,
		"gid":  gid,
		"root": isRoot,
	})
}

// Logout cierra la sesión de un usuario
// POST /api/auth/logout
func (vc *ViewerController) Logout(ctx *gin.Context) {
	// Cerrar sesión usando SessionStore
	vc.sess.Logout()

	ctx.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Sesión cerrada exitosamente",
	})
}
