package fs

import (
	"context"
	"time"
)

// FS — interfaz común para EXT2 y EXT3.
type FS interface {
	// Formateo / Montaje
	Mkfs(ctx context.Context, req MkfsRequest) error
	Mount(ctx context.Context, req MountRequest) (MountHandle, error)
	Unmount(ctx context.Context, h MountHandle) error

	// Árbol/archivos (mínimo para P2)
	Tree(ctx context.Context, h MountHandle, path string) (TreeNode, error)
	ReadFile(ctx context.Context, h MountHandle, path string) ([]byte, FileStat, error)
	WriteFile(ctx context.Context, h MountHandle, req WriteFileRequest) error
	Mkdir(ctx context.Context, h MountHandle, req MkdirRequest) error
	Remove(ctx context.Context, h MountHandle, path string) error
	Rename(ctx context.Context, h MountHandle, from, to string) error
	Copy(ctx context.Context, h MountHandle, from, to string) error
	Move(ctx context.Context, h MountHandle, from, to string) error
	Find(ctx context.Context, h MountHandle, req FindRequest) ([]string, error)
	Chown(ctx context.Context, h MountHandle, path, user, group string) error
	Chmod(ctx context.Context, h MountHandle, path string, perm uint16) error

	// EXT3-only (no-op en EXT2)
	Journaling(ctx context.Context, h MountHandle) ([]JournalEntry, error)
	Recovery(ctx context.Context, h MountHandle) error
	Loss(ctx context.Context, h MountHandle) error
}

type MountHandle struct {
	DiskID      string // path del .mia o id de disco
	PartitionID string // nombre/índice lógico
	User        string // opcional
	Group       string // opcional
}

type JournalEntry struct {
	Op        string
	Path      string
	Content   []byte
	Timestamp time.Time
}
