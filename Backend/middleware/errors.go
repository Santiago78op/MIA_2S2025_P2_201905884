package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Interface opcional para errores con código HTTP propio.
type StatusCoder interface {
	StatusCode() int
}

// Estructura de salida uniforme.
type apiError struct {
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
	Code      string `json:"code,omitempty"` // opcional para categorizar
}

// ErrorHandler convierte c.Errors en JSON consistente.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		// Tomamos el primer error (o necesito agregarlos todos /pendiente).
		err := c.Errors[0].Err
		status := http.StatusBadRequest

		// Si el error implementa StatusCoder, respétalo.
		var sc StatusCoder
		if errors.As(err, &sc) {
			status = sc.StatusCode()
		} else {
			// Heurísticas simples
			if errors.Is(err, gin.ErrorTypeBind) {
				status = http.StatusBadRequest
			}
			// Se deben añadir más mapeos si te conviene (os.ErrNotExist -> 404, etc.)
		}

		reqID := c.Writer.Header().Get("X-Request-Id")
		if reqID == "" {
			reqID = c.GetString("request_id")
		}

		c.JSON(status, apiError{
			Message:   err.Error(),
			RequestID: reqID,
		})
	}
}
