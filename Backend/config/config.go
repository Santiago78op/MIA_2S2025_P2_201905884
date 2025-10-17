package config

import (
	"log"
	"os"
	"path/filepath"
)

// Config almacena los valores globales de configuración del backend.
type Config struct {
	Port          string // Puerto HTTP del servidor (por defecto 8080)
	DisksPath     string // Carpeta donde se crean los discos .mia
	ReportsPath   string // Carpeta donde se guardan los reportes (Graphviz, bitmaps, etc.)
	CarnetLastTwo string // Últimos 2 dígitos del carnet (para IDs de mount)
	Debug         bool   // Modo debug: imprime logs detallados
}

// Load lee las variables de entorno o usa valores por defecto.
func Load() *Config {
	// Determinar el directorio base del proyecto (Backend/)
	// Si DISKS_PATH o REPORTS_PATH no son rutas absolutas, se resuelven desde Backend/
	baseDir := getEnv("BASE_DIR", "")
	if baseDir == "" {
		// Intentar detectar la raíz del proyecto
		if wd, err := os.Getwd(); err == nil {
			// Si estamos en Backend/cmd/server, subir dos niveles
			if filepath.Base(wd) == "server" && filepath.Base(filepath.Dir(wd)) == "cmd" {
				baseDir = filepath.Join(filepath.Dir(filepath.Dir(wd)))
			} else if filepath.Base(wd) == "Backend" {
				baseDir = wd
			} else {
				// Por defecto, asumir que estamos en Backend/
				baseDir = wd
			}
		}
	}

	disksPath := getEnv("DISKS_PATH", "./Discos")
	reportsPath := getEnv("REPORTS_PATH", "./Reports")

	// Convertir rutas relativas a absolutas desde baseDir
	if !filepath.IsAbs(disksPath) && baseDir != "" {
		disksPath = filepath.Join(baseDir, disksPath)
	}
	if !filepath.IsAbs(reportsPath) && baseDir != "" {
		reportsPath = filepath.Join(baseDir, reportsPath)
	}

	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		DisksPath:     disksPath,
		ReportsPath:   reportsPath,
		CarnetLastTwo: getEnv("CARNET_LAST_TWO", "84"), // 201905884 → "84"
		Debug:         getEnv("DEBUG", "false") == "true",
	}

	// Crear carpetas si no existen
	createIfNotExists(cfg.DisksPath)
	createIfNotExists(cfg.ReportsPath)

	log.Printf("[CONFIG] Servidor en puerto %s | Discos: %s | Reportes: %s | Debug: %v\n",
		cfg.Port, cfg.DisksPath, cfg.ReportsPath, cfg.Debug)
	return cfg
}

// getEnv obtiene una variable de entorno o usa valor por defecto.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// createIfNotExists crea la carpeta si no existe.
func createIfNotExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			log.Fatalf("Error creando carpeta %s: %v", path, err)
		}
	}
}
