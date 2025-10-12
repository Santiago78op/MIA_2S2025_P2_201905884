package router

import (
	"fmt"
	"net/http"
	"time"

	"backend/config"
	"backend/controllers"

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
