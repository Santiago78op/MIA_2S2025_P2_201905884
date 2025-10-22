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

// Unmount desmonta una partición por ID y resetea el correlativo del disco a 0
func (s *State) Unmount(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cargar estado si está vacío
	if len(s.DiskLetter) == 0 && len(s.DiskSeq) == 0 {
		s.load()
	}

	// Buscar la entrada para obtener el diskSignature antes de eliminar
	var foundEntry *ports.MountedEntry
	for i := range s.Entries {
		if s.Entries[i].ID == id {
			foundEntry = &s.Entries[i]
			break
		}
	}

	if foundEntry == nil {
		return nil // No existe, no hacer nada
	}

	// Encontrar diskSignature buscando en Path
	// El diskSignature se puede derivar del Path consultando el MBR
	// Por ahora, resetear todos los correlativos a 0 cuando se desmonta
	// (esto es una simplificación, idealmente se debería buscar el diskSignature específico)

	// Buscar todas las entradas del mismo disco (mismo Path base)
	diskPath := foundEntry.Path

	// Contar cuántas particiones quedan montadas del mismo disco
	remainingFromDisk := 0
	for _, entry := range s.Entries {
		if entry.Path == diskPath && entry.ID != id {
			remainingFromDisk++
		}
	}

	// Si no quedan más particiones de este disco, resetear el correlativo
	if remainingFromDisk == 0 {
		// Buscar el diskSignature correspondiente al Path
		for diskSig, letter := range s.DiskLetter {
			// Verificar si alguna entrada usa esta letra
			stillInUse := false
			for _, entry := range s.Entries {
				if entry.ID != id && entry.Path == diskPath {
					stillInUse = true
					break
				}
			}

			if !stillInUse {
				// Resetear el correlativo a 0 para este disco
				s.DiskSeq[diskSig] = 0
			}
			_ = letter // Evitar warning de variable no usada
		}
	}

	// Buscar y eliminar la entrada con el ID especificado
	newEntries := []ports.MountedEntry{}
	for _, entry := range s.Entries {
		if entry.ID != id {
			newEntries = append(newEntries, entry)
		}
	}

	s.Entries = newEntries
	s.save()
	return nil
}

// SetPartitionSeq establece manualmente el correlativo de particiones para un disco
func (s *State) SetPartitionSeq(diskSignature string, seq int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cargar estado si está vacío
	if len(s.DiskLetter) == 0 && len(s.DiskSeq) == 0 {
		s.load()
	}

	s.DiskSeq[diskSignature] = seq
	s.save()
	return nil
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

// Clear limpia completamente el estado (útil para unmountall)
func (s *State) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.DiskLetter = map[string]string{}
	s.DiskSeq = map[string]int{}
	s.Entries = []ports.MountedEntry{}
	s.save()
}
