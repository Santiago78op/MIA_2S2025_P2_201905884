package fs

import (
	"context"
	"fmt"
	"os"
)

// BlockIO abstrae lectura/escritura binaria a “bloques” dentro del .mia.
// Puedes reemplazarlo por tu implementación real (mem-mapped, bufio, etc).
type BlockIO interface {
	ReadAt(ctx context.Context, p []byte, off int64) (int, error)
	WriteAt(ctx context.Context, p []byte, off int64) (int, error)
	Sync(ctx context.Context) error
	Close() error
	Size() (int64, error)
}

// Simple implementación sobre *os.File
type fileBlockIO struct{ f *os.File }

func NewFileBlockIO(path string) (BlockIO, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0o666)
	if err != nil {
		return nil, err
	}
	return &fileBlockIO{f: f}, nil
}

func (io *fileBlockIO) ReadAt(ctx context.Context, p []byte, off int64) (int, error) {
	return io.f.ReadAt(p, off)
}
func (io *fileBlockIO) WriteAt(ctx context.Context, p []byte, off int64) (int, error) {
	return io.f.WriteAt(p, off)
}
func (io *fileBlockIO) Sync(ctx context.Context) error { return io.f.Sync() }
func (io *fileBlockIO) Close() error                   { return io.f.Close() }
func (io *fileBlockIO) Size() (int64, error) {
	st, err := io.f.Stat()
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

// Offset calcula desplazamiento: inicio + index*blockSize
func Offset(base int64, idx int, blockSize int) int64 {
	return base + int64(idx*blockSize)
}

func MustFixedSize(sz int) {
	if sz <= 0 {
		panic(fmt.Errorf("struct size must be > 0, got %d", sz))
	}
}
