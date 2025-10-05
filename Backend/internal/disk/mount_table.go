package disk

import "sync"

type PartitionRef struct {
	DiskPath    string
	PartitionID string // nombre para P/E, para L usa nombre lógico
}

type mountTable struct {
	mu  sync.RWMutex
	set map[string]PartitionRef // key: path|name
}

func newMountTable() *mountTable {
	return &mountTable{set: map[string]PartitionRef{}}
}

func key(path, name string) string { return path + "|" + name }

func (t *mountTable) put(path, name string, ref PartitionRef) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	k := key(path, name)
	if _, ok := t.set[k]; ok {
		return ErrAlreadyMounted
	}
	t.set[k] = ref
	return nil
}

func (t *mountTable) del(path, name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.set, key(path, name))
}

func (t *mountTable) get(path, name string) (PartitionRef, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	ref, ok := t.set[key(path, name)]
	return ref, ok
}

func (t *mountTable) list() []PartitionRef {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]PartitionRef, 0, len(t.set))
	for _, v := range t.set {
		out = append(out, v)
	}
	return out
}
