package fs

import (
	"path"
	"strings"
)

// CleanPath normaliza rutas al estilo UNIX. Devuelve error si no inicia en "/".
func CleanPath(p string) (string, error) {
	if p == "" || !strings.HasPrefix(p, "/") {
		return "", ErrInvalidPath
	}
	// path.Clean ya resuelve .. y .
	cp := path.Clean(p)
	// mantener "/" para raíz
	if cp == "." {
		cp = "/"
	}
	return cp, nil
}

// SplitParts divide /a/b/c en ["a","b","c"] (raíz -> [] vacía)
func SplitParts(p string) ([]string, error) {
	cp, err := CleanPath(p)
	if err != nil {
		return nil, err
	}
	if cp == "/" {
		return []string{}, nil
	}
	return strings.Split(strings.TrimPrefix(cp, "/"), "/"), nil
}
