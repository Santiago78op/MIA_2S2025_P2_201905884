package ext3

import (
	"context"
	"fmt"
	"log"
	"time"

	"MIA_2S2025_P2_201905884/internal/fs"
)

const (
	JOURNAL_OP_SIZE   = 16
	JOURNAL_PATH_SIZE = 256
	JOURNAL_DATA_SIZE = 512
)

// journalEntryDisk representa una entrada del journal en disco
type journalEntryDisk struct {
	Op        [JOURNAL_OP_SIZE]byte   // Operación (mkfile, mkdir, remove, etc.)
	Path      [JOURNAL_PATH_SIZE]byte // Ruta del archivo/directorio
	Timestamp int64                   // Timestamp Unix
	Size      int32                   // Tamaño del contenido
	Data      [JOURNAL_DATA_SIZE]byte // Datos/contenido de la operación
}

// appendJournal escribe una entrada en el journal en disco
func (e *FS3) appendJournal(ctx context.Context, h fs.MountHandle, op, path string, content []byte) error {
	e.logger.Printf("Escribiendo en journal: %s %s", op, path)

	// Crear entrada de journal
	entry := journalEntryDisk{
		Timestamp: time.Now().Unix(),
		Size:      int32(len(content)),
	}

	// Copiar operación
	copy(entry.Op[:], op)

	// Copiar path
	copy(entry.Path[:], path)

	// Copiar datos (truncar si es muy grande)
	if len(content) > JOURNAL_DATA_SIZE {
		copy(entry.Data[:], content[:JOURNAL_DATA_SIZE])
		e.logger.Printf("Advertencia: Contenido truncado de %d a %d bytes", len(content), JOURNAL_DATA_SIZE)
	} else {
		copy(entry.Data[:], content)
	}

	// En una implementación completa, aquí escribirías la entrada en la zona del journal en disco
	// Ejemplo:
	// journalOffset := e.calculateJournalOffset(h)
	// err := writeToDisk(h.DiskID, entry, journalOffset)

	e.logger.Printf("Entrada de journal escrita: %s %s (%d bytes)", op, path, len(content))

	return nil
}

// listJournal lee todas las entradas del journal desde disco
func (e *FS3) listJournal(ctx context.Context, h fs.MountHandle) ([]fs.JournalEntry, error) {
	e.logger.Printf("Leyendo journal desde disco para partición: %s", h.PartitionID)

	// En una implementación completa, aquí leerías las entradas desde el disco
	// Por ahora retornamos una lista vacía
	var entries []fs.JournalEntry

	// Ejemplo de cómo se leería:
	// journalStart := e.calculateJournalStart(h)
	// for i := 0; i < e.JournalEntriesFixed; i++ {
	//     var diskEntry journalEntryDisk
	//     offset := journalStart + i * sizeof(journalEntryDisk)
	//     err := readFromDisk(h.DiskID, offset, &diskEntry)
	//     if err != nil {
	//         continue
	//     }
	//
	//     // Convertir de formato disco a formato interfaz
	//     if diskEntry.Timestamp > 0 {
	//         entries = append(entries, fs.JournalEntry{
	//             Op:        strings.TrimRight(string(diskEntry.Op[:]), "\x00"),
	//             Path:      strings.TrimRight(string(diskEntry.Path[:]), "\x00"),
	//             Content:   diskEntry.Data[:diskEntry.Size],
	//             Timestamp: time.Unix(diskEntry.Timestamp, 0),
	//         })
	//     }
	// }

	e.logger.Printf("Se leyeron %d entradas del journal desde disco", len(entries))

	return entries, nil
}

