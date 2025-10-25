package models

import (
	"encoding/binary"
	"errors"
	"math"
)

// ========================================
// MODELOS DE JOURNAL SEGÚN ENUNCIADO
// ========================================

// Information representa la estructura i_information del enunciado
// Tamaño: 10 + 32 + 64 + 8 = 114 bytes
type Information struct {
	Op      [10]byte // i_operation: char[10]
	Path    [32]byte // i_path: char[32]
	Content [64]byte // i_content: char[64]
	Date    float64  // i_date: float (timestamp como double)
}

// Journal representa la estructura j_journal del enunciado
// Tamaño total: 600 bytes (4 + 114 + padding)
type Journal struct {
	Count   int32       // j_count: int
	Content Information // j_content: Information
}

// JournalInfo mantiene información sobre el estado del journal
type JournalInfo struct {
	Count int32 // cantidad de entradas usadas
	Max   int32 // cantidad máxima de entradas (50)
}

// Constante de tamaño
const JournalEntrySize = 600 // Tamaño fijo por entrada de journal

// ========================================
// HELPERS PARA JOURNAL
// ========================================

// SetInformation asigna valores a una estructura Information
func (inf *Information) Set(op, path, content string, ts int64) {
	toFixed(inf.Op[:], op)
	toFixed(inf.Path[:], path)
	toFixed(inf.Content[:], content)
	inf.Date = float64(ts)
}

// toFixed copia un string a un array de bytes de tamaño fijo
func toFixed(dst []byte, s string) {
	copy(dst, []byte(s))
	// Rellenar con ceros si es más corto
	if len(s) < len(dst) {
		for i := len(s); i < len(dst); i++ {
			dst[i] = 0
		}
	}
}

// String devuelve el operation como string limpio
func (inf *Information) String() string {
	return trimNull(inf.Op[:])
}

// PathString devuelve el path como string limpio
func (inf *Information) PathString() string {
	return trimNull(inf.Path[:])
}

// ContentString devuelve el content como string limpio
func (inf *Information) ContentString() string {
	return trimNull(inf.Content[:])
}

// trimNull elimina bytes nulos y espacios de un array de bytes
func trimNull(b []byte) string {
	n := 0
	for n < len(b) && b[n] != 0 {
		n++
	}
	return string(b[:n])
}

// ========================================
// TIPOS LEGACY (por compatibilidad)
// ========================================

// JournalOp define los tipos de operaciones (legacy)
type JournalOp uint8

const (
	OpNone JournalOp = iota
	OpMkFile
	OpMkDir
	OpRemove
	OpEdit
	OpRename
	OpCopy
	OpMove
	OpChmod
	OpChown
	OpMkgrp
	OpRmgrp
	OpMkusr
	OpRmusr
)

// String devuelve el nombre de la operación
func (op JournalOp) String() string {
	switch op {
	case OpMkFile:
		return "MKFILE"
	case OpMkDir:
		return "MKDIR"
	case OpRemove:
		return "REMOVE"
	case OpEdit:
		return "EDIT"
	case OpRename:
		return "RENAME"
	case OpCopy:
		return "COPY"
	case OpMove:
		return "MOVE"
	case OpChmod:
		return "CHMOD"
	case OpChown:
		return "CHOWN"
	case OpMkgrp:
		return "MKGRP"
	case OpRmgrp:
		return "RMGRP"
	case OpMkusr:
		return "MKUSR"
	case OpRmusr:
		return "RMUSR"
	default:
		return "NONE"
	}
}

// ========================================
// SERIALIZACIÓN DE JOURNAL
// ========================================

// Marshal convierte un Journal a bytes (600 bytes fijos)
func (j *Journal) Marshal() []byte {
	buf := make([]byte, JournalEntrySize)
	off := 0
	binary.LittleEndian.PutUint32(buf[off:off+4], uint32(j.Count))
	off += 4
	copy(buf[off:off+10], j.Content.Op[:])
	off += 10
	copy(buf[off:off+32], j.Content.Path[:])
	off += 32
	copy(buf[off:off+64], j.Content.Content[:])
	off += 64
	binary.LittleEndian.PutUint64(buf[off:off+8], math.Float64bits(j.Content.Date))
	off += 8
	// El resto queda en cero hasta 600
	return buf
}

// UnmarshalJournal lee un Journal desde bytes
func UnmarshalJournal(b []byte) (Journal, error) {
	if int64(len(b)) < JournalEntrySize {
		return Journal{}, errors.New("buffer journal muy pequeño")
	}
	off := 0
	j := Journal{}
	j.Count = int32(binary.LittleEndian.Uint32(b[off : off+4]))
	off += 4
	copy(j.Content.Op[:], b[off:off+10])
	off += 10
	copy(j.Content.Path[:], b[off:off+32])
	off += 32
	copy(j.Content.Content[:], b[off:off+64])
	off += 64
	j.Content.Date = math.Float64frombits(binary.LittleEndian.Uint64(b[off : off+8]))
	off += 8
	return j, nil
}
