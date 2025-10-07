package main

import (
	"MIA_2S2025_P2_201905884/internal/commands"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Server struct {
	adapter     *commands.Adapter
	corsWrapper func(http.Handler) http.Handler
}

func NewServer(adapter *commands.Adapter, allowOrigin string) *Server {
	return &Server{
		adapter:     adapter,
		corsWrapper: CORS(allowOrigin),
	}
}

func registerRoutes(mux *http.ServeMux, s *Server) {
	// Health & Info
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/version", s.handleVersion)
	mux.HandleFunc("/api/commands", s.handleGetCommands)

	// Comandos
	mux.Handle("/api/cmd/run", s.corsWrapper(http.HandlerFunc(s.handleRunCommand)))
	mux.Handle("/api/cmd/execute", s.corsWrapper(http.HandlerFunc(s.handleExecuteCommand)))
	mux.Handle("/api/cmd/validate", s.corsWrapper(http.HandlerFunc(s.handleValidateCommand)))
	mux.Handle("/api/cmd/script", s.corsWrapper(http.HandlerFunc(s.handleExecuteScript)))

	// Discos y Particiones
	mux.HandleFunc("/api/disks", s.handleListDisks)
	mux.HandleFunc("/api/disks/info", s.handleGetDiskInfo)
	mux.HandleFunc("/api/mounted", s.handleListMounted)

	// Reportes (DOT)
	mux.Handle("/api/reports/mbr", s.corsWrapper(http.HandlerFunc(s.handleReportMBR)))
	mux.Handle("/api/reports/disk", s.corsWrapper(http.HandlerFunc(s.handleReportDisk)))
	mux.Handle("/api/reports/sb", s.corsWrapper(http.HandlerFunc(s.handleReportSuperblock)))
	mux.Handle("/api/reports/tree", s.corsWrapper(http.HandlerFunc(s.handleReportTree)))
	mux.Handle("/api/reports/journal", s.corsWrapper(http.HandlerFunc(s.handleReportJournal)))
	mux.Handle("/api/reports/generate", s.corsWrapper(http.HandlerFunc(s.handleGenerateReport)))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    "GoDisk 2.0 API",
		"version": "p2-dev",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleRunCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RunCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	// Ejecutar el comando directamente a través del adapter
	out, execErr := s.adapter.Run(r.Context(), req.Line)
	if execErr != nil {
		log.Printf("[cmd] error: %v", execErr)
		writeJSON(w, http.StatusOK, RunCommandResponse{
			OK:     false,
			Output: out,
			Error:  execErr.Error(),
			Input:  req.Line,
		})
		return
	}
	writeJSON(w, http.StatusOK, RunCommandResponse{
		OK:     true,
		Output: out,
		Input:  req.Line,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
