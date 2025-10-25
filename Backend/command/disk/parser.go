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

func mustInt64(flags map[string]string, key string) (int64, error) {
	v, ok := flags[key]
	if !ok {
		return 0, fmt.Errorf("falta %s", key)
	}
	var val int64
	_, err := fmt.Sscanf(v, "%d", &val)
	if err != nil {
		return 0, fmt.Errorf("%s inválido: %s", key, v)
	}
	return val, nil
}

func mustString(flags map[string]string, key string) (string, error) {
	v, ok := flags[key]
	if !ok || v == "" {
		return "", fmt.Errorf("falta %s", key)
	}
	return v, nil
}

// validateNoUnknownFlags valida que no haya parámetros desconocidos
func validateNoUnknownFlags(flags map[string]string, allowedKeys []string) error {
	allowed := make(map[string]bool)
	for _, k := range allowedKeys {
		allowed[strings.ToLower(k)] = true
	}

	for key := range flags {
		if !allowed[key] {
			return fmt.Errorf("parámetro desconocido: -%s", key)
		}
	}
	return nil
}

// ParseMkDisk parsea un comando mkdisk
func ParseMkDisk(line string) (MkDiskArgs, error) {
	_, flags := ParseLine(line)
	var args MkDiskArgs

	// Validar que no haya parámetros desconocidos
	if err := validateNoUnknownFlags(flags, []string{"size", "unit", "fit", "path"}); err != nil {
		return args, err
	}

	size, err := mustInt64(flags, "size")
	if err != nil {
		return args, err
	}
	args.Size = size

	unit, err := mustRune(flags, "unit", 'M', "KM")
	if err != nil {
		return args, err
	}
	args.Unit = unit

	fit, err := mustRune(flags, "fit", 'F', "FBW")
	if err != nil {
		return args, err
	}
	args.Fit = fit

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	return args, nil
}

// ParseRmDisk parsea un comando rmdisk
func ParseRmDisk(line string) (RmDiskArgs, error) {
	_, flags := ParseLine(line)
	var args RmDiskArgs

	// Validar que no haya parámetros desconocidos
	if err := validateNoUnknownFlags(flags, []string{"path"}); err != nil {
		return args, err
	}

	path, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Path = path

	return args, nil
}

// ParseFDisk parsea un comando fdisk
func ParseFDisk(line string) (FDiskArgs, error) {
	_, flags := ParseLine(line)
	var args FDiskArgs

	// Validar que no haya parámetros desconocidos
	if err := validateNoUnknownFlags(flags, []string{"size", "unit", "type", "fit", "path", "name"}); err != nil {
		return args, err
	}

	size, err := mustInt64(flags, "size")
	if err != nil {
		return args, err
	}
	args.Size = size

	unit, err := mustRune(flags, "unit", 'K', "BKM")
	if err != nil {
		return args, err
	}
	args.Unit = unit

	typ, err := mustRune(flags, "type", 'P', "PEL")
	if err != nil {
		return args, err
	}
	args.Type = typ

	fit, err := mustRune(flags, "fit", 'W', "FBW")
	if err != nil {
		return args, err
	}
	args.Fit = fit

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

// ParseMount parsea un comando mount
func ParseMount(line string) (MountArgs, error) {
	_, flags := ParseLine(line)
	var args MountArgs

	// Validar que no haya parámetros desconocidos
	if err := validateNoUnknownFlags(flags, []string{"path", "name"}); err != nil {
		return args, err
	}

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


// ============================================
// PARSERS NUEVOS P2
// ============================================

// UnmountArgs representa los argumentos para unmount
type UnmountArgs struct {
	ID string
}

// ParseUnmount parsea un comando unmount
func ParseUnmount(line string) (UnmountArgs, error) {
	_, flags := ParseLine(line)
	var args UnmountArgs

	id, err := mustString(flags, "id")
	if err != nil {
		return args, err
	}
	args.ID = id

	return args, nil
}
