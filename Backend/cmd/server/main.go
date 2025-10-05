package main

import (
	"MIA_2S2025_P2_201905884/internal/commands"
	"MIA_2S2025_P2_201905884/internal/disk"
	"MIA_2S2025_P2_201905884/internal/fs"
	"MIA_2S2025_P2_201905884/internal/fs/ext2"
	"MIA_2S2025_P2_201905884/internal/fs/ext3"
	"MIA_2S2025_P2_201905884/internal/journal"
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

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  75 * time.Second,
	}

	log.Printf("[server] listening on :%s\n", port)

	// Shutdown gracioso
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[server] fatal: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("[server] shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[server] graceful shutdown failed: %v", err)
	}
	log.Println("[server] bye")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
