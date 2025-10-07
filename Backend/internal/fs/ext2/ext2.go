package ext2

import (
	"context"
	"fmt"
	"log"
	"os"

	"MIA_2S2025_P2_201905884/internal/fs"
)

// Constantes EXT2
const (
	EXT2_MAGIC          = 0xEF53
	EXT2_FILESYSTEM_TYPE = 2
	MIN_INODES          = 2
	MIN_BLOCKS          = 2
	DEFAULT_BLOCK_SIZE  = 64
)

type FS2 struct {
	state  *fs.MetaState
	logger *log.Logger
}

func New(state *fs.MetaState) *FS2 {
	return &FS2{
		state:  state,
		logger: log.New(os.Stdout, "[EXT2] ", log.LstdFlags),
	}
}

func (e *FS2) Mkfs(ctx context.Context, req fs.MkfsRequest) error {
	if req.FSKind != "2fs" {
		return fs.ErrUnsupported
	}

	e.logger.Printf("Formateando partición %s con EXT2...", req.MountID)

	// El req.MountID contiene el nombre de la partición
	// Necesitamos obtener información del disco y partición
	// Por ahora, asumimos que la info viene en el contexto o debemos buscarla

	// TEMPORAL: Usar valores de ejemplo hasta integrar con mount real
	// En producción, esto vendría del mount handle
	diskPath := "/tmp/test.mia"        // TODO: Obtener del mount handle
	partitionName := req.MountID       // TODO: Obtener del mount handle
	partitionSize := int64(10 * 1024 * 1024) // 10MB por defecto

	// Intentar obtener info real de la partición
	partStart, partSize, err := getPartitionInfo(diskPath, partitionName)
	if err != nil {
		e.logger.Printf("Advertencia: No se pudo obtener info de partición, usando valores por defecto: %v", err)
		partStart = 512 // Después del MBR
		partSize = partitionSize
	} else {
		partitionSize = partSize
	}

	e.logger.Printf("Partición encontrada: start=%d, size=%d bytes", partStart, partitionSize)

	// Crear superbloque
	sb := NewSuperblock(partitionSize, DEFAULT_BLOCK_SIZE)

	// Validar superbloque
	if err := ValidateEXT2Structures(sb); err != nil {
		e.logger.Printf("Error al validar superbloque: %v", err)
		return fmt.Errorf("error al validar superbloque: %v", err)
	}

	e.logger.Printf("Superbloque creado: %d inodos, %d bloques", sb.S_inodes_count, sb.S_blocks_count)

	// Escribir todas las estructuras al disco
	if err := writeEXT2ToDisk(diskPath, partStart, sb); err != nil {
		e.logger.Printf("Error al escribir EXT2 al disco: %v", err)
		return fmt.Errorf("error al escribir EXT2 al disco: %v", err)
	}

	// Crear metadatos del filesystem
	meta := fs.Meta{
		FSKind:   "2fs",
		BlockSz:  int(sb.S_block_size),
		InodeSz:  int(sb.S_inode_size),
		JournalN: 0, // EXT2 no tiene journal
	}

	// Guardar en el estado
	e.state.Set(req.MountID, meta)

	e.logger.Printf("✅ Formateo EXT2 completado exitosamente:")
	e.logger.Printf("   - %d inodos (%d libres)", sb.S_inodes_count, sb.S_free_inodes_count)
	e.logger.Printf("   - %d bloques (%d libres)", sb.S_blocks_count, sb.S_free_blocks_count)
	e.logger.Printf("   - Superbloque escrito en offset %d", partStart)
	e.logger.Printf("   - Estructura inicial: raíz + users.txt")

	return nil
}

func (e *FS2) Mount(ctx context.Context, req fs.MountRequest) (fs.MountHandle, error) {
	e.logger.Printf("Montando partición %s desde %s", req.Partition, req.DiskPath)

	// Validar que el disco existe
	if _, err := os.Stat(req.DiskPath); err != nil {
		return fs.MountHandle{}, fmt.Errorf("disco no encontrado: %v", err)
	}

	// Crear handle
	handle := fs.MountHandle{
		DiskID:      req.DiskPath,
		PartitionID: req.Partition,
		User:        "root",
		Group:       "root",
	}

	return handle, nil
}

