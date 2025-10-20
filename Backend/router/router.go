package router

import (
	"fmt"
	"net/http"
	"time"

	"Backend/config"
	"Backend/controllers"

	"github.com/gin-gonic/gin"
)

// CORS mínimo: ajusta origins si tu front corre en otro puerto/host.
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*") // o fija tu origen: http://localhost:5173
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Disposition")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Límite simple al tamaño de cuerpo (p. ej., 10MB)
func maxBodySize(limitBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limitBytes)
		c.Next()
	}
}

// Logger sencillo (si no usas zap/logrus)
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		lat := time.Since(start)
		status := c.Writer.Status()
		fmt.Printf("[HTTP] %s %s -> %d (%s)\n", method, path, status, lat)
	}
}

// Error handler JSON consistente
func errorHandler() gin.HandlerFunc {
	type apiError struct {
		Message string `json:"message"`
	}
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		// tomamos el primer error
		err := c.Errors[0].Err
		c.JSON(http.StatusBadRequest, apiError{Message: err.Error()})
	}
}

//
// Router
//

// SetupRouter construye el *gin.Engine con middlewares y rutas.
func SetupRouter(cfg *config.Config,
	cs *controllers.CommandsController,
	sc *controllers.ScriptController,
	rc *controllers.ReportsController,
	vc *controllers.ViewerController, // NUEVO: ViewerController para P2
) *gin.Engine {

	// Modo Gin según DEBUG
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())
	r.Use(errorHandler())
	r.Use(cors())
	r.Use(maxBodySize(10 << 20)) // 10 MB

	// Healthcheck
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "ok",
			"disks_path":   cfg.DisksPath,
			"reports_path": cfg.ReportsPath,
			"debug":        cfg.Debug,
		})
	})

	// Servir archivos estáticos de reportes (Graphviz, bm_*, etc.)
	// Quedarán disponibles en: GET /reports/static/<archivo>
	// Ej: /reports/static/tree_841A.jpg
	r.StaticFS("/reports/static", gin.Dir(cfg.ReportsPath, false))

	// API principal
	api := r.Group("/api")
	{
		// Ejecutar un solo comando
		// Body: { "input": "mkdisk -size=50 -unit=M -path=\\".../Disco1.mia\\"" }
		api.POST("/commands", cs.RunOnce)

		// Ejecutar script .smia completo
		// Body: { "script": "...", "stopOnError": true }
		api.POST("/script", sc.RunSMIA)

		// Generar reporte (mbr, disk, inode, block, bm_inode, bm_block, tree, sb, file, ls)
		// Body: { "name": "mbr", "id": "841A", "out": "/abs/or/rel/path.jpg", "extra": "" }
		api.POST("/reports", rc.Generate)

		// Listar todos los archivos en la carpeta Reports
		// Response: { "files": [...], "count": N }
		api.GET("/reports/list", rc.List)

		// === NUEVAS RUTAS P2: Autenticación ===
		auth := api.Group("/auth")
		{
			// Login: autenticar usuario en una partición
			// POST Body: { "user": "username", "pass": "password", "id": "mount_id" }
			auth.POST("/login", vc.Login)

			// Logout: cerrar sesión
			// POST Body: { "id": "mount_id" }
			auth.POST("/logout", vc.Logout)
		}

		// === NUEVAS RUTAS P2: Visualizador ===
		// Listar discos disponibles
		api.GET("/disks", vc.ListDisks)

		// Listar particiones de un disco
		api.GET("/disks/:disk/partitions", vc.ListPartitions)

		// Árbol de directorios de una partición montada
		// Query params: ?path=/ruta
		api.GET("/fs/:id/tree", vc.GetTree)

		// Contenido de un archivo
		// Query params: ?path=/archivo.txt
		api.GET("/fs/:id/file", vc.GetFile)

		// Listar entradas del journal (solo EXT3)
		api.GET("/journal/:id", vc.GetJournal)

		// Listar entradas del journal en formato tabla (solo EXT3)
		api.GET("/journal/:id/table", vc.GetJournalTable)
	}

	// 404 Handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "ruta no encontrada",
			"path":    c.Request.URL.Path,
		})
	})

	return r
}
