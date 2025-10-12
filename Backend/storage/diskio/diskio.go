package diskio

import (
	"errors"
	"fmt"
	"os"
)

var ErrOutOfBounds = errors.New("operación fuera de los límites de la partición")

// Region define el rango [Start, Start+Size) permitido para lecturas/escrituras.
type Region struct {
	Start int64
	Size  int64
}

func (r Region) End() int64 { return r.Start + r.Size }

func boundsCheck(absOff int64, length int64, r Region) error {
	if length < 0 {
		return fmt.Errorf("%w: longitud negativa", ErrOutOfBounds)
	}
	if absOff < r.Start || absOff+length > r.End() {
		return fmt.Errorf("%w: off=%d len=%d rango=[%d,%d)", ErrOutOfBounds, absOff, length, r.Start, r.End())
	}
	return nil
}

func AbsOffset(rel int64, r Region) int64 { return r.Start + rel }

// SafeWriteAt verifica límites antes de escribir.
func SafeWriteAt(f *os.File, data []byte, absOff int64, r Region) (int, error) {
	if err := boundsCheck(absOff, int64(len(data)), r); err != nil {
		return 0, err
	}
	return f.WriteAt(data, absOff)
}

// SafeReadAt verifica límites antes de leer.
func SafeReadAt(f *os.File, dst []byte, absOff int64, r Region) (int, error) {
	if err := boundsCheck(absOff, int64(len(dst)), r); err != nil {
		return 0, err
	}
	return f.ReadAt(dst, absOff)
}
