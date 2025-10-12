package adapters

import (
	"time"

	"Backend/command/fs"
	"Backend/storage/diskio"
)

// FsAdapter adapta FileFsRepository para cumplir con fs.FsRepository
type FsAdapter struct {
	repo *diskio.FileFsRepository
}

func NewFsAdapter(repo *diskio.FileFsRepository) fs.FsRepository {
	return &FsAdapter{repo: repo}
}

func (a *FsAdapter) Mkfs(id string) error {
	return a.repo.Mkfs(id)
}

func (a *FsAdapter) Mkdir(id string, absPath []string, parents bool, now time.Time) error {
	// La implementación actual no usa el parámetro now, pero está preparado para cuando se implemente
	_ = now
	return a.repo.Mkdir(id, absPath, parents)
}

func (a *FsAdapter) Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool, now time.Time) error {
	// La implementación actual no usa el parámetro now, pero está preparado para cuando se implemente
	_ = now
	return a.repo.Mkfile(id, absPath, size, contentHostPath, recursive)
}

func (a *FsAdapter) Cat(id string, files [][]string) (string, error) {
	return a.repo.Cat(id, files)
}