func (e *FS2) Unmount(ctx context.Context, h fs.MountHandle) error {
	e.logger.Printf("Desmontando partición %s", h.PartitionID)
	return nil
}

func (e *FS2) Tree(ctx context.Context, h fs.MountHandle, path string) (fs.TreeNode, error) {
	// Construir árbol básico con directorio raíz
	e.logger.Printf("Construyendo árbol para path: %s", path)

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

func (e *FS2) ReadFile(ctx context.Context, h fs.MountHandle, path string) ([]byte, fs.FileStat, error) {
	e.logger.Printf("Leyendo archivo: %s", path)

	// Por ahora retornamos contenido de ejemplo para users.txt
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

func (e *FS2) WriteFile(ctx context.Context, h fs.MountHandle, req fs.WriteFileRequest) error {
	e.logger.Printf("Escribiendo archivo: %s (%d bytes)", req.Path, len(req.Content))

	// Implementación básica: aceptamos la escritura pero no persistimos aún
	// En una implementación completa, aquí escribirías en el disco
	e.logger.Printf("Advertencia: WriteFile no persistente todavía")

	return nil
}

func (e *FS2) Mkdir(ctx context.Context, h fs.MountHandle, req fs.MkdirRequest) error {
	e.logger.Printf("Creando directorio: %s (deep=%v)", req.Path, req.Deep)

	// Implementación básica
	e.logger.Printf("Advertencia: Mkdir no persistente todavía")

	return nil
}

func (e *FS2) Remove(ctx context.Context, h fs.MountHandle, path string) error {
	e.logger.Printf("Eliminando: %s", path)

	// Validar que no se elimine la raíz
	if path == "" || path == "/" {
		return fmt.Errorf("no se puede eliminar la ruta raíz")
	}

	// TODO: Implementación completa con validación de permisos
	// - Verificar permisos de escritura en el archivo/directorio
	// - Si es directorio, verificar permisos recursivamente en todos los hijos
	// - Si algún hijo no tiene permisos, no eliminar nada (rollback completo)

	e.logger.Printf("Advertencia: Remove no persistente todavía")
	return nil
}

func (e *FS2) Rename(ctx context.Context, h fs.MountHandle, from, to string) error {
	e.logger.Printf("Renombrando: %s -> %s", from, to)
	e.logger.Printf("Advertencia: Rename no persistente todavía")
	return nil
}

func (e *FS2) Copy(ctx context.Context, h fs.MountHandle, from, to string) error {
	e.logger.Printf("Copiando: %s -> %s", from, to)
	e.logger.Printf("Advertencia: Copy no persistente todavía")
	return nil
}

func (e *FS2) Move(ctx context.Context, h fs.MountHandle, from, to string) error {
	e.logger.Printf("Moviendo: %s -> %s", from, to)
	e.logger.Printf("Advertencia: Move no persistente todavía")
	return nil
}

func (e *FS2) Find(ctx context.Context, h fs.MountHandle, req fs.FindRequest) ([]string, error) {
	e.logger.Printf("Buscando archivos: base=%s, pattern=%s", req.BasePath, req.Pattern)

	// Retornar lista básica de ejemplo
	results := []string{"/users.txt"}

	return results, nil
}

func (e *FS2) Chown(ctx context.Context, h fs.MountHandle, path, user, group string) error {
	e.logger.Printf("Cambiando propietario de %s a %s:%s", path, user, group)
	e.logger.Printf("Advertencia: Chown no persistente todavía")
	return nil
}

func (e *FS2) Chmod(ctx context.Context, h fs.MountHandle, path string, perm uint16) error {
	e.logger.Printf("Cambiando permisos de %s a %o", path, perm)
	e.logger.Printf("Advertencia: Chmod no persistente todavía")
	return nil
}

// Métodos de journaling (no aplica para EXT2, retornan valores vacíos)
func (e *FS2) Journaling(ctx context.Context, h fs.MountHandle) ([]fs.JournalEntry, error) {
	return nil, nil
}

func (e *FS2) Recovery(ctx context.Context, h fs.MountHandle) error {
	return nil
}

func (e *FS2) Loss(ctx context.Context, h fs.MountHandle) error {
	return nil
}

// Funciones helper movidas a superblock.go, inode.go y blocks.go
