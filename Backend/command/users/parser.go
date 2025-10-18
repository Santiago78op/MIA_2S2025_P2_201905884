package users

import (
	"fmt"
	"strings"
)

type LoginArgs struct {
	ID   string
	User string
	Pass string
}

type MkgrpArgs struct {
	ID   string
	Name string
}

type RmgrpArgs struct {
	ID   string
	Name string
}

type MkusrArgs struct {
	ID   string
	User string
	Pass string
	Grp  string
}

type RmusrArgs struct {
	ID   string
	User string
}

type ChgrpArgs struct {
	ID   string
	User string
	Grp  string
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

func ParseLogin(line string) (LoginArgs, error) {
	_, flags := parseLine(line)
	var args LoginArgs

	id, err := mustString(flags, "id")
	if err != nil {
		return args, err
	}
	args.ID = id

	user, err := mustString(flags, "user")
	if err != nil {
		return args, err
	}
	args.User = user

	pass, err := mustString(flags, "pass")
	if err != nil {
		return args, err
	}
	args.Pass = pass

	return args, nil
}

func ParseMkgrp(line string) (MkgrpArgs, error) {
	_, flags := parseLine(line)
	var args MkgrpArgs

	// -id es opcional, se usará el de la sesión actual si no se especifica
	args.ID = flags["id"]

	name, err := mustString(flags, "name")
	if err != nil {
		return args, err
	}
	args.Name = name

	return args, nil
}

func ParseRmgrp(line string) (RmgrpArgs, error) {
	_, flags := parseLine(line)
	var args RmgrpArgs

	// -id es opcional, se usará el de la sesión actual si no se especifica
	args.ID = flags["id"]

	name, err := mustString(flags, "name")
	if err != nil {
		return args, err
	}
	args.Name = name

	return args, nil
}

func ParseMkusr(line string) (MkusrArgs, error) {
	_, flags := parseLine(line)
	var args MkusrArgs

	// -id es opcional, se usará el de la sesión actual si no se especifica
	args.ID = flags["id"]

	user, err := mustString(flags, "user")
	if err != nil {
		return args, err
	}
	args.User = user

	pass, err := mustString(flags, "pass")
	if err != nil {
		return args, err
	}
	args.Pass = pass

	grp, err := mustString(flags, "grp")
	if err != nil {
		return args, err
	}
	args.Grp = grp

	return args, nil
}

func ParseRmusr(line string) (RmusrArgs, error) {
	_, flags := parseLine(line)
	var args RmusrArgs

	// -id es opcional, se usará el de la sesión actual si no se especifica
	args.ID = flags["id"]

	user, err := mustString(flags, "user")
	if err != nil {
		return args, err
	}
	args.User = user

	return args, nil
}

func ParseChgrp(line string) (ChgrpArgs, error) {
	_, flags := parseLine(line)
	var args ChgrpArgs

	// -id es opcional, se usará el de la sesión actual si no se especifica
	args.ID = flags["id"]

	user, err := mustString(flags, "user")
	if err != nil {
		return args, err
	}
	args.User = user

	grp, err := mustString(flags, "grp")
	if err != nil {
		return args, err
	}
	args.Grp = grp

	return args, nil
}
