package main

import (
	"MIA_2S2025_P2_201905884/internal/commands"
	"MIA_2S2025_P2_201905884/internal/disk"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// handleExecuteCommand maneja la ejecución de un comando individual
func (s *Server) handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RunCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, RunCommandResponse{
			OK:    false,
			Error: "invalid json body: " + err.Error(),
		})
		return
	}

	if req.Line == "" {
		writeJSON(w, http.StatusBadRequest, RunCommandResponse{
			OK:    false,
			Error: "no command provided",
		})
		return
	}

	// Ejecutar el comando
	output, err := s.adapter.Run(r.Context(), req.Line)
	if err != nil {
		log.Printf("[cmd] error: %v", err)
		writeJSON(w, http.StatusOK, RunCommandResponse{
			OK:     false,
			Output: output,
			Error:  err.Error(),
			Input:  req.Line,
		})
		return
	}

	writeJSON(w, http.StatusOK, RunCommandResponse{
		OK:     true,
		Output: output,
		Input:  req.Line,
	})
}

// handleExecuteScript maneja la ejecución de múltiples comandos (script)
func (s *Server) handleExecuteScript(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ScriptResponse{
			OK:    false,
			Error: "invalid json body: " + err.Error(),
		})
		return
	}

	if req.Script == "" {
		writeJSON(w, http.StatusBadRequest, ScriptResponse{
			OK:    false,
			Error: "no script provided",
		})
		return
	}

	// Dividir el script en líneas
	lines := strings.Split(req.Script, "\n")
	var results []CommandResult
	successCount := 0
	errorCount := 0

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Saltar líneas vacías y comentarios
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Ejecutar comando
		output, err := s.adapter.Run(r.Context(), line)

		result := CommandResult{
			Line:   i + 1,
			Input:  line,
			Output: output,
		}

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			errorCount++
		} else {
			result.Success = true
			successCount++
		}

		results = append(results, result)
	}

	writeJSON(w, http.StatusOK, ScriptResponse{
		OK:           errorCount == 0,
		Results:      results,
		TotalLines:   len(lines),
		Executed:     len(results),
		SuccessCount: successCount,
		ErrorCount:   errorCount,
	})
}

// handleListMounted lista todas las particiones montadas
func (s *Server) handleListMounted(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mounted, err := s.adapter.DM.ListMounted(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	// Convertir a formato más amigable
	var partitions []map[string]string
	for _, ref := range mounted {
		partitions = append(partitions, map[string]string{
			"disk_path":    ref.DiskPath,
			"partition_id": ref.PartitionID,
			"mount_id":     commands.MakeID(ref.DiskPath, ref.PartitionID),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"partitions": partitions,
		"count":      len(partitions),
	})
}

// handleListDisks lista todos los archivos .mia en un directorio
func (s *Server) handleListDisks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Obtener el directorio desde query params
	searchPath := r.URL.Query().Get("path")
	if searchPath == "" {
		searchPath = "." // Directorio actual por defecto
	}

	files, err := os.ReadDir(searchPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": fmt.Sprintf("cannot read directory: %v", err),
		})
		return
	}

	var disks []map[string]interface{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".mia") {
			info, err := file.Info()
			if err != nil {
				continue
			}

			disks = append(disks, map[string]interface{}{
				"name":     file.Name(),
				"path":     filepath.Join(searchPath, file.Name()),
				"size":     info.Size(),
				"modified": info.ModTime().Format(time.RFC3339),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"disks":       disks,
		"count":       len(disks),
		"search_path": searchPath,
	})
}

