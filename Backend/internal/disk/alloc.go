package disk

import (
	"os"
	"sort"
)

// segmento [start, end) en bytes
type seg struct{ start, end int64 }

func (s seg) size() int64 { return s.end - s.start }

// buildFreePrimaries devuelve segmentos libres dentro del disco excluyendo primarias/extendidas usadas.
func buildFreePrimaries(mbr *MBR) []seg {
	used := make([]seg, 0, MaxPrimaries)
	for _, p := range mbr.Parts {
		if p.Status == PartStatusUsed && (p.Type == PartTypePrimary || p.Type == PartTypeExtended) {
			used = append(used, seg{p.Start, p.Start + p.Size})
		}
	}
	sort.Slice(used, func(i, j int) bool { return used[i].start < used[j].start })

	free := []seg{}
	cur := int64(binarySizeOf(mbr)) // primer byte libre después del MBR
	limit := mbr.SizeBytes
	for _, u := range used {
		if u.start > cur {
			free = append(free, seg{cur, u.start})
		}
		if u.end > cur {
			cur = u.end
		}
	}
	if cur < limit {
		free = append(free, seg{cur, limit})
	}
	return free
}

// pickByFit elige un segmento por fit (FF/BF/WF)
func pickByFit(free []seg, need int64, fit byte) (seg, bool) {
	candidates := []seg{}
	for _, s := range free {
		if s.size() >= need {
			candidates = append(candidates, s)
		}
	}
	if len(candidates) == 0 {
		return seg{}, false
	}
	switch fit {
	case FitFF:
		return candidates[0], true
	case FitBF:
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].size() < candidates[j].size() })
		return candidates[0], true
	case FitWF:
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].size() > candidates[j].size() })
		return candidates[0], true
	default:
		return candidates[0], true
	}
}

// Helpers para lógicas: recorrer EBRs dentro del rango de extendida y encontrar espacios.
func listEBRs(f *os.File, extStart, extEnd int64) ([]EBR, error) {
	return ListEBRs(f, extStart, extEnd)
}

// binarySizeOf devuelve el tamaño en bytes de structs fijos (reemplaza por binary.Size si prefieres).
func binarySizeOf[T any](t *T) int64 {
	// rápido y explícito: si ajustas campos arriba, actualiza aquí también
	switch any(t).(type) {
	case *MBR:
		// SizeBytes(int64)+CreatedAt(int64)+DiskSig(int32)+Fit(byte)+pad(7)+Parts(4*Partition)
		// Partition: Status byte + Type byte + Fit byte + Start int64 + Size int64 + Name[16]+ pad[8]
		const part = 1 + 1 + 1 + 8 + 8 + 16 + 8
		return 8 + 8 + 4 + 1 + 7 + 4*part
	default:
		return 0
	}
}
