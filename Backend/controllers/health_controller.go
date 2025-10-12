package controllers

import (
	"net/http"

	"backend/config"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
	cfg *config.Config
}

func NewHealthController(cfg *config.Config) *HealthController {
	return &HealthController{cfg: cfg}
}

func (h *HealthController) Get(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"port":         h.cfg.Port,
		"disks_path":   h.cfg.DisksPath,
		"reports_path": h.cfg.ReportsPath,
		"debug":        h.cfg.Debug,
	})
}
