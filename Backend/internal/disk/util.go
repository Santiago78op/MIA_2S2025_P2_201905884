package disk

import "strings"

func unitToBytes(size int64, unit string) int64 {
	u := strings.ToLower(unit)
	switch u {
	case "", "b":
		return size
	case "k", "kb":
		return size * 1024
	case "m", "mb":
		return size * 1024 * 1024
	default:
		return size // caller valida
	}
}
