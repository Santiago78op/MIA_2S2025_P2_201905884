package fs

import "encoding/binary"

// Sizeof usa binary.Size (requiere structs de tamaño fijo)
func Sizeof[T any](sample T) int {
	return binary.Size(sample) // -1 si hay campos dinámicos
}

func AlignUp(v, mult int) int {
	if mult <= 1 {
		return v
	}
	r := v % mult
	if r == 0 {
		return v
	}
	return v + (mult - r)
}
