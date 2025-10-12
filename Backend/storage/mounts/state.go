package mounts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"Backend/core/ports"
)

const statePath = "/tmp/mia_mount_state.json"

type State struct {
	DiskLetter map[string]string `json:"disk_letter"` // firma -> letra
	DiskSeq    map[string]int    `json:"disk_seq"`    // firma -> correlativo
	mu         sync.Mutex
	entries    []ports.MountedEntry
}

func NewState() *State {
	return &State{
		DiskLetter: map[string]string{},
		DiskSeq:    map[string]int{},
		entries:    []ports.MountedEntry{},
	}
}

func (s *State) load() {
	b, err := os.ReadFile(statePath)
	if err == nil {
		_ = json.Unmarshal(b, s)
	}
}
func (s *State) save() {
	_ = os.MkdirAll(filepath.Dir(statePath), 0o755)
	b, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(statePath, b, 0o644)
}

func nextLetter(used map[string]string) string {
	taken := map[string]bool{}
	for _, v := range used {
		taken[v] = true
	}
	for c := 'A'; c <= 'Z'; c++ {
		if !taken[string(c)] {
			return string(c)
		}
	}
	return "Z"
}

func (s *State) NextID(carnet2, diskSignature string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.DiskLetter) == 0 && len(s.DiskSeq) == 0 {
		s.load()
	}
	letter, ok := s.DiskLetter[diskSignature]
	if !ok {
		letter = nextLetter(s.DiskLetter)
		s.DiskLetter[diskSignature] = letter
	}
	seq := s.DiskSeq[diskSignature] + 1
	s.DiskSeq[diskSignature] = seq

	id := carnet2 + itoa(seq) + letter // "84" + "1" + "A" => "841A"
	s.save()
	return id, nil
}

func (s *State) SetMounted(id, path, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, ports.MountedEntry{ID: id, Path: path, Name: name})
	return nil
}

func (s *State) List() []ports.MountedEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ports.MountedEntry, len(s.entries))
	copy(out, s.entries)
	return out
}

// itoa minimal para números pequeños
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
