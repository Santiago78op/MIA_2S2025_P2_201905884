package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CommandRunner es la interfaz mínima que debe exponer tu capa de aplicación.
// Implementación típica: parsea la línea y despacha al caso de uso correcto.
type CommandRunner interface {
	Run(line string) (string, error)
}

type CommandsController struct {
	runner CommandRunner
}

func NewCommandsController(runner CommandRunner) *CommandsController {
	return &CommandsController{runner: runner}
}

type runOnceReq struct {
	Input string `json:"input"` // ej: mkdisk -size=50 -unit=M -path="/tmp/Disco1.mia"
}

type runOnceRes struct {
	Output string `json:"output"`
}

func (c *CommandsController) RunOnce(ctx *gin.Context) {
	var req runOnceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}
	line := strings.TrimSpace(req.Input)
	if line == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "input vacío"})
		return
	}

	out, err := c.runner.Run(line)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, runOnceRes{Output: out})
}
