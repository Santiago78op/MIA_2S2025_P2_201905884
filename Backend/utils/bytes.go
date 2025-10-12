package utils

import (
	"bytes"
	"encoding/binary"
)

// Little endian fijo en todo el proyecto
var le = binary.LittleEndian

// ToBytes serializa cualquier struct a []byte en little-endian.
func ToBytes(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, le, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FromBytes deserializa desde []byte a la referencia del struct.
func FromBytes(data []byte, dst any) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, le, dst)
}

// PutUint32/64 helpers
func PutUint32(b []byte, v uint32) { le.PutUint32(b, v) }
func PutUint64(b []byte, v uint64) { le.PutUint64(b, v) }
func Uint32(b []byte) uint32       { return le.Uint32(b) }
func Uint64(b []byte) uint64       { return le.Uint64(b) }

// PadOrTrimName16 rellena/recorta a 16 bytes (para nombres de partición/EBR).
func PadOrTrimName16(s string) [16]byte {
	var out [16]byte
	n := len(s)
	if n > 16 {
		copy(out[:], s[:16])
		return out
	}
	copy(out[:], s)
	return out
}

// BytesForBits redondea bits -> bytes hacia arriba.
func BytesForBits(bits int) int {
	return (bits + 7) / 8
}
