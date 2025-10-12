package fs

import (
	"path"
	"strings"
)

// SplitPath normaliza y parte un path absoluto del FS (no del host).
// Ej: "/home/docs/a.txt" -> ["home","docs","a.txt"]
func SplitPath(p string) []string {
	p = path.Clean(p)
	p = strings.TrimSpace(p)
	if p == "/" || p == "" {
		return nil
	}
	p = strings.TrimPrefix(p, "/")
	parts := strings.Split(p, "/")
	out := make([]string, 0, len(parts))
	for _, s := range parts {
		if s == "" || s == "." {
			continue
		}
		if s == ".." {
			// en FS real no navegas fuera de /; puedes ignorar o colapsar
			continue
		}
		out = append(out, s)
	}
	return out
}
