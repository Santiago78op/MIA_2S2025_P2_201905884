package main

import (
	"MIA_2S2025_P2_201905884/internal/commands"
	"MIA_2S2025_P2_201905884/internal/disk"
	"MIA_2S2025_P2_201905884/internal/fs"
	"MIA_2S2025_P2_201905884/internal/fs/ext2"
	"MIA_2S2025_P2_201905884/internal/fs/ext3"
	"MIA_2S2025_P2_201905884/internal/journal"
	"MIA_2S2025_P2_201905884/internal/logger"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// ===== Config =====
	port := getenv("PORT", "8080")
	allowOrigin := getenv("ALLOW_ORIGIN", "*")
	logFile := getenv("LOG_FILE", "godisk.log")

	// Inicializar logger
	if err := logger.Init(logFile, 1000, true); err != nil {
		log.Fatalf("[main] failed to init logger: %v", err)
	}
	defer logger.GetLogger().Close()

	logger.Info("Starting GoDisk 2.0 Server", map[string]interface{}{
		"port":   port,
		"origin": allowOrigin,
	})

	// ===== Wiring de dependencias =====
	// Inicializar disk manager
	dm := disk.NewManager()

	// Inicializar metadata state para filesystems
	meta := fs.NewMetaState()
	fs2 := ext2.New(meta)

	// Crear journal store temporal (implementar más tarde si es necesario)
	var jstore journal.Store = nil
	fs3 := ext3.New(meta, 128, 50, jstore) // blockSize=128, journal=50

	// Índice de montajes
	idx := commands.NewMemoryIndex()

	adapter := &commands.Adapter{
		FS2:   fs2,
		FS3:   fs3,
		DM:    dm,
		Index: idx,
		State: meta,
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
