package disk

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// RW sobre archivo .mia — little endian fijo.
var byteOrder = binary.LittleEndian

func openRW(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR, 0o666)
}

func writeStruct(f *os.File, off int64, v any) error {
	if _, err := f.Seek(off, io.SeekStart); err != nil {
		return err
	}
	return binary.Write(f, byteOrder, v)
}

func readStruct(f *os.File, off int64, v any) error {
	if _, err := f.Seek(off, io.SeekStart); err != nil {
		return err
	}
	return binary.Read(f, byteOrder, v)
}

// ReadStruct lee una estructura desde un archivo (versión exportada)
func ReadStruct(f *os.File, off int64, v any) error {
	return readStruct(f, off, v)
}

func zeroRange(f *os.File, off int64, n int64) error {
	const chunk = 64 * 1024
	buf := make([]byte, chunk)
	written := int64(0)
	for written < n {
		toWrite := chunk
		if n-written < chunk {
			toWrite = int(n - written)
		}
		if _, err := f.WriteAt(buf[:toWrite], off+written); err != nil {
			return err
		}
		written += int64(toWrite)
	}
	return nil
}

func ensureSize(path string, size int64) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o666)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Truncate(size); err != nil {
		return err
	}
	return nil
}

func fileSize(path string) (int64, error) {
	st, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fmtName(src string) ([NameLen]byte, error) {
	var out [NameLen]byte
	if len(src) == 0 {
		return out, fmt.Errorf("nombre vacío")
	}
	if len(src) > NameLen {
		return out, fmt.Errorf("nombre demasiado largo (max %d)", NameLen)
	}
	copy(out[:], []byte(src))
	return out, nil
}

// WriteBytes escribe bytes en una posición específica del archivo
func WriteBytes(f *os.File, off int64, data []byte) error {
	if _, err := f.Seek(off, io.SeekStart); err != nil {
		return err
	}
	_, err := f.Write(data)
	return err
}

// ReadBytes lee bytes desde una posición específica del archivo
func ReadBytes(f *os.File, off int64, size int) ([]byte, error) {
	if _, err := f.Seek(off, io.SeekStart); err != nil {
		return nil, err
	}
	data := make([]byte, size)
	n, err := f.Read(data)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, fmt.Errorf("no se leyeron suficientes bytes: %d de %d", n, size)
	}
	return data, nil
}

// WriteBytesAt escribe bytes usando WriteAt (más seguro para concurrencia)
func WriteBytesAt(f *os.File, off int64, data []byte) error {
	_, err := f.WriteAt(data, off)
	return err
}

// ReadBytesAt lee bytes usando ReadAt (más seguro para concurrencia)
func ReadBytesAt(f *os.File, off int64, size int) ([]byte, error) {
	data := make([]byte, size)
	n, err := f.ReadAt(data, off)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n != size {
		return nil, fmt.Errorf("no se leyeron suficientes bytes: %d de %d", n, size)
	}
	return data, nil
}
