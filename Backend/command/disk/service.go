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

	// Nuevos métodos P2 para operaciones avanzadas de particiones
	FindPartition(path string, name string) (interface{}, error) // Retorna PartitionRef
	NextPartitionAfter(path string, ref interface{}) (interface{}, error)
	ReadPartition(path string, ref interface{}) (int64, int64, error)
	ResizePartition(path string, ref interface{}, newSize int64) error
	DeletePartitionFast(path string, ref interface{}) error
	DeletePartitionFull(path string, ref interface{}) error
	GetExtendedBounds(path string) (int64, int64, bool, error)
	GetPartitionUsedBytes(path string, ref interface{}) (int64, error)
	CreatePartition(path string, name string, sizeBytes int64, ptype interface{}, fit byte) (string, error)
}

type MountStore interface {
	NextID(carnet2, diskSig string) (string, error) // ej. "841A"
	SetMounted(id, path, name string) error
	List() []MountedEntry
	Unmount(id string) error                           // NUEVO P2
	SetPartitionSeq(diskSignature string, seq int) error // NUEVO P2
	Clear()                                            // NUEVO P2 - Limpia completamente el estado
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
	// Convertir size+unit a bytes
	mult := int64(1)
	switch a.Unit {
	case 'B':
		mult = 1
	case 'K':
		mult = 1024
	case 'M':
		mult = 1024 * 1024
	default:
		mult = 1024 // default K
	}
	a.Size = a.Size * mult
	if a.Size <= 0 {
		return "", fmt.Errorf("tamaño inválido")
	}

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
	// Validación: verificar que la partición existe
	if err := s.repo.ValidatePrimaryForMount(a.Path, a.Name); err != nil {
		return "", err
	}

	// Verificar si la partición ya está montada
	mounted := s.mounts.List()
	for _, entry := range mounted {
		if entry.Path == a.Path && entry.Name == a.Name {
			return "", fmt.Errorf("la partición ya está montada con ID: %s", entry.ID)
		}
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

// ============================================
// MÉTODOS NUEVOS P2
// ============================================

func (s *DiskService) Unmount(id string) (string, error) {
	// Validar que el ID existe en el mount store
	mounted := s.mounts.List()
	found := false
	for _, entry := range mounted {
		if entry.ID == id {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("la partición con ID '%s' no está montada", id)
	}

	// Desmontar partición (esto también resetea el correlativo si es necesario)
	if err := s.mounts.Unmount(id); err != nil {
		return "", fmt.Errorf("error desmontando partición: %w", err)
	}

	return fmt.Sprintf("Partición %s desmontada correctamente", id), nil
}

func (s *DiskService) UnmountAll() (string, error) {
	mounted := s.mounts.List()

	if len(mounted) == 0 {
		return "No hay particiones montadas", nil
	}

	count := len(mounted)

	// Limpiar completamente el estado (desmonta todas y resetea las letras de disco)
	s.mounts.Clear()

	return fmt.Sprintf("Se desmontaron %d particiones correctamente y se reinició la secuencia de letras", count), nil
}
