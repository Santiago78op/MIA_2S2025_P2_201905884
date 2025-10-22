package adapters

import (
	"fmt"

	"Backend/command/disk"
	"Backend/core/ports"
	"Backend/storage/mounts"
)

// MountAdapter adapta mounts.State para cumplir con disk.MountStore
type MountAdapter struct {
	state *mounts.State
}

func NewMountAdapter(state *mounts.State) disk.MountStore {
	return &MountAdapter{state: state}
}

func (a *MountAdapter) NextID(carnet2, diskSig string) (string, error) {
	return a.state.NextID(carnet2, diskSig)
}

func (a *MountAdapter) SetMounted(id, path, name string) error {
	return a.state.SetMounted(id, path, name)
}

func (a *MountAdapter) List() []disk.MountedEntry {
	portEntries := a.state.List()
	diskEntries := make([]disk.MountedEntry, len(portEntries))

	for i, pe := range portEntries {
		diskEntries[i] = disk.MountedEntry{
			ID:   pe.ID,
			Path: pe.Path,
			Name: pe.Name,
		}
	}

	return diskEntries
}

// Unmount desmonta una partición por ID
func (a *MountAdapter) Unmount(id string) error {
	return a.state.Unmount(id)
}

// SetPartitionSeq establece el correlativo de particiones para un disco
func (a *MountAdapter) SetPartitionSeq(diskSignature string, seq int) error {
	return a.state.SetPartitionSeq(diskSignature, seq)
}

// Clear limpia completamente el estado
func (a *MountAdapter) Clear() {
	a.state.Clear()
}

// PortsMountStore adapta mounts.State para cumplir con ports.MountStore
type PortsMountStore struct {
	state *mounts.State
}

func NewPortsMountStore(state *mounts.State) ports.MountStore {
	return &PortsMountStore{state: state}
}

func (p *PortsMountStore) NextID(carnet2, diskSig string) (string, error) {
	return p.state.NextID(carnet2, diskSig)
}

func (p *PortsMountStore) SetMounted(id, path, name string) error {
	return p.state.SetMounted(id, path, name)
}

func (p *PortsMountStore) List() []ports.MountedEntry {
	return p.state.List()
}

func (p *PortsMountStore) Unmount(id string) error {
	return p.state.Unmount(id)
}

func (p *PortsMountStore) SetPartitionSeq(diskSignature string, seq int) error {
	return p.state.SetPartitionSeq(diskSignature, seq)
}

func (p *PortsMountStore) GetMount(id string) (*ports.MountedEntry, error) {
	for _, m := range p.state.List() {
		if m.ID == id {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("mount not found: %s", id)
}

func (p *PortsMountStore) Clear() {
	p.state.Clear()
}
