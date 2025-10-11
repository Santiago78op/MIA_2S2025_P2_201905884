package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"MIA_2S2025_P2_201905884/internal/auth"
	"MIA_2S2025_P2_201905884/internal/commands"
	"MIA_2S2025_P2_201905884/internal/disk"
	"MIA_2S2025_P2_201905884/internal/fs"
	"MIA_2S2025_P2_201905884/internal/fs/ext2"
	"MIA_2S2025_P2_201905884/internal/fs/ext3"
	"MIA_2S2025_P2_201905884/internal/logger"
	"MIA_2S2025_P2_201905884/internal/reports"
)

func main() {
	// ===== Config =====
	port := getenv("PORT", "8080")
	allowOrigin := getenv("ALLOW_ORIGIN", "*")
	logFile := getenv("LOG_FILE", "Logs/godisk.log")

	// Inicializar logger
	if err := logger.Init(logFile, 1000, true); err != nil {
		log.Fatalf("[main] failed to init logger: %v", err)
	}
	defer logger.GetLogger().Close()

	logger.Info("Starting GoDisk 2.0 Server", map[string]interface{}{
		"port":     port,
		"origin":   allowOrigin,
		"log_file": logFile,
	})

	// ===== Wiring de dependencias =====
	// Inicializar disk manager
	dm := disk.NewManager()

	// Inicializar metadata state para filesystems
	meta := fs.NewMetaState()
	fs2 := ext2.New(meta)
	fs3 := ext3.New(meta, 128, nil) // blockSize=128

	// Índice de montajes (limpio en cada inicio para IDs predecibles)
	idx := commands.NewMemoryIndex()
	// Ya está limpio al crearse, no necesita Reset() adicional

	// Inicializar sesión y reportes para P1
	session := auth.NewSessionManager(fs2) // Usar fs2 para validación de credenciales
	reportGen := reports.NewSimpleGenerator()

	adapter := &commands.Adapter{
		FS2:     fs2,
		FS3:     fs3,
		DM:      dm,
		Index:   idx,
		State:   meta,
		Session: session,
		Reports: reportGen,
	}

	// ===== HTTP Server =====
	mux := http.NewServeMux()
	s := NewServer(adapter, allowOrigin)
	registerRoutes(mux, s)

	// Aplicar middleware de logging
	handler := LoggingMiddleware(mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  75 * time.Second,
	}

	logger.Info("HTTP Server listening", map[string]interface{}{"port": port})

	// Shutdown gracioso
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[server] fatal: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Graceful shutdown failed", map[string]interface{}{"error": err.Error()})
	}
	logger.Info("Server stopped successfully")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
