package fs

import (
	"fmt"
	"strings"
)

type MkfsArgs struct {
	ID   string
	Fs   string // "2fs" | "3fs" (por defecto "2fs")
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

	// Parsear sistema de archivos (opcional, por defecto "2fs")
	fsVal := strings.ToLower(flags["fs"])
	if fsVal == "" {
		fsVal = "2fs" // Valor por defecto
	}

	// Validar que el fs sea válido
	if fsVal != "2fs" && fsVal != "3fs" {
		return args, fmt.Errorf("sistema de archivos no válido: %s (solo se acepta '2fs' o '3fs')", fsVal)
	}

	args.Fs = fsVal

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

// ============================================
// PARSERS NUEVOS P2
// ============================================

type RemoveArgs struct {
	ID   string
	Path string
}

type EditArgs struct {
	ID      string
	Path    string
	Content string
}

type RenameArgs struct {
	ID   string
	Path string
	Name string
}

type CopyArgs struct {
	ID       string
	SrcPath  string
	DestPath string
}

type MoveArgs struct {
	ID       string
	SrcPath  string
	DestPath string
}

type FindArgs struct {
	ID   string
	Path string
	Name string
}

type ChmodArgs struct {
	ID        string
	Path      string
	Ugo       string
	Recursive bool
}

type ChownArgs struct {
	ID        string
	Path      string
	User      string
	Recursive bool
}

type LossArgs struct {
	ID string
}

type RecoveryArgs struct {
	ID string
}

func ParseRemove(line string) (RemoveArgs, error) {
	_, flags := parseLine(line)
	var args RemoveArgs

	// ID es opcional; si no se proporciona, se usa el de la sesión actual
	args.ID = flags["id"] // puede ser ""

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	return args, nil
}

func ParseEdit(line string) (EditArgs, error) {
	_, flags := parseLine(line)
	var args EditArgs

	// ID es opcional; si no se proporciona, se usa el de la sesión actual
	args.ID = flags["id"] // puede ser ""

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	content, err := mustString(flags, "contenido")
	if err != nil {
		return args, err
	}
	args.Content = content

	return args, nil
}

func ParseRename(line string) (RenameArgs, error) {
	_, flags := parseLine(line)
	var args RenameArgs

	// ID es opcional; si no se proporciona, se usa el de la sesión actual
	args.ID = flags["id"] // puede ser ""

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	name, err := mustString(flags, "name")
	if err != nil {
		return args, err
	}
	args.Name = name

	return args, nil
}

func ParseCopy(line string) (CopyArgs, error) {
	_, flags := parseLine(line)
	var args CopyArgs

	// ID es opcional; si no se proporciona, se usa el de la sesión actual
	args.ID = flags["id"] // puede ser ""

	src, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.SrcPath = src

	// Aceptar tanto -dest como -destino
	dest := flags["dest"]
	if dest == "" {
		dest = flags["destino"]
	}
	if dest == "" {
		return args, fmt.Errorf("falta parámetro -dest o -destino")
	}
	args.DestPath = dest

	return args, nil
}

func ParseMove(line string) (MoveArgs, error) {
	_, flags := parseLine(line)
	var args MoveArgs

	// ID es opcional; si no se proporciona, se usa el de la sesión actual
	args.ID = flags["id"] // puede ser ""

	src, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.SrcPath = src

	// Aceptar tanto -dest como -destino
	dest := flags["dest"]
	if dest == "" {
		dest = flags["destino"]
	}
	if dest == "" {
		return args, fmt.Errorf("falta parámetro -dest o -destino")
	}
	args.DestPath = dest

	return args, nil
}

func ParseFind(line string) (FindArgs, error) {
	_, flags := parseLine(line)
	var args FindArgs

	// ID es opcional, se tomará de la sesión si no se proporciona
	args.ID = flags["id"]

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	name, err := mustString(flags, "name")
	if err != nil {
		return args, err
	}
	args.Name = name

	return args, nil
}

func ParseChmod(line string) (ChmodArgs, error) {
	_, flags := parseLine(line)
	var args ChmodArgs

	// ID es opcional, se tomará de la sesión si no se proporciona
	args.ID = flags["id"]

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	ugo, err := mustString(flags, "ugo")
	if err != nil {
		return args, err
	}
	args.Ugo = ugo

	args.Recursive = flags["r"] == "true"

	return args, nil
}

func ParseChown(line string) (ChownArgs, error) {
	_, flags := parseLine(line)
	var args ChownArgs

	// ID es opcional, se tomará de la sesión si no se proporciona
	args.ID = flags["id"]

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	user, err := mustString(flags, "user")
	if err != nil {
		return args, err
	}
	args.User = user

	args.Recursive = flags["r"] == "true"

	return args, nil
}

func ParseLoss(line string) (LossArgs, error) {
	_, flags := parseLine(line)
	var args LossArgs

	id, err := mustString(flags, "id")
	if err != nil {
		return args, err
	}
	args.ID = id

	return args, nil
}

func ParseRecovery(line string) (RecoveryArgs, error) {
	_, flags := parseLine(line)
	var args RecoveryArgs

	id, err := mustString(flags, "id")
	if err != nil {
		return args, err
	}
	args.ID = id

	return args, nil
}
