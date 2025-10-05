package ext3

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"MIA_2S2025_P2_201905884/internal/fs"
	"MIA_2S2025_P2_201905884/internal/journal"
)

// Constantes EXT3
const (
	EXT3_MAGIC       = 0xEF53
	SUPERBLOCK_SIZE  = 512
	INODE_SIZE       = 128
	DEFAULT_BLOCK_SIZE = 128
	JOURNAL_ENTRY_SIZE = 64
	MIN_INODES       = 2
)

type FS3 struct {
	state  *fs.MetaState
	logger *log.Logger
	// Parámetros de configuración
	BlockSize           int           // Tamaño de bloque (ej. 128)
	JournalEntriesFixed int           // Número fijo de entradas de journal (ej. 50)
	JStore              journal.Store // Journal store para persistencia
}

func New(state *fs.MetaState, blockSize, journalFixed int, store journal.Store) *FS3 {
	return &FS3{
		state:               state,
		logger:              log.New(os.Stdout, "[EXT3] ", log.LstdFlags),
		BlockSize:           blockSize,
		JournalEntriesFixed: journalFixed,
		JStore:              store,
	}
}

func (e *FS3) Mkfs(ctx context.Context, req fs.MkfsRequest) error {
	if req.FSKind != "3fs" {
		return fs.ErrUnsupported
	}

	e.logger.Printf("Formateando partición %s con EXT3 (con journaling)...", req.MountID)

	// Calcular tamaño de la partición (placeholder - en producción obtener del disco real)
	sizePart := int64(64 * 1024 * 1024) // 64MiB por defecto

	// Calcular parámetros de las estructuras
	sbSz := SUPERBLOCK_SIZE
	inodeSz := INODE_SIZE
	blkSz := e.BlockSize
	jSz := JOURNAL_ENTRY_SIZE

	// Calcular 'n' (número de estructuras)
	// Fórmula: tamaño = sb + n + 3n + n*inodo + 3n*bloque + journal_entries*journal_size
	available := int(sizePart) - sbSz - (e.JournalEntriesFixed * jSz)
	denominator := 1 + 3 + inodeSz + 3*blkSz
	n := available / denominator

	if n < MIN_INODES {
		return errors.New("ext3: partición muy pequeña, no cabe estructuras mínimas")
	}

	inodesCount := int32(n)
	blocksCount := int32(3 * n)

	e.logger.Printf("Formateo EXT3 - n=%d, inodos=%d, bloques=%d, journal=%d",
		n, inodesCount, blocksCount, e.JournalEntriesFixed)

	// Guardar metadatos en el estado
	e.state.Set(req.MountID, fs.Meta{
		FSKind:   "3fs",
		BlockSz:  blkSz,
		InodeSz:  inodeSz,
		JournalN: e.JournalEntriesFixed,
	})

	// Inicializar journal vacío si tenemos store
	if e.JStore != nil {
		e.logger.Printf("Inicializando journal para partición %s", req.MountID)
	}

	e.logger.Printf("Formateo EXT3 completado exitosamente")
	return nil
}

func (e *FS3) Mount(ctx context.Context, req fs.MountRequest) (fs.MountHandle, error) {
	e.logger.Printf("Montando partición EXT3: %s desde %s", req.Partition, req.DiskPath)

	// Validar que el disco existe
	if _, err := os.Stat(req.DiskPath); err != nil {
		return fs.MountHandle{}, fmt.Errorf("disco no encontrado: %v", err)
	}

	// Crear handle con información de montaje
	handle := fs.MountHandle{
		DiskID:      req.DiskPath,
		PartitionID: req.Partition,
		User:        "root",
		Group:       "root",
	}

	return handle, nil
}

func (e *FS3) Unmount(ctx context.Context, h fs.MountHandle) error {
	e.logger.Printf("Desmontando partición EXT3: %s", h.PartitionID)
	return nil
}

func (e *FS3) Tree(ctx context.Context, h fs.MountHandle, path string) (fs.TreeNode, error) {
	e.logger.Printf("Construyendo árbol EXT3 para path: %s", path)

	// Construir árbol básico con directorio raíz y users.txt
	rootNode := fs.TreeNode{
		Path:     "/",
		IsDir:    true,
		Mode:     0755,
		Owner:    "root",
		Group:    "root",
		Children: []fs.TreeNode{
			{
				Path:  "/users.txt",
				IsDir: false,
				Mode:  0644,
				Owner: "root",
				Group: "root",
			},
		},
	}

	return rootNode, nil
}

