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
	// NOTA: Start se calculará dentro de CreateLogical
	ebr := models.EBR{
		Status: models.PartStatusFree, // Se establecerá a Used dentro de CreateLogical
		Fit:    byte(args.Fit),
		Start:  0, // Se calculará dentro de CreateLogical usando el algoritmo de fit
		Size:   args.Size,
		Next:   -1,
	}
	copy(ebr.Name[:], args.Name)

	return a.repo.CreateLogical(path, extPart.Start, ebr)
}

// ========================= MÉTODOS P2 =========================

func (a *DiskAdapter) FindPartition(path string, name string) (interface{}, error) {
	return a.repo.FindPartition(path, name)
}

func (a *DiskAdapter) NextPartitionAfter(path string, ref interface{}) (interface{}, error) {
	return a.repo.NextPartitionAfter(path, ref.(diskio.PartitionRef))
}

func (a *DiskAdapter) ReadPartition(path string, ref interface{}) (int64, int64, error) {
	return a.repo.ReadPartition(path, ref.(diskio.PartitionRef))
}

func (a *DiskAdapter) ResizePartition(path string, ref interface{}, newSize int64) error {
	return a.repo.ResizePartition(path, ref.(diskio.PartitionRef), newSize)
}

func (a *DiskAdapter) DeletePartitionFast(path string, ref interface{}) error {
	return a.repo.DeletePartitionFast(path, ref.(diskio.PartitionRef))
}

func (a *DiskAdapter) DeletePartitionFull(path string, ref interface{}) error {
	return a.repo.DeletePartitionFull(path, ref.(diskio.PartitionRef))
}

func (a *DiskAdapter) GetExtendedBounds(path string) (int64, int64, bool, error) {
	return a.repo.GetExtendedBounds(path)
}

func (a *DiskAdapter) GetPartitionUsedBytes(path string, ref interface{}) (int64, error) {
	return a.repo.GetPartitionUsedBytes(path, ref.(diskio.PartitionRef))
}

func (a *DiskAdapter) CreatePartition(path string, name string, sizeBytes int64, ptype interface{}, fit byte) (string, error) {
	return a.repo.CreatePartition(path, name, sizeBytes, ptype.(diskio.PartType), fit)
}

