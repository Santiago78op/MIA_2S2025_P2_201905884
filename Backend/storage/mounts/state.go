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
	DiskLetter map[string]string    `json:"disk_letter"` // firma -> letra
	DiskSeq    map[string]int       `json:"disk_seq"`    // firma -> correlativo
	Entries    []ports.MountedEntry `json:"entries"`     // lista de montajes
	mu         sync.Mutex           `json:"-"`
}

func NewState() *State {
	return &State{
		DiskLetter: map[string]string{},
		DiskSeq:    map[string]int{},
		Entries:    []ports.MountedEntry{},
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

	// Asignar letra si es la primera vez que vemos este disco
	letter, ok := s.DiskLetter[diskSignature]
	if !ok {
		letter = nextLetter(s.DiskLetter)
		s.DiskLetter[diskSignature] = letter
		s.DiskSeq[diskSignature] = 0 // Inicializar contador para este disco
	}

	// Incrementar el contador de particiones para este disco específico
	s.DiskSeq[diskSignature]++
	seq := s.DiskSeq[diskSignature]

	id := carnet2 + itoa(seq) + letter // "84" + "1" + "A" => "841A"
	s.save()
	return id, nil
}

func (s *State) SetMounted(id, path, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Entries = append(s.Entries, ports.MountedEntry{ID: id, Path: path, Name: name})
	s.save()
	return nil
}

func (s *State) List() []ports.MountedEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cargar estado si está vacío
	if len(s.Entries) == 0 && len(s.DiskLetter) == 0 {
		s.load()
	}

	out := make([]ports.MountedEntry, len(s.Entries))
	copy(out, s.Entries)
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
