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
	GenerateID() string // Genera próximo ID secuencial
}

// In-memory implementación thread-safe.
type memoryIndex struct {
	mu      sync.RWMutex
	ref     map[string]disk.PartitionRef
	hand    map[string]fs.MountHandle
	counter int // Contador secuencial para IDs
}

func NewMemoryIndex() MountIndex {
	return &memoryIndex{
		ref:     make(map[string]disk.PartitionRef),
		hand:    make(map[string]fs.MountHandle),
		counter: 0,
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

// GenerateID genera un ID secuencial tipo vd84, vd841, vd842, etc.
// Formato: vd84 + sufijo numérico (solo si counter > 0)
func (m *memoryIndex) GenerateID() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var id string
	if m.counter == 0 {
		id = "vd84"
	} else {
		id = fmt.Sprintf("vd84%d", m.counter)
	}

	m.counter++
	return id
}
