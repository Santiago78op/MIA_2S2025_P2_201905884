package adapters

import (
	"Backend/command/disk"
	"Backend/core/models"
	"Backend/storage/diskio"
)

// DiskAdapter adapta FileDiskRepository para cumplir con disk.DiskRepository
type DiskAdapter struct {
	repo *diskio.FileDiskRepository
}

func NewDiskAdapter(repo *diskio.FileDiskRepository) disk.DiskRepository {
	return &DiskAdapter{repo: repo}
}

func (a *DiskAdapter) CreateDisk(path string, sizeBytes int64, fit rune) error {
	return a.repo.CreateDisk(path, sizeBytes, fit)
}

func (a *DiskAdapter) RemoveDisk(path string) error {
	return a.repo.RemoveDisk(path)
}

func (a *DiskAdapter) DiskSignature(path string) (string, error) {
	return a.repo.DiskSignature(path)
}

func (a *DiskAdapter) ValidatePrimaryForMount(path, name string) error {
	return a.repo.ValidatePrimaryForMount(path, name)
}

func (a *DiskAdapter) FDiskPrimary(path string, args disk.FDiskArgs) error {
	// Convertir disk.FDiskArgs a models.Partition
	part := models.Partition{
		Size: args.Size,
		Fit:  byte(args.Fit),
	}
	copy(part.Name[:], args.Name)
	return a.repo.CreatePrimary(path, part)
}

func (a *DiskAdapter) FDiskExtended(path string, args disk.FDiskArgs) error {
	// Convertir disk.FDiskArgs a models.Partition
	part := models.Partition{
		Size: args.Size,
		Fit:  byte(args.Fit),
	}
	copy(part.Name[:], args.Name)
	return a.repo.CreateExtended(path, part)
}

func (a *DiskAdapter) FDiskLogical(path string, args disk.FDiskArgs) error {
	// Primero encontrar la partición extendida
	extPart, found, err := a.repo.FindExtended(path)
	if err != nil {
		return err
	}
	if !found {
		return models.ErrNoExtendedPartition
	}

	// Crear EBR para la partición lógica
	ebr := models.EBR{
		Status: models.PartStatusUsed,
		Fit:    byte(args.Fit),
		Start:  extPart.Start, // Se ajustará dentro de CreateLogical
		Size:   args.Size,
		Next:   -1,
	}
	copy(ebr.Name[:], args.Name)

	return a.repo.CreateLogical(path, extPart.Start, ebr)
}
