package journal

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Formato sidecar (.jrnl) por partición (circular buffer):
//
// Header (fixed, 32 bytes):
//  0  int32  magic  = 0x4A524E4C ("JRNL")
//  4  int32  version= 1
//  8  int32  cap    = capacidad (número de entries)
// 12  int32  count  = cuántos válidos (<= cap)
// 16  int32  head   = índice lógico más antiguo (0..cap-1)
// 20  int32  tail   = siguiente posición de escritura (0..cap-1)
// 24  int64  created_unix
// Entries [cap] contiguas, cada una de SizeEntryDisk bytes.

const (
	fileMagic   = 0x4A524E4C
	fileVersion = 1
	headerSize  = 32
)

type header struct {
	Magic   int32
	Version int32
	Cap     int32
	Count   int32
	Head    int32
	Tail    int32
	Created int64
}

type SidecarStore struct {
	baseDir string // dir raíz p/ archivos .jrnl
	cap     int    // capacidad (J del enunciado, p.ej. 50)
	mu      sync.Mutex
}

func NewSidecarStore(baseDir string, capacity int) (*SidecarStore, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacidad inválida: %d", capacity)
	}
	if baseDir == "" {
		baseDir = "./journal"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	return &SidecarStore{baseDir: baseDir, cap: capacity}, nil
}

func (s *SidecarStore) pathOf(partID string) string {
	name := fmt.Sprintf("%s.jrnl", partID)
	return filepath.Join(s.baseDir, name)
}

func (s *SidecarStore) ensureFile(path string) (*os.File, header, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o664)
	if err != nil {
		return nil, header{}, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, header{}, err
	}
	if st.Size() == 0 {
		// Inicializa header + espacio para entries
		h := header{
			Magic:   fileMagic,
			Version: fileVersion,
			Cap:     int32(s.cap),
			Count:   0,
			Head:    0,
			Tail:    0,
			Created: time.Now().Unix(),
		}
		if err := writeHeader(f, &h); err != nil {
			f.Close()
			return nil, header{}, err
		}
		// expandir archivo al tamaño total
		total := int64(headerSize + s.cap*SizeEntryDisk())
		if err := f.Truncate(total); err != nil {
			f.Close()
			return nil, header{}, err
		}
		return f, h, nil
	}
	// Leer header existente
	var h header
	if err := readHeader(f, &h); err != nil {
		f.Close()
		return nil, header{}, err
	}
	if h.Magic != fileMagic || h.Version != fileVersion || h.Cap <= 0 {
		f.Close()
		return nil, header{}, ErrCorrupted
	}
	return f, h, nil
}

func writeHeader(f *os.File, h *header) error {
	buf := make([]byte, headerSize)
	binary.LittleEndian.PutUint32(buf[0:], uint32(h.Magic))
	binary.LittleEndian.PutUint32(buf[4:], uint32(h.Version))
	binary.LittleEndian.PutUint32(buf[8:], uint32(h.Cap))
	binary.LittleEndian.PutUint32(buf[12:], uint32(h.Count))
	binary.LittleEndian.PutUint32(buf[16:], uint32(h.Head))
	binary.LittleEndian.PutUint32(buf[20:], uint32(h.Tail))
	binary.LittleEndian.PutUint64(buf[24:], uint64(h.Created))
	_, err := f.WriteAt(buf, 0)
	return err
}

func readHeader(f *os.File, h *header) error {
	buf := make([]byte, headerSize)
	if _, err := io.ReadFull(io.NewSectionReader(f, 0, headerSize), buf); err != nil {
		return err
	}
	h.Magic = int32(binary.LittleEndian.Uint32(buf[0:]))
	h.Version = int32(binary.LittleEndian.Uint32(buf[4:]))
	h.Cap = int32(binary.LittleEndian.Uint32(buf[8:]))
	h.Count = int32(binary.LittleEndian.Uint32(buf[12:]))
	h.Head = int32(binary.LittleEndian.Uint32(buf[16:]))
	h.Tail = int32(binary.LittleEndian.Uint32(buf[20:]))
	h.Created = int64(binary.LittleEndian.Uint64(buf[24:]))
	return nil
}

func entryOffset(idx int32, cap int32) int64 {
	// header + idx*entrySize
	return int64(headerSize + int(idx)*SizeEntryDisk())
}

// ---------------- Store impl ----------------

func (s *SidecarStore) Append(ctx context.Context, partID string, e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.pathOf(partID)
	f, h, err := s.ensureFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// escribir en Tail
	d := fromEntry(e)
	off := entryOffset(h.Tail, h.Cap)
	if err := writeEntry(f, off, &d); err != nil {
		return err
	}

	// mover Tail y Count; si se llena, rotar Head
	h.Tail = (h.Tail + 1) % h.Cap
	if h.Count < h.Cap {
		h.Count++
	} else {
		h.Head = (h.Head + 1) % h.Cap
	}
	return writeHeader(f, &h)
}

func (s *SidecarStore) List(ctx context.Context, partID string) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.pathOf(partID)
	f, h, err := s.ensureFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := make([]Entry, 0, h.Count)
	idx := h.Head
	for i := int32(0); i < h.Count; i++ {
		var d EntryDisk
		off := entryOffset(idx, h.Cap)
		if err := readEntry(f, off, &d); err != nil {
			return nil, err
		}
		out = append(out, toEntry(d))
		idx = (idx + 1) % h.Cap
	}
	return out, nil
}

func (s *SidecarStore) Replay(ctx context.Context, partID string, apply func(Entry) error) error {
	entries, err := s.List(ctx, partID)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := apply(e); err != nil {
			return err
		}
	}
	return nil
}

func (s *SidecarStore) ClearAll(ctx context.Context, partID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.pathOf(partID)
	f, h, err := s.ensureFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h.Count, h.Head, h.Tail = 0, 0, 0
	// opcional: zero entries (no necesario si basta con marcar vacío)
	// aquí solo reescribimos header
	return writeHeader(f, &h)
}

// ---------------- IO entries ----------------

func writeEntry(f *os.File, off int64, d *EntryDisk) error {
	buf := make([]byte, SizeEntryDisk())
	// unix sec
	binary.LittleEndian.PutUint64(buf[0:], uint64(d.UnixSec))
	// op
	copy(buf[8:8+opMax], d.Op[:])
	// path
	copy(buf[8+opMax:8+opMax+pathMax], d.Path[:])
	// datalen + pad
	binary.LittleEndian.PutUint32(buf[8+opMax+pathMax:], d.DataLen)
	// padding (4 bytes) ya es cero
	// data
	copy(buf[8+opMax+pathMax+8:], d.Data[:])
	_, err := f.WriteAt(buf, off)
	return err
}

func readEntry(f *os.File, off int64, d *EntryDisk) error {
	buf := make([]byte, SizeEntryDisk())
	if _, err := io.ReadFull(io.NewSectionReader(f, off, int64(len(buf))), buf); err != nil {
		return err
	}
	d.UnixSec = int64(binary.LittleEndian.Uint64(buf[0:]))
	copy(d.Op[:], buf[8:8+opMax])
	copy(d.Path[:], buf[8+opMax:8+opMax+pathMax])
	d.DataLen = binary.LittleEndian.Uint32(buf[8+opMax+pathMax:])
	copy(d.Data[:], buf[8+opMax+pathMax+8:])
	return nil
}
