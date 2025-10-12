package graphviz

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Dot struct{}

func New() *Dot { return &Dot{} }

// Generate ejecuta `dot` con el contenido DOT (graph) y escribe la imagen en outPath.
// outPath puede ser .png/.jpg/.svg según extensión.
func (d *Dot) Generate(dotContent string, outPath string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", err
	}
	// dot -T<ext> -o outPath
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(outPath)), ".")
	if ext == "" {
		ext = "png"
	}
	cmd := exec.Command("dot", "-T"+ext, "-o", outPath)
	cmd.Stdin = strings.NewReader(dotContent)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("graphviz error: %v | %s", err, string(out))
	}
	return outPath, nil
}
