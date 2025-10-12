package disk

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func ParseMkDisk(flags map[string]string) (MkDiskArgs, error) {
	var a MkDiskArgs
	size, err := parseInt64(flags, "size")
	if err != nil || size <= 0 {
		return a, fmt.Errorf("size inválido")
	}
	unit, err := mustRune(flags, "unit", 'M', "KM")
	if err != nil {
		return a, err
	}
	fit, err := mustRune(flags, "fit", 'F', "FBW") // FF por defecto (F)
	if err != nil {
		return a, err
	}
	path := flags["path"]
	if path == "" {
		return a, fmt.Errorf("path requerido")
	}
	a.Size, a.Unit, a.Fit, a.Path = size, unit, fit, filepath.Clean(path)
	return a, nil
}

func ParseRmDisk(flags map[string]string) (RmDiskArgs, error) {
	p := flags["path"]
	if p == "" {
		return RmDiskArgs{}, fmt.Errorf("path requerido")
	}
	return RmDiskArgs{Path: filepath.Clean(p)}, nil
}

func ParseFDisk(flags map[string]string) (FDiskArgs, error) {
	var a FDiskArgs
	size, err := parseInt64(flags, "size")
	if err != nil || size <= 0 {
		return a, fmt.Errorf("size inválido")
	}
	unit, err := mustRune(flags, "unit", 'K', "KM")
	if err != nil {
		return a, err
	}
	typ, err := mustRune(flags, "type", 'P', "PEL")
	if err != nil {
		return a, err
	}
	fit, err := mustRune(flags, "fit", 'W', "FBW") // WF por defecto
	if err != nil {
		return a, err
	}
	path := flags["path"]
	name := flags["name"]
	if path == "" || name == "" {
		return a, fmt.Errorf("path y name requeridos")
	}
	a.Size, a.Unit, a.Type, a.Fit, a.Path, a.Name = size, unit, typ, fit, filepath.Clean(path), name
	return a, nil
}

func ParseMount(flags map[string]string) (MountArgs, error) {
	path := flags["path"]
	name := flags["name"]
	if path == "" || name == "" {
		return MountArgs{}, fmt.Errorf("path y name requeridos")
	}
	return MountArgs{Path: filepath.Clean(path), Name: name}, nil
}

func parseInt64(flags map[string]string, key string) (int64, error) {
	v := strings.TrimSpace(flags[key])
	if v == "" {
		return 0, fmt.Errorf("falta %s", key)
	}
	return strconv.ParseInt(v, 10, 64)
}