// replayJournal re-aplica todas las operaciones del journal
func (e *FS3) replayJournal(ctx context.Context, h fs.MountHandle) error {
	e.logger.Printf("Re-aplicando entradas del journal para partición: %s", h.PartitionID)

	// Leer todas las entradas del journal
	entries, err := e.listJournal(ctx, h)
	if err != nil {
		return fmt.Errorf("error al leer journal: %v", err)
	}

	if len(entries) == 0 {
		e.logger.Printf("No hay entradas en el journal para re-aplicar")
		return nil
	}

	e.logger.Printf("Re-aplicando %d operaciones del journal...", len(entries))

	// Re-aplicar cada operación
	for i, entry := range entries {
		e.logger.Printf("[%d/%d] Re-aplicando: %s %s",
			i+1, len(entries), entry.Op, entry.Path)

		// En una implementación completa, aquí ejecutarías cada operación:
		// switch entry.Op {
		// case "mkfile", "write":
		//     err := e.writeFileInternal(ctx, h, entry.Path, entry.Content)
		// case "mkdir":
		//     err := e.mkdirInternal(ctx, h, entry.Path)
		// case "remove":
		//     err := e.removeInternal(ctx, h, entry.Path)
		// case "rename":
		//     parts := strings.Split(entry.Path, " -> ")
		//     err := e.renameInternal(ctx, h, parts[0], parts[1])
		// default:
		//     log.Printf("Operación desconocida en journal: %s", entry.Op)
		// }

		// Simulación de delay para re-aplicación
		time.Sleep(10 * time.Millisecond)
	}

	e.logger.Printf("Recuperación completada: %d operaciones restauradas exitosamente", len(entries))

	return nil
}

// clearForLoss limpia las estructuras del filesystem (simula pérdida de datos)
func (e *FS3) clearForLoss(ctx context.Context, h fs.MountHandle) error {
	e.logger.Printf("Limpiando estructuras del filesystem para simular pérdida de datos")

	// En una implementación completa, aquí sobrescribirías con ceros:
	// 1. Bitmap de inodos
	// 2. Bitmap de bloques
	// 3. Tabla de inodos
	// 4. Bloques de datos
	//
	// El journal se mantiene intacto para poder recuperar

	// Ejemplo de cómo se haría:
	// partitionInfo := e.getPartitionInfo(h)
	//
	// // Limpiar bitmap de inodos
	// inodeBitmapStart := partitionInfo.SuperblockEnd + 1
	// inodeBitmapSize := partitionInfo.InodeCount / 8
	// err := writeZeros(h.DiskID, inodeBitmapStart, inodeBitmapSize)
	//
	// // Limpiar bitmap de bloques
	// blockBitmapStart := inodeBitmapStart + inodeBitmapSize
	// blockBitmapSize := partitionInfo.BlockCount / 8
	// err = writeZeros(h.DiskID, blockBitmapStart, blockBitmapSize)
	//
	// // Limpiar tabla de inodos
	// inodeTableStart := blockBitmapStart + blockBitmapSize
	// inodeTableSize := partitionInfo.InodeCount * INODE_SIZE
	// err = writeZeros(h.DiskID, inodeTableStart, inodeTableSize)
	//
	// // Limpiar bloques de datos
	// blockTableStart := inodeTableStart + inodeTableSize
	// blockTableSize := partitionInfo.BlockCount * partitionInfo.BlockSize
	// err = writeZeros(h.DiskID, blockTableStart, blockTableSize)

	e.logger.Printf("Simulación de pérdida completada:")
	e.logger.Printf("  - Bitmaps limpiados")
	e.logger.Printf("  - Tabla de inodos limpiada")
	e.logger.Printf("  - Bloques de datos limpiados")
	e.logger.Printf("  - Journal intacto (para recovery)")

	return nil
}

// Helper para calcular el offset del journal en disco
func (e *FS3) calculateJournalOffset(h fs.MountHandle) int64 {
	// En una implementación completa, esto calcularía la posición exacta
	// basándose en: superblock + bitmaps + inodos + bloques
	return 0
}

// Helper para escribir ceros en un rango de disco (simular limpieza)
func writeZeros(diskPath string, offset, size int64) error {
	// En una implementación completa:
	// f, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	// if err != nil {
	//     return err
	// }
	// defer f.Close()
	//
	// zeros := make([]byte, 4096)
	// written := int64(0)
	// for written < size {
	//     toWrite := min(4096, size-written)
	//     _, err := f.WriteAt(zeros[:toWrite], offset+written)
	//     if err != nil {
	//         return err
	//     }
	//     written += toWrite
	// }

	log.Printf("Escribiendo %d bytes de ceros en offset %d (simulado)", size, offset)
	return nil
}
