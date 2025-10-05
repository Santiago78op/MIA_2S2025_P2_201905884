package fs

import (
	"fmt"
	"strconv"
)

// ParsePerm convierte "755" → 0755 (uint16)
func ParsePerm(octal string) (uint16, error) {
	if octal == "" {
		return 0, ErrInvalidPerm
	}
	// base 8
	i, err := strconv.ParseUint(octal, 8, 16)
	if err != nil {
		return 0, fmt.Errorf("perm inválido: %w", err)
	}
	return uint16(i), nil
}

// StringPerm convierte 0755 → "755"
func StringPerm(p uint16) string {
	return strconv.FormatUint(uint64(p), 8)
}
