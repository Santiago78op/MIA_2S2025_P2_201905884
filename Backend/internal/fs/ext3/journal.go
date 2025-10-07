package ext3

import (
	"encoding/binary"
	"time"
)

const (
	JournalEntrySize  = 64
	JournalEntryCount = 50 // FIJO según enunciado
)

// JournalEntry representa una entrada del journal EXT3
type JournalEntry struct {
	Operation   [16]byte // mkdir, mkfile, edit, remove, rename, copy, move, chown, chmod
	Path        [24]byte // ruta del archivo/directorio
	Content     [8]byte  // información adicional (ej: permisos, owner)
	Timestamp   int64    // timestamp de la operación
	UserID      int32    // ID del usuario que ejecutó
	GroupID     int32    // ID del grupo
	Permissions uint16   // permisos (chmod)
	_           [2]byte  // padding para alinear a 64 bytes
}

// Journal es un array fijo de 50 entradas
type Journal struct {
	Entries [JournalEntryCount]JournalEntry
	Current int32 // Índice circular actual
}

// NewJournal crea un journal vacío
func NewJournal() *Journal {
	return &Journal{
		Entries: [JournalEntryCount]JournalEntry{},
		Current: 0,
	}
}

// Append agrega una entrada al journal (circular buffer)
func (j *Journal) Append(entry JournalEntry) {
	// Escribir en posición actual
	j.Entries[j.Current] = entry

	// Avanzar índice circular
	j.Current = (j.Current + 1) % JournalEntryCount
}

// GetAll retorna todas las entradas no vacías ordenadas
func (j *Journal) GetAll() []JournalEntry {
	var result []JournalEntry

	// Leer desde Current hasta el final, luego desde el inicio hasta Current
	for i := int32(0); i < JournalEntryCount; i++ {
		idx := (j.Current + i) % JournalEntryCount
		entry := j.Entries[idx]

		// Solo incluir entradas con operación válida
		if entry.Timestamp > 0 {
			result = append(result, entry)
		}
	}

	return result
}

// Clear limpia el journal (para recovery)
func (j *Journal) Clear() {
	j.Entries = [JournalEntryCount]JournalEntry{}
	j.Current = 0
}

// Serialize convierte el journal completo a bytes
func (j *Journal) Serialize() []byte {
	// 50 entradas * 64 bytes = 3200 bytes
	buf := make([]byte, JournalEntryCount*JournalEntrySize)

	for i := 0; i < JournalEntryCount; i++ {
		offset := i * JournalEntrySize
		entryBytes := j.Entries[i].Serialize()
		copy(buf[offset:], entryBytes)
	}

	return buf
}

// DeserializeJournal lee el journal desde bytes
func DeserializeJournal(data []byte) *Journal {
	j := NewJournal()

	for i := 0; i < JournalEntryCount; i++ {
		offset := i * JournalEntrySize
		j.Entries[i] = DeserializeJournalEntry(data[offset : offset+JournalEntrySize])
	}

	// Encontrar el índice actual (última entrada escrita + 1)
	for i := int32(0); i < JournalEntryCount; i++ {
		if j.Entries[i].Timestamp == 0 {
			j.Current = i
			break
		}
	}

	return j
}

// Serialize convierte una entrada a bytes
func (je *JournalEntry) Serialize() []byte {
	buf := make([]byte, JournalEntrySize)

	copy(buf[0:16], je.Operation[:])
	copy(buf[16:40], je.Path[:])
	copy(buf[40:48], je.Content[:])
	binary.LittleEndian.PutUint64(buf[48:56], uint64(je.Timestamp))
	binary.LittleEndian.PutUint32(buf[56:60], uint32(je.UserID))
	binary.LittleEndian.PutUint32(buf[60:64], uint32(je.GroupID))

	return buf
}

// DeserializeJournalEntry lee una entrada desde bytes
func DeserializeJournalEntry(data []byte) JournalEntry {
	var je JournalEntry

	copy(je.Operation[:], data[0:16])
	copy(je.Path[:], data[16:40])
	copy(je.Content[:], data[40:48])
	je.Timestamp = int64(binary.LittleEndian.Uint64(data[48:56]))
	je.UserID = int32(binary.LittleEndian.Uint32(data[56:60]))
	je.GroupID = int32(binary.LittleEndian.Uint32(data[60:64]))

	return je
}

// Helper para crear entradas de journal
func NewJournalEntry(op, path, content string, userID, groupID int32, perms uint16) JournalEntry {
	var entry JournalEntry

	copy(entry.Operation[:], op)
	copy(entry.Path[:], path)
	copy(entry.Content[:], content)
	entry.Timestamp = time.Now().Unix()
	entry.UserID = userID
	entry.GroupID = groupID
	entry.Permissions = perms

	return entry
}
