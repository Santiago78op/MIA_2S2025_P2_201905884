package commands

import (
	"fmt"
	"sync"

	"MIA_2S2025_P2_201905884/internal/disk"
	"MIA_2S2025_P2_201905884/internal/fs"
)

// MountIndex define la API para mapear id <-> refs/handles.
type MountIndex interface {
	Put(id string, ref disk.PartitionRef, h fs.MountHandle)
	GetRef(id string) (disk.PartitionRef, bool)
	GetHandle(id string) (fs.MountHandle, bool)
	GetByID(id string) (disk.PartitionRef, bool) // Alias para GetRef
	Del(id string)
	List() []string
	GenerateID(diskPath string) string // Genera ID según formato P1
	Reset()                            // Limpia todo el índice (para test/calificador)
}

// In-memory implementación thread-safe.
type memoryIndex struct {
	mu         sync.RWMutex
	ref        map[string]disk.PartitionRef
	hand       map[string]fs.MountHandle
	diskLetter map[string]rune // path -> letra (A, B, C...)
	diskSeq    map[string]int  // path -> correlativo (1, 2, 3...)
}

const carnetSuffix = "84" // Últimos 2 dígitos del carnet

func NewMemoryIndex() MountIndex {
	return &memoryIndex{
		ref:        make(map[string]disk.PartitionRef),
		hand:       make(map[string]fs.MountHandle),
		diskLetter: make(map[string]rune),
		diskSeq:    make(map[string]int),
	}
}

func (m *memoryIndex) Put(id string, ref disk.PartitionRef, h fs.MountHandle) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ref[id] = ref
	m.hand[id] = h
}

func (m *memoryIndex) GetRef(id string) (disk.PartitionRef, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.ref[id]
	return r, ok
}

func (m *memoryIndex) GetHandle(id string) (fs.MountHandle, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h, ok := m.hand[id]
	return h, ok
}

func (m *memoryIndex) GetByID(id string) (disk.PartitionRef, bool) {
	// Alias para GetRef
	return m.GetRef(id)
}

func (m *memoryIndex) Del(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.ref, id)
	delete(m.hand, id)
}

func (m *memoryIndex) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.ref))
	for k := range m.ref {
		out = append(out, k)
	}
	return out
}

// GenerateID genera un ID según formato P1: <Carnet><Correlativo><LetraDisco>
// Ejemplo: 841A (primera partición del Disco1)
//          842A (segunda partición del Disco1)
//          841B (primera partición del Disco3)
func (m *memoryIndex) GenerateID(diskPath string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Asignar letra al disco si es la primera vez que se ve
	if _, ok := m.diskLetter[diskPath]; !ok {
		// Asignar siguiente letra disponible (A, B, C...)
		nextLetter := rune('A' + len(m.diskLetter))
		m.diskLetter[diskPath] = nextLetter
	}

	// Incrementar correlativo para este disco
	m.diskSeq[diskPath]++

	// Formato: <84><correlativo><letra>
	return fmt.Sprintf("%s%d%c", carnetSuffix, m.diskSeq[diskPath], m.diskLetter[diskPath])
}

// Reset limpia completamente el índice (para garantizar IDs predecibles en cada corrida)
func (m *memoryIndex) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ref = make(map[string]disk.PartitionRef)
	m.hand = make(map[string]fs.MountHandle)
	m.diskLetter = make(map[string]rune)
	m.diskSeq = make(map[string]int)
}
