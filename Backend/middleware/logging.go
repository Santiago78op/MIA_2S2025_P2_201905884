package middleware

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestID asegura un X-Request-Id en cada request y lo expone en el contexto.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.Request.Header.Get("X-Request-Id")
		if rid == "" {
			// ID simple: ts (base36) + random (base36).
			rid = fmt.Sprintf("%s-%s",
				strings.ToLower(strconvBase36(time.Now().UnixNano())),
				strings.ToLower(strconvBase36(rand.Int64())),
			)
		}
		// Disponible en headers de respuesta y en contexto.
		c.Writer.Header().Set("X-Request-Id", rid)
		c.Set("request_id", rid)
		c.Next()
	}
}

// Logger imprime una línea por request con datos clave.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery
		ip := clientIP(c)
		ua := c.Request.UserAgent()
		rid := c.GetString("request_id")

		c.Next()

		lat := time.Since(start)
		status := c.Writer.Status()
		sz := c.Writer.Size()

		if rawQuery != "" {
			path += "?" + rawQuery
		}

		// Formato compacto, fácil de grepear.
		fmt.Printf("[HTTP] %3d %7s %9s | %-15s | %s | size=%d rid=%s ua=%q\n",
			status, method, lat.Truncate(time.Microsecond), ip, path, sz, rid, ua)
	}
}

// MaxBodySize limita el tamaño del cuerpo del request (p.ej. 10<<20 = 10MB).
func MaxBodySize(limitBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limitBytes)
		c.Next()
	}
}

// Helpers

func clientIP(c *gin.Context) string {
	// Respeta X-Forwarded-For / X-Real-IP si usas proxy.
	if xff := c.Request.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xr := c.Request.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	return c.ClientIP()
}

func strconvBase36(n int64) string {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 && i > 0 {
		i--
		// n % 36
		b[i] = alphabet[n%36]
		n /= 36
	}
	if neg && i > 0 {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
