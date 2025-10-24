package adapters

import (
	"Backend/core/models"
	"Backend/core/ports"
	"Backend/storage/diskio"
)

// PortsFsAdapter adapta FileFsRepository para cumplir con ports.FsRepository
type PortsFsAdapter struct {
	repo *diskio.FileFsRepository
}

func NewPortsFsAdapter(repo *diskio.FileFsRepository) ports.FsRepository {
	return &PortsFsAdapter{repo: repo}
}

// ====== MKFS ======

func (a *PortsFsAdapter) Mkfs(id string) error {
	return a.repo.Mkfs(id, "full")
}

func (a *PortsFsAdapter) MkfsWithType(id string, fsType int32, full bool) error {
	return a.repo.MkfsWithType(id, fsType, full)
}

// ====== CRUD de directorios/archivos ======

func (a *PortsFsAdapter) Mkdir(id string, absPath []string, parents bool) error {
	// FileFsRepository.Mkdir requiere uid y gid (que están en la sesión)
	// Por ahora usamos uid=1, gid=1 (root) como fallback
	return a.repo.Mkdir(id, absPath, parents, 1, 1)
}

func (a *PortsFsAdapter) Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool) error {
	return a.repo.Mkfile(id, absPath, size, contentHostPath, recursive, 1, 1)
}

func (a *PortsFsAdapter) Cat(id string, files [][]string) (string, error) {
	return a.repo.Cat(id, files, 1, 1)
}

// ====== Nuevas operaciones P2 ======

func (a *PortsFsAdapter) Remove(id string, path []string) error {
	return a.repo.Remove(id, path)
}

func (a *PortsFsAdapter) Edit(id string, path []string, contentHostPath string) error {
	return a.repo.Edit(id, path, []byte(contentHostPath), 1, 1)
}

func (a *PortsFsAdapter) Rename(id string, path []string, newName string) error {
	return a.repo.Rename(id, path, newName, 1, 1) // uid=1 (root), gid=1
}

func (a *PortsFsAdapter) Copy(id string, srcPath []string, destPath []string) error {
	_, _, err := a.repo.Copy(id, srcPath, destPath, 1, 1)
	return err
}

func (a *PortsFsAdapter) Move(id string, srcPath []string, destPath []string) error {
	return a.repo.Move(id, srcPath, destPath, 1, 1)
}

func (a *PortsFsAdapter) Find(id string, basePath []string, pattern string) ([]string, error) {
	return a.repo.Find(id, basePath, pattern, 1, 1)
}

// ====== Permisos y propietarios ======

func (a *PortsFsAdapter) Chmod(id string, path []string, ugo [3]byte, recursive bool) error {
	return a.repo.Chmod(id, path, ugo, recursive, 1, 1)
}

func (a *PortsFsAdapter) Chown(id string, path []string, user string, recursive bool) error {
	return a.repo.Chown(id, path, user, recursive, 1, 1)
}

// ====== Lecturas/escrituras de estructuras ======

func (a *PortsFsAdapter) ReadSuper(id string) (models.SuperBlock, ports.Region, error) {
	sb, region, err := a.repo.ReadSuper(id)
	return sb, ports.Region{Start: region.Start, Size: region.Size}, err
}

func (a *PortsFsAdapter) WriteSuper(id string, sb models.SuperBlock) error {
	return a.repo.WriteSuper(id, sb)
}

func (a *PortsFsAdapter) ReadBitmapInode(id string) ([]byte, error) {
	return a.repo.ReadBitmapInode(id)
}

func (a *PortsFsAdapter) ReadBitmapBlock(id string) ([]byte, error) {
	return a.repo.ReadBitmapBlock(id)
}

func (a *PortsFsAdapter) WriteBitmapInode(id string, data []byte) error {
	return a.repo.WriteBitmapInode(id, data)
}

func (a *PortsFsAdapter) WriteBitmapBlock(id string, data []byte) error {
	return a.repo.WriteBitmapBlock(id, data)
}

func (a *PortsFsAdapter) ReadInode(id string, index int) (models.Inode, error) {
	return a.repo.ReadInode(id, index)
}

func (a *PortsFsAdapter) WriteInode(id string, index int, ino models.Inode) error {
	return a.repo.WriteInode(id, index, ino)
}

func (a *PortsFsAdapter) ReadBlock(id string, index int) ([]byte, error) {
	return a.repo.ReadBlock(id, index)
}

func (a *PortsFsAdapter) WriteBlock(id string, index int, data []byte) error {
	return a.repo.WriteBlock(id, index, data)
}

// ====== Journal (solo para EXT3) ======

func (a *PortsFsAdapter) JournalAppend(id string, entry models.Journal) error {
	op := string(entry.Content.Op[:])
	path := string(entry.Content.Path[:])
	content := string(entry.Content.Content[:])
	ts := int64(entry.Content.Date)
	return a.repo.JournalAppendPublic(id, op, path, content, ts)
}

func (a *PortsFsAdapter) JournalList(id string) ([]models.Journal, error) {
	return a.repo.JournalListPublic(id)
}

func (a *PortsFsAdapter) JournalClear(id string) error {
	return a.repo.JournalClearPublic(id)
}

// ====== Recovery y Loss ======

func (a *PortsFsAdapter) WipeDataAreas(id string) error {
	return a.repo.WipeDataAreas(id)
}

func (a *PortsFsAdapter) Recovery(id string) error {
	return a.repo.Recovery(id)
}

// ====== Viewer helpers ======

func (a *PortsFsAdapter) ListDirectory(id string, path []string) ([]ports.DirectoryEntry, error) {
	return a.repo.ListDirectory(id, path)
}
