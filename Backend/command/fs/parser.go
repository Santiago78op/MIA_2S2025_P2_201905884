package fs

import (
	"fmt"
	"strings"
)

type MkfsArgs struct {
	ID   string
	Type string // "full" (por defecto)
}

type MkdirArgs struct {
	ID      string
	Path    string
	Parents bool // -r flag
}

type MkfileArgs struct {
	ID        string
	Path      string
	Size      int
	Cont      string // path al archivo host
	Recursive bool   // -r flag
}

type CatArgs struct {
	ID    string
	Files []string
}

// Parse usando el mismo enfoque que disk
func parseLine(line string) (string, map[string]string) {
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

func mustString(flags map[string]string, key string) (string, error) {
	v, ok := flags[key]
	if !ok || v == "" {
		return "", fmt.Errorf("falta %s", key)
	}
	return v, nil
}

func mustInt(flags map[string]string, key string) (int, error) {
	v, ok := flags[key]
	if !ok {
		return 0, fmt.Errorf("falta %s", key)
	}
	var val int
	_, err := fmt.Sscanf(v, "%d", &val)
	if err != nil {
		return 0, fmt.Errorf("%s inválido: %s", key, v)
	}
	return val, nil
}

func ParseMkfs(line string) (MkfsArgs, error) {
	_, flags := parseLine(line)
	var args MkfsArgs

	id, err := mustString(flags, "id")
	if err != nil {
		return args, err
	}
	args.ID = id

	// Parsear tipo de formateo (opcional, por defecto "full")
	typeVal := strings.ToLower(flags["type"])
	if typeVal == "" {
		typeVal = "full" // Valor por defecto
	}

	// Validar que el tipo sea válido
	if typeVal != "full" {
		return args, fmt.Errorf("tipo de formateo no válido: %s (solo se acepta 'full')", typeVal)
	}

	args.Type = typeVal

	return args, nil
}

func ParseMkdir(line string) (MkdirArgs, error) {
	_, flags := parseLine(line)
	var args MkdirArgs

	// -id es opcional, se usará el de la sesión actual si no se especifica
	args.ID = flags["id"]

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	args.Parents = flags["r"] == "true" || flags["p"] == "true"

	return args, nil
}

func ParseMkfile(line string) (MkfileArgs, error) {
	_, flags := parseLine(line)
	var args MkfileArgs

	// -id es opcional, se usará el de la sesión actual si no se especifica
	args.ID = flags["id"]

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	// -size es opcional, por defecto 0
	if sizeStr, ok := flags["size"]; ok && sizeStr != "" {
		size, err := mustInt(flags, "size")
		if err != nil {
			return args, err
		}
		args.Size = size
	} else {
		args.Size = 0 // Valor por defecto según la guía
	}

	args.Cont = flags["cont"] // opcional
	args.Recursive = flags["r"] == "true"

	return args, nil
}

func ParseCat(line string) (CatArgs, error) {
	_, flags := parseLine(line)
	var args CatArgs

	// -id es opcional, se usará el de la sesión actual si no se especifica
	args.ID = flags["id"]

	// Los archivos pueden venir en múltiples flags file1, file2, etc.
	// o como valores separados
	// Por simplicidad, asumimos que se pasa como -file1, -file2, etc.
	for i := 1; i <= 10; i++ {
		key := fmt.Sprintf("file%d", i)
		if f, ok := flags[key]; ok && f != "" {
			args.Files = append(args.Files, f)
		}
	}

	// También soportar un solo -file
	if f, ok := flags["file"]; ok && f != "" {
		args.Files = append(args.Files, f)
	}

	if len(args.Files) == 0 {
		return args, fmt.Errorf("falta al menos un archivo")
	}

	return args, nil
}
