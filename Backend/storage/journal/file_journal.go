package journal

import (
	"Backend/core/models"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var (
	ErrJournalFull       = errors.New("journal is full")
	ErrJournalNotFound   = errors.New("journal not found or not EXT3")
	ErrInvalidJournalOfs = errors.New("invalid journal offset")
)

// FileJournal maneja operaciones de lectura/escritura del journal en disco
type FileJournal struct {
	diskPath string
	start    int64 // offset absoluto del inicio del journal
	maxCount int32 // cantidad máxima de entradas
}

// NewFileJournal crea un nuevo manejador de journal
func NewFileJournal(diskPath string, journalStart int64, maxCount int32) *FileJournal {
	return &FileJournal{
		diskPath: diskPath,
		start:    journalStart,
		maxCount: maxCount,
	}
}

// Append agrega una entrada al journal
func (fj *FileJournal) Append(entry models.JournalEntry) error {
	entries, err := fj.List()
	if err != nil {
		return err
	}

	if int32(len(entries)) >= fj.maxCount {
		return ErrJournalFull
	}

	f, err := os.OpenFile(fj.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Calcular offset de la siguiente entrada libre
	entryOffset := fj.start + int64(len(entries))*288 // 288 bytes por entrada

	_, err = f.Seek(entryOffset, io.SeekStart)
	if err != nil {
		return err
	}

	return binary.Write(f, binary.LittleEndian, &entry)
}

// List obtiene todas las entradas del journal
func (fj *FileJournal) List() ([]models.JournalEntry, error) {
	f, err := os.Open(fj.diskPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.Seek(fj.start, io.SeekStart)
	if err != nil {
		return nil, err
	}

	var entries []models.JournalEntry
	for i := int32(0); i < fj.maxCount; i++ {
		var entry models.JournalEntry
		err := binary.Read(f, binary.LittleEndian, &entry)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Si la operación es OpNone, consideramos que es una entrada vacía
		if entry.Op == models.OpNone {
			break
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// Clear limpia todas las entradas del journal
func (fj *FileJournal) Clear() error {
	f, err := os.OpenFile(fj.diskPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Seek(fj.start, io.SeekStart)
	if err != nil {
		return err
	}

	// Escribir entradas vacías (OpNone)
	emptyEntry := models.JournalEntry{Op: models.OpNone}
	for i := int32(0); i < fj.maxCount; i++ {
		err := binary.Write(f, binary.LittleEndian, &emptyEntry)
		if err != nil {
			return err
		}
	}

	return nil
}

// Count devuelve la cantidad de entradas usadas
func (fj *FileJournal) Count() (int32, error) {
	entries, err := fj.List()
	if err != nil {
		return 0, err
	}
	return int32(len(entries)), nil
}