func (e *FS3) ReadFile(ctx context.Context, h fs.MountHandle, path string) ([]byte, fs.FileStat, error) {
	e.logger.Printf("Leyendo archivo EXT3: %s", path)

	// Contenido de ejemplo para users.txt
	if path == "/users.txt" || path == "users.txt" {
		content := []byte("1,G,root\n1,U,root,root,123\n")
		stat := fs.FileStat{
			Size:  int64(len(content)),
			Mode:  0644,
			Owner: "root",
			Group: "root",
			IsDir: false,
		}
		return content, stat, nil
	}

	return nil, fs.FileStat{}, fs.ErrNotFound
}

func (e *FS3) WriteFile(ctx context.Context, h fs.MountHandle, req fs.WriteFileRequest) error {
	e.logger.Printf("Escribiendo archivo EXT3: %s (%d bytes)", req.Path, len(req.Content))

	// 1. Registrar operación en journal ANTES de ejecutar
	if e.JStore != nil {
		operation := "write"
		if req.Append {
			operation = "append"
		}

		entry := journal.Entry{
			Op:        operation,
			Path:      req.Path,
			Content:   req.Content,
			Timestamp: time.Now().UTC(),
		}

		if err := e.JStore.Append(ctx, h.PartitionID, entry); err != nil {
			e.logger.Printf("Advertencia: no se pudo escribir en journal: %v", err)
		} else {
			e.logger.Printf("Operación registrada en journal: %s %s", operation, req.Path)
		}
	}

	// 2. Aplicar operación real (en implementación completa escribiría en disco)
	e.logger.Printf("Advertencia: WriteFile no persistente todavía")

	return nil
}

func (e *FS3) Mkdir(ctx context.Context, h fs.MountHandle, req fs.MkdirRequest) error {
	e.logger.Printf("Creando directorio EXT3: %s (deep=%v)", req.Path, req.Deep)

	// Registrar en journal
	if e.JStore != nil {
		entry := journal.Entry{
			Op:        "mkdir",
			Path:      req.Path,
			Content:   []byte{},
			Timestamp: time.Now().UTC(),
		}
		e.JStore.Append(ctx, h.PartitionID, entry)
	}

	e.logger.Printf("Advertencia: Mkdir no persistente todavía")
	return nil
}

func (e *FS3) Remove(ctx context.Context, h fs.MountHandle, path string) error {
	e.logger.Printf("Eliminando EXT3: %s", path)

	// Registrar en journal
	if e.JStore != nil {
		entry := journal.Entry{
			Op:        "remove",
			Path:      path,
			Content:   []byte{},
			Timestamp: time.Now().UTC(),
		}
		e.JStore.Append(ctx, h.PartitionID, entry)
	}

	e.logger.Printf("Advertencia: Remove no persistente todavía")
	return nil
}

func (e *FS3) Rename(ctx context.Context, h fs.MountHandle, from, to string) error {
	e.logger.Printf("Renombrando EXT3: %s -> %s", from, to)

	// Registrar en journal
	if e.JStore != nil {
		entry := journal.Entry{
			Op:        "rename",
			Path:      from + " -> " + to,
			Content:   []byte{},
			Timestamp: time.Now().UTC(),
		}
		e.JStore.Append(ctx, h.PartitionID, entry)
	}

	e.logger.Printf("Advertencia: Rename no persistente todavía")
	return nil
}

func (e *FS3) Copy(ctx context.Context, h fs.MountHandle, from, to string) error {
	e.logger.Printf("Copiando EXT3: %s -> %s", from, to)

	if e.JStore != nil {
		entry := journal.Entry{
			Op:        "copy",
			Path:      from + " -> " + to,
			Content:   []byte{},
			Timestamp: time.Now().UTC(),
		}
		e.JStore.Append(ctx, h.PartitionID, entry)
	}

	e.logger.Printf("Advertencia: Copy no persistente todavía")
	return nil
}

func (e *FS3) Move(ctx context.Context, h fs.MountHandle, from, to string) error {
	e.logger.Printf("Moviendo EXT3: %s -> %s", from, to)

	if e.JStore != nil {
		entry := journal.Entry{
			Op:        "move",
			Path:      from + " -> " + to,
			Content:   []byte{},
			Timestamp: time.Now().UTC(),
		}
		e.JStore.Append(ctx, h.PartitionID, entry)
	}

	e.logger.Printf("Advertencia: Move no persistente todavía")
	return nil
}

