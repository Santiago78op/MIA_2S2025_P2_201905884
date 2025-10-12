package adapters

import (
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
