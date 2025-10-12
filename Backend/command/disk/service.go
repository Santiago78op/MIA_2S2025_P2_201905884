package disk

import (
	"fmt"
	// Define tus puertos/Interfaces reales en un paquete común (ej. core/ports)
)

type DiskRepository interface {
	CreateDisk(path string, sizeBytes int64, fit rune) error
	RemoveDisk(path string) error
	FDiskPrimary(path string, args FDiskArgs) error
	FDiskExtended(path string, args FDiskArgs) error
	FDiskLogical(path string, args FDiskArgs) error
	DiskSignature(path string) (string, error) // firma única MBR o hash del path
	ValidatePrimaryForMount(path, name string) error
}

type MountStore interface {
	NextID(carnet2, diskSig string) (string, error) // ej. "841A"
	SetMounted(id, path, name string) error
	List() []MountedEntry
}

type MountedEntry struct {
	ID   string
	Path string
	Name string
}

type DiskService struct {
	repo        DiskRepository
	mounts      MountStore
	carnetLast2 string
}

func NewDiskService(repo DiskRepository, mounts MountStore, carnetLastTwo string) *DiskService {
	return &DiskService{repo: repo, mounts: mounts, carnetLast2: carnetLastTwo}
}

func (s *DiskService) MkDisk(a MkDiskArgs) (string, error) {
	// convertir size+unit a bytes
	mult := int64(1)
	switch a.Unit {
	case 'K':
		mult = 1024
	case 'M':
		mult = 1024 * 1024
	default:
		mult = 1024 * 1024
	}
	bytes := a.Size * mult
	if bytes <= 0 {
		return "", fmt.Errorf("tamaño inválido")
	}
	if err := s.repo.CreateDisk(a.Path, bytes, a.Fit); err != nil {
		return "", err
	}
	return "Disco creado", nil
}

func (s *DiskService) RmDisk(a RmDiskArgs) (string, error) {
	if err := s.repo.RemoveDisk(a.Path); err != nil {
		return "", err
	}
	return "Disco eliminado", nil
}

func (s *DiskService) FDisk(a FDiskArgs) (string, error) {
	switch a.Type {
	case 'P':
		return "Partición primaria creada", s.repo.FDiskPrimary(a.Path, a)
	case 'E':
		return "Partición extendida creada", s.repo.FDiskExtended(a.Path, a)
	case 'L':
		return "Partición lógica creada", s.repo.FDiskLogical(a.Path, a)
	default:
		return "", fmt.Errorf("type inválido")
	}
}

func (s *DiskService) Mount(a MountArgs) (string, error) {
	// Validación del PDF: sólo primarias pueden montarse en P1 (tu repo valida nombre/tipo)
	if err := s.repo.ValidatePrimaryForMount(a.Path, a.Name); err != nil {
		return "", err
	}
	sig, err := s.repo.DiskSignature(a.Path)
	if err != nil {
		return "", err
	}
	id, err := s.mounts.NextID(s.carnetLast2, sig)
	if err != nil {
		return "", err
	}
	if err := s.mounts.SetMounted(id, a.Path, a.Name); err != nil {
		return "", err
	}
	return id, nil
}

func (s *DiskService) Mounted() []MountedEntry { return s.mounts.List() }
