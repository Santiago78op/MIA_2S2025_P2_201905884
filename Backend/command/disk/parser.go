package disk

import (
	"fmt"
	"strings"
)

type MkDiskArgs struct {
	Size int64
	Unit rune // 'K' | 'M'
	Fit  rune // 'F' (FF) | 'B' (BF) | 'W' (WF)
	Path string
}

type RmDiskArgs struct {
	Path string
}

type FDiskArgs struct {
	Size int64
	Unit rune // 'K' | 'M' (por defecto K/M según enunciado)
	Type rune // 'P' | 'E' | 'L'
	Fit  rune // 'F' | 'B' | 'W'
	Path string
	Name string
}

type MountArgs struct {
	Path string
	Name string
	// CarnetLastTwo lo inyecta el servicio desde config.
}

// ParseLine retorna (cmd, flags) desde una línea tipo: mkdisk -size=50 -unit=M -path="/tmp/Disco1.mia"
func ParseLine(line string) (string, map[string]string) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", nil
	}
	parts := splitPreservingQuotes(line)
	if len(parts) == 0 {
		return "", nil
	}
	cmd := strings.ToLower(parts[0])
	flags := map[string]string{}
	for _, p := range parts[1:] {
		if !strings.HasPrefix(p, "-") {
			continue
		}
		p = strings.TrimPrefix(p, "-")
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			flags[strings.ToLower(k)] = "true"
			continue
		}
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"`)
		flags[strings.ToLower(k)] = v
	}
	return cmd, flags
}

func splitPreservingQuotes(s string) []string {
	var res []string
	var cur strings.Builder
	inQuotes := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			inQuotes = !inQuotes
			cur.WriteByte(c)
		case ' ':
			if inQuotes {
				cur.WriteByte(c)
			} else if cur.Len() > 0 {
				res = append(res, strings.TrimSpace(cur.String()))
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		res = append(res, strings.TrimSpace(cur.String()))
	}
	return res
}

// Helpers de obtención tipada
func mustRune(flags map[string]string, key string, def rune, allowed string) (rune, error) {
	if v, ok := flags[key]; ok && v != "" {
		r := rune(strings.ToUpper(v)[0])
		if !strings.ContainsRune(allowed, r) {
			return 0, fmt.Errorf("%s inválido: %q", key, v)
		}
		return r, nil
	}
	if def != 0 {
		return def, nil
	}
	return 0, fmt.Errorf("falta %s", key)
}
