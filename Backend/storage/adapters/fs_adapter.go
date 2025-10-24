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

func (a *FsAdapter) Mkfs(id string, formatType string) error {
	return a.repo.Mkfs(id, formatType)
}

func (a *FsAdapter) MkfsWithType(id string, fsType int32, full bool) error {
	return a.repo.MkfsWithType(id, fsType, full)
}

func (a *FsAdapter) Mkdir(id string, absPath []string, parents bool, uid int, gid int, now time.Time) error {
	// La implementación actual no usa el parámetro now, pero está preparado para cuando se implemente
	_ = now
	return a.repo.Mkdir(id, absPath, parents, uid, gid)
}

func (a *FsAdapter) Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool, uid int, gid int, now time.Time) error {
	// La implementación actual no usa el parámetro now, pero está preparado para cuando se implemente
	_ = now
	return a.repo.Mkfile(id, absPath, size, contentHostPath, recursive, uid, gid)
}

func (a *FsAdapter) Cat(id string, files [][]string, uid int, gid int) (string, error) {
	return a.repo.Cat(id, files, uid, gid)
}

// Comandos avanzados P2
func (a *FsAdapter) Remove(id string, absPath []string, uid int, gid int) error {
	return a.repo.Remove(id, absPath, uid, gid)
}

func (a *FsAdapter) Edit(id string, absPath []string, content []byte, uid int, gid int) error {
	return a.repo.Edit(id, absPath, content, uid, gid)
}

func (a *FsAdapter) Rename(id string, absPath []string, newName string, uid int, gid int) error {
	return a.repo.Rename(id, absPath, newName, uid, gid)
}

func (a *FsAdapter) Copy(id string, srcPath []string, destPath []string, uid int, gid int) (copied int, skipped int, err error) {
	return a.repo.Copy(id, srcPath, destPath, uid, gid)
}

func (a *FsAdapter) Move(id string, srcPath []string, destPath []string, uid int, gid int) error {
	return a.repo.Move(id, srcPath, destPath, uid, gid)
}

func (a *FsAdapter) Find(id string, startPath []string, pattern string, uid int, gid int) ([]string, error) {
	return a.repo.Find(id, startPath, pattern, uid, gid)
}

func (a *FsAdapter) Chmod(id string, absPath []string, ugo [3]byte, recursive bool, uid int, gid int) error {
	return a.repo.Chmod(id, absPath, ugo, recursive, uid, gid)
}

func (a *FsAdapter) Chown(id string, absPath []string, user string, recursive bool, uid int, gid int) error {
	return a.repo.Chown(id, absPath, user, recursive, uid, gid)
}

func (a *FsAdapter) WipeDataAreas(id string) error {
	return a.repo.WipeDataAreas(id)
}

func (a *FsAdapter) Recovery(id string) error {
	return a.repo.Recovery(id)
}
