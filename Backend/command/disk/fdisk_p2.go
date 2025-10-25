// Package: command/disk
// File   : fdisk_p2.go
// Goal   : Parser y lógica para FDISK ADD/DELETE (Proyecto 2)
//          Complementa el fdisk de creación original.

package disk

import (
	"Backend/storage/diskio"
	"errors"
	"fmt"
	"strings"
)

// ========================= TIPOS P2 =========================

type Unit string

const (
	UnitB Unit = "B"
	UnitK Unit = "K"
	UnitM Unit = "M"
)

type DeleteMode string

const (
	DelFast DeleteMode = "fast"
	DelFull DeleteMode = "full"
)

type FDiskAddArgs struct {
	Path string
	Name string
	Add  int64 // valor en unidades indicadas (puede ser negativo)
	Unit Unit  // default K
}

type FDiskDeleteArgs struct {
	Path string
	Name string
	Mode DeleteMode
}

// ========================= PARSERS P2 =========================

func parseKV(line, key string) string {
	parts := strings.Fields(line)
	for _, p := range parts {
		if strings.HasPrefix(strings.ToLower(p), "-"+strings.ToLower(key)+"=") {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 {
				return kv[1]
			}
		}
	}
	return ""
}

func ParseFDiskAdd(line string) (FDiskAddArgs, error) {
	p := parseKV(line, "path")
	n := parseKV(line, "name")
	a := parseKV(line, "add")
	u := strings.ToUpper(parseKV(line, "unit"))

	if p == "" || n == "" || a == "" {
		return FDiskAddArgs{}, errors.New("fdisk -add requiere -path, -name y -add")
	}

	// Parse signo y valor
	sign := int64(1)
	val := a
	if strings.HasPrefix(val, "+") {
		val = val[1:]
	}
	if strings.HasPrefix(val, "-") {
		sign = -1
		val = val[1:]
	}

	var num int64
	for i := 0; i < len(val); i++ {
		ch := val[i]
		if ch < '0' || ch > '9' {
			return FDiskAddArgs{}, errors.New("-add debe ser entero")
		}
		num = num*10 + int64(ch-'0')
	}

	unit := UnitK
	if u == "B" {
		unit = UnitB
	} else if u == "M" {
		unit = UnitM
	} else if u == "K" || u == "" {
		unit = UnitK
	} else {
		return FDiskAddArgs{}, errors.New("unit inválida (B|K|M)")
	}

	return FDiskAddArgs{Path: p, Name: n, Add: sign * num, Unit: unit}, nil
}

func ParseFDiskDelete(line string) (FDiskDeleteArgs, error) {
	p := parseKV(line, "path")
	n := parseKV(line, "name")
	m := strings.ToLower(parseKV(line, "delete"))

	if p == "" || n == "" || m == "" {
		return FDiskDeleteArgs{}, errors.New("fdisk -delete requiere -path, -name y -delete")
	}

	switch m {
	case "fast", "full":
		return FDiskDeleteArgs{Path: p, Name: n, Mode: DeleteMode(m)}, nil
	default:
		return FDiskDeleteArgs{}, errors.New("-delete debe ser fast o full")
	}
}

// ========================= SERVICIO P2 =========================

// FDiskAdd redimensiona una partición (puede ser positivo o negativo)
func (s *DiskService) FDiskAdd(args FDiskAddArgs) (string, error) {
	// Convertir a bytes
	sizeBytes := int64(args.Add)
	switch args.Unit {
	case UnitB:
		// Ya está en bytes
	case UnitK:
		sizeBytes *= 1024
	case UnitM:
		sizeBytes *= 1024 * 1024
	}

	// Buscar la partición
	refInterface, err := s.repo.FindPartition(args.Path, args.Name)
	if err != nil {
		return "", fmt.Errorf("partición no encontrada: %v", err)
	}

	// Convertir a tipo concreto
	ref, ok := refInterface.(diskio.PartitionRef)
	if !ok {
		return "", errors.New("tipo de referencia inválido")
	}

	// Leer tamaño actual
	_, currentSize, err := s.repo.ReadPartition(args.Path, ref)
	if err != nil {
		return "", err
	}

	newSize := currentSize + sizeBytes
	if newSize <= 0 {
		return "", errors.New("el nuevo tamaño debe ser positivo")
	}

	// Si se reduce, validar contra espacio usado (si disponible)
	if sizeBytes < 0 {
		used, err := s.repo.GetPartitionUsedBytes(args.Path, ref)
		if err == nil && used >= 0 && newSize < used {
			return "", fmt.Errorf("no se puede reducir a %d bytes, se usan %d bytes", newSize, used)
		}
	}

	// Si se expande, verificar espacio disponible
	if sizeBytes > 0 {
		nextRefInterface, err := s.repo.NextPartitionAfter(args.Path, ref)
		if err == nil {
			nextRef, ok := nextRefInterface.(diskio.PartitionRef)
			if ok {
				nextStart, _, err := s.repo.ReadPartition(args.Path, nextRef)
				if err == nil {
					available := nextStart - ref.Start
					if newSize > available {
						return "", fmt.Errorf("espacio insuficiente: disponible=%d, requerido=%d", available, newSize)
					}
				}
			}
		}

		// Si es lógica, verificar límites de la extendida
		if ref.Type == diskio.TypeLogical {
			extStart, extSize, found, err := s.repo.GetExtendedBounds(args.Path)
			if err != nil {
				return "", err
			}
			if !found {
				return "", errors.New("no hay partición extendida")
			}
			if ref.Start+newSize > extStart+extSize {
				return "", errors.New("excede límites de la partición extendida")
			}
		}
	}

	// Ejecutar resize
	if err := s.repo.ResizePartition(args.Path, ref, newSize); err != nil {
		return "", err
	}

	op := "expandida"
	if sizeBytes < 0 {
		op = "reducida"
	}
	return fmt.Sprintf("Partición '%s' %s exitosamente a %d bytes", args.Name, op, newSize), nil
}

// FDiskDelete elimina una partición (fast o full)
func (s *DiskService) FDiskDelete(args FDiskDeleteArgs) (string, error) {
	// Buscar la partición
	ref, err := s.repo.FindPartition(args.Path, args.Name)
	if err != nil {
		return "", fmt.Errorf("partición no encontrada: %v", err)
	}

	// Ejecutar delete según modo
	switch args.Mode {
	case DelFast:
		if err := s.repo.DeletePartitionFast(args.Path, ref); err != nil {
			return "", err
		}
		return fmt.Sprintf("Partición '%s' eliminada (fast)", args.Name), nil
	case DelFull:
		if err := s.repo.DeletePartitionFull(args.Path, ref); err != nil {
			return "", err
		}
		return fmt.Sprintf("Partición '%s' eliminada y limpiada (full)", args.Name), nil
	default:
		return "", errors.New("modo de eliminación inválido")
	}
}