// handleGetDiskInfo obtiene información detallada de un disco
func (s *Server) handleGetDiskInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	diskPath := r.URL.Query().Get("path")
	if diskPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "disk path is required",
		})
		return
	}

	// Verificar que el archivo existe
	info, err := os.Stat(diskPath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"ok":    false,
			"error": "disk not found",
		})
		return
	}

	// Abrir el archivo y leer el MBR
	file, err := os.Open(diskPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"ok":    false,
			"error": "cannot open disk file",
		})
		return
	}
	defer file.Close()

	var mbr disk.MBR
	if err := disk.ReadStruct(file, 0, &mbr); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"ok":    false,
			"error": "cannot read MBR",
		})
		return
	}

	// Construir información de particiones
	var partitions []map[string]interface{}
	for i, part := range mbr.Parts {
		if part.Status == disk.PartStatusUsed {
			partName := strings.TrimRight(string(part.Name[:]), "\x00")
			partitions = append(partitions, map[string]interface{}{
				"index": i,
				"name":  partName,
				"type":  getPartitionTypeName(part.Type),
				"fit":   getFitName(part.Fit),
				"start": part.Start,
				"size":  part.Size,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"path":       diskPath,
		"size":       info.Size(),
		"modified":   info.ModTime().Format(time.RFC3339),
		"mbr_size":   mbr.SizeBytes,
		"created_at": time.Unix(mbr.CreatedAt, 0).Format(time.RFC3339),
		"signature":  mbr.DiskSig,
		"fit":        getFitName(mbr.Fit),
		"partitions": partitions,
	})
}

// Helper para obtener el nombre del tipo de partición
func getPartitionTypeName(ptype byte) string {
	switch ptype {
	case disk.PartTypePrimary:
		return "Primary"
	case disk.PartTypeExtended:
		return "Extended"
	case disk.PartTypeLogical:
		return "Logical"
	default:
		return "Unknown"
	}
}

// Helper para obtener el nombre del fit
func getFitName(fit byte) string {
	switch fit {
	case disk.FitFF:
		return "First Fit"
	case disk.FitBF:
		return "Best Fit"
	case disk.FitWF:
		return "Worst Fit"
	default:
		return "Unknown"
	}
}

// handleValidateCommand valida la sintaxis de un comando sin ejecutarlo
func (s *Server) handleValidateCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RunCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, RunCommandResponse{
			OK:    false,
			Error: "invalid json body",
		})
		return
	}

	// Parsear el comando
	handler, err := commands.ParseCommand(req.Line)
	if err != nil {
		writeJSON(w, http.StatusOK, RunCommandResponse{
			OK:    false,
			Error: err.Error(),
			Input: req.Line,
		})
		return
	}

	// Validar el comando
	if err := handler.Validate(); err != nil {
		writeJSON(w, http.StatusOK, RunCommandResponse{
			OK:    false,
			Error: err.Error(),
			Usage: commands.Usage(handler.Name()),
			Input: req.Line,
		})
		return
	}

	writeJSON(w, http.StatusOK, RunCommandResponse{
		OK:      true,
		Output:  "Command syntax is valid",
		Input:   req.Line,
		Command: string(handler.Name()),
	})
}

// handleGetCommands devuelve la lista de comandos soportados
func (s *Server) handleGetCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	commandList := map[string][]string{
		"disk": {
			"mkdisk -path <path> -size <size> [-unit b|k|m] [-fit bf|ff|wf]",
			"fdisk -path <path> -mode add|delete -name <name> [-size <size>] [-unit b|k|m] [-type p|e|l] [-fit bf|ff|wf]",
			"mount -path <path> -name <name>",
			"unmount -id <id>",
		},
		"filesystem": {
			"mkfs -id <id> -fs 2fs|3fs",
		},
		"files": {
			"mkdir -id <id> -path <path> [-p]",
			"mkfile -id <id> -path <path> [-cont <content>] [-size <size>]",
			"remove -id <id> -path <path>",
			"edit -id <id> -path <path> -cont <content> [-append]",
			"rename -id <id> -from <from> -to <to>",
			"copy -id <id> -from <from> -to <to>",
			"move -id <id> -from <from> -to <to>",
			"find -id <id> [-base <path>] [-name <pattern>] [-limit <n>]",
			"chown -id <id> -path <path> -user <user> -group <group>",
			"chmod -id <id> -path <path> -perm <permissions>",
		},
		"ext3": {
			"journaling -id <id>",
			"recovery -id <id>",
			"loss -id <id>",
		},
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"commands": commandList,
	})
}
