package fs

import "sync"

// MetaState: guarda metadatos por MountID (p.ej. si es 2fs o 3fs)
type MetaState struct {
	mu   sync.RWMutex
	data map[string]Meta
}

type Meta struct {
	FSKind   string // "2fs"|"3fs"
	BlockSz  int
	InodeSz  int
	JournalN int // entradas reservadas
}

func NewMetaState() *MetaState {
	return &MetaState{data: map[string]Meta{}}
}
func (s *MetaState) Set(id string, m Meta) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = m
}
func (s *MetaState) Get(id string) (Meta, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.data[id]
	return m, ok
}
func (s *MetaState) Del(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, id)
}
