package reports

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// WriteDOT guarda el DOT a un archivo .dot.
func WriteDOT(path string, dot string) error {
	return os.WriteFile(path, []byte(dot), 0o664)
}

// RenderWithGraphviz llama al binario `dot` para generar PNG o SVG.
// format: "png" | "svg"
func RenderWithGraphviz(dotStr, outPath, format string) error {
	if format != "png" && format != "svg" {
		return fmt.Errorf("formato inválido: %s", format)
	}
	cmd := exec.Command("dot", "-T"+format, "-o", outPath)
	cmd.Stdin = bytes.NewBufferString(dotStr)
	cmd.Dir = filepath.Dir(outPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("graphviz dot error: %w", err)
	}
	return nil
}
