package reports

import (
	"fmt"
	"strings"
)

type ReportArgs struct {
	Name  string // mbr, disk, inode, block, bm_inode, bm_block, tree, sb, file, ls
	ID    string
	Out   string // path de salida
	Extra string // opcional: path_file o path_file_ls
}

// Parse usando el mismo enfoque
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

func ParseReport(line string) (ReportArgs, error) {
	_, flags := parseLine(line)
	var args ReportArgs

	name, err := mustString(flags, "name")
	if err != nil {
		return args, err
	}
	args.Name = strings.ToLower(name)

	id, err := mustString(flags, "id")
	if err != nil {
		return args, err
	}
	args.ID = id

	out, err := mustString(flags, "path")
	if err != nil {
		return args, err
	}
	args.Out = out

	// Extra es opcional
	if ruta, ok := flags["ruta"]; ok && ruta != "" {
		args.Extra = ruta
	}
	if pathFile, ok := flags["path_file_ls"]; ok && pathFile != "" {
		args.Extra = pathFile
	}
	if pathFile, ok := flags["path_file"]; ok && pathFile != "" {
		args.Extra = pathFile
	}

	return args, nil
}
