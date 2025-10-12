package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ScriptRunner ejecuta un .smia completo (respetando # y líneas vacías).
// Debe concatenar los outputs (o devolver los agregados) según tu implementación.
type ScriptRunner interface {
	Run(script string, stopOnError bool) (string, error)
}

type ScriptController struct {
	runner ScriptRunner
}

func NewScriptController(runner ScriptRunner) *ScriptController {
	return &ScriptController{runner: runner}
}

type runScriptReq struct {
	Script      string `json:"script"`      // contenido completo .smia
	StopOnError bool   `json:"stopOnError"` // si true, se detiene en el primer error
}

type runScriptRes struct {
	Output string `json:"output"` // salida agregada (línea a línea)
}

func (s *ScriptController) RunSMIA(ctx *gin.Context) {
	var req runScriptReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(err)
		return
	}
	script := strings.TrimSpace(req.Script)
	if script == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "script vacío"})
		return
	}

	out, err := s.runner.Run(script, req.StopOnError)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, runScriptRes{Output: out})
}
