package journal

import (
	"encoding/binary"
	"time"
)

// Layout binario fijo (para binary.Read/Write sin sorpresas).
// Ajusta tamaños si tu enunciado define otros límites.
const (
	opMax   = 16
	pathMax = 256
	dataMax = 512 // payload máx por entrada (puedes subirlo si quieres)
)

var byteOrder = binary.LittleEndian

// EntryDisk — EXACTAMENTE fijo en tamaño.
type EntryDisk struct {
	UnixSec int64         // Timestamp en segundos
	Op      [opMax]byte   // op ascii, null-terminated
	Path    [pathMax]byte // path ascii/utf8, null-terminated
	DataLen uint32        // bytes válidos en Data[0:DataLen]
	_       [4]byte       // padding (alinea a 8 bytes)
	Data    [dataMax]byte // contenido
}

// SizeEntryDisk es útil si necesitas reservar espacio.
func SizeEntryDisk() int {
	return 8 + opMax + pathMax + 4 + 4 + dataMax // = 8 + 16 + 256 + 4 + 4 + 512 = 800
}

// Helpers de mapeo
func fromEntry(e Entry) EntryDisk {
	var d EntryDisk
	d.UnixSec = e.Timestamp.Unix()
	copy(d.Op[:], trimTo(e.Op, opMax))
	copy(d.Path[:], trimTo(e.Path, pathMax))
	n := len(e.Content)
	if n > dataMax {
		n = dataMax
	}
	d.DataLen = uint32(n)
	copy(d.Data[:], e.Content[:n])
	return d
}

func toEntry(d EntryDisk) Entry {
	n := int(d.DataLen)
	if n > dataMax {
		n = dataMax
	}
	return Entry{
		Op:        cstring(d.Op[:]),
		Path:      cstring(d.Path[:]),
		Content:   append([]byte(nil), d.Data[:n]...),
		Timestamp: time.Unix(d.UnixSec, 0).UTC(),
	}
}

func cstring(b []byte) string {
	i := 0
	for i < len(b) && b[i] != 0 {
		i++
	}
	return string(b[:i])
}

func trimTo(s string, max int) []byte {
	b := []byte(s)
	if len(b) > max {
		b = b[:max]
	}
	return b
}
