package ports

import "Backend/core/models"

// Acceso a MBR/particiones (crear discos, P/E/L, etc.)
type DiskRepository interface {
	CreateDisk(path string, sizeBytes int64, fit rune) error
	RemoveDisk(path string) error

	ReadMBR(path string) (models.MBR, error)
	WriteMBR(path string, mbr models.MBR) error

	// Crear particiones según tipo/fit (incluye validaciones de espacio y nombres únicos)
	CreatePrimary(path string, part models.Partition) error
	CreateExtended(path string, part models.Partition) error
	CreateLogical(path string, extStart int64, ebr models.EBR) error

	// Helpers
	FindExtended(path string) (ext models.Partition, ok bool, err error)
	NormalizePath(p string) string
	DiskSignature(path string) (string, error) // usar MBR.Signature o derivado
	ValidatePrimaryForMount(path, name string) error
}