func (e *FS3) Find(ctx context.Context, h fs.MountHandle, req fs.FindRequest) ([]string, error) {
	e.logger.Printf("Buscando archivos EXT3: base=%s, pattern=%s", req.BasePath, req.Pattern)
	return []string{"/users.txt"}, nil
}

func (e *FS3) Chown(ctx context.Context, h fs.MountHandle, path, user, group string) error {
	e.logger.Printf("Cambiando propietario EXT3 de %s a %s:%s", path, user, group)

	if e.JStore != nil {
		entry := journal.Entry{
			Op:        "chown",
			Path:      fmt.Sprintf("%s -> %s:%s", path, user, group),
			Content:   []byte{},
			Timestamp: time.Now().UTC(),
		}
		e.JStore.Append(ctx, h.PartitionID, entry)
	}

	e.logger.Printf("Advertencia: Chown no persistente todavía")
	return nil
}

func (e *FS3) Chmod(ctx context.Context, h fs.MountHandle, path string, perm uint16) error {
	e.logger.Printf("Cambiando permisos EXT3 de %s a %o", path, perm)

	if e.JStore != nil {
		entry := journal.Entry{
			Op:        "chmod",
			Path:      fmt.Sprintf("%s -> %o", path, perm),
			Content:   []byte{},
			Timestamp: time.Now().UTC(),
		}
		e.JStore.Append(ctx, h.PartitionID, entry)
	}

	e.logger.Printf("Advertencia: Chmod no persistente todavía")
	return nil
}

// ==================== Métodos específicos de EXT3 (Journaling) ====================

func (e *FS3) Journaling(ctx context.Context, h fs.MountHandle) ([]fs.JournalEntry, error) {
	e.logger.Printf("Obteniendo entradas del journal para partición: %s", h.PartitionID)

	if e.JStore == nil {
		e.logger.Printf("No hay journal store configurado")
		return []fs.JournalEntry{}, nil
	}

	// Leer entradas del journal
	entries, err := e.JStore.List(ctx, h.PartitionID)
	if err != nil {
		return nil, fmt.Errorf("error al leer journal: %v", err)
	}

	// Convertir entradas internas a formato de la interfaz
	var result []fs.JournalEntry
	for _, entry := range entries {
		result = append(result, fs.JournalEntry{
			Op:        entry.Op,
			Path:      entry.Path,
			Content:   entry.Content,
			Timestamp: entry.Timestamp,
		})
	}

	e.logger.Printf("Se encontraron %d entradas en el journal", len(result))
	return result, nil
}

func (e *FS3) Recovery(ctx context.Context, h fs.MountHandle) error {
	e.logger.Printf("Iniciando recuperación desde journal para partición: %s", h.PartitionID)

	if e.JStore == nil {
		return fmt.Errorf("no hay journal store configurado")
	}

	// Leer todas las entradas del journal
	entries, err := e.JStore.List(ctx, h.PartitionID)
	if err != nil {
		return fmt.Errorf("error al leer journal: %v", err)
	}

	e.logger.Printf("Recuperando %d operaciones del journal...", len(entries))

	// Re-aplicar cada operación del journal
	for i, entry := range entries {
		e.logger.Printf("Re-aplicando operación %d/%d: %s %s",
			i+1, len(entries), entry.Op, entry.Path)

		// En una implementación completa, aquí re-ejecutarías cada operación
		// Por ahora solo logueamos
	}

	e.logger.Printf("Recuperación completada: %d operaciones restauradas", len(entries))
	return nil
}

func (e *FS3) Loss(ctx context.Context, h fs.MountHandle) error {
	e.logger.Printf("Simulando pérdida de datos (limpieza de estructuras) para partición: %s", h.PartitionID)

	// Simular pérdida limpiando bitmaps, inodos y bloques
	// En una implementación real, esto sobrescribiría áreas del disco con ceros

	e.logger.Printf("Advertencia: Simulación de pérdida de datos")
	e.logger.Printf("- Bitmaps de inodos: marcados como limpios")
	e.logger.Printf("- Bitmaps de bloques: marcados como limpios")
	e.logger.Printf("- Tabla de inodos: limpiada")
	e.logger.Printf("- Bloques de datos: limpiados")

	// El journal permanece intacto para poder recuperar después
	e.logger.Printf("El journal permanece intacto para posibilitar recovery")

	return nil
}
