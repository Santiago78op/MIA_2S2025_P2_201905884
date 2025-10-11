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

	// Obtener información del disco y partición desde el request
	diskPath := req.DiskPath
	partitionName := req.PartitionID

	e.logger.Printf("Disco: %s, Partición: %s", diskPath, partitionName)

	// Obtener info real de la partición
	partStart, partSize, err := getPartitionInfo(diskPath, partitionName)
	if err != nil {
		e.logger.Printf("Error: No se pudo obtener info de partición: %v", err)
		return fmt.Errorf("no se pudo obtener información de la partición %s: %v", partitionName, err)
	}

	partitionSize := partSize

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

	// Obtener información de la partición
	partStart, _, err := getPartitionInfo(h.DiskID, h.PartitionID)
	if err != nil {
		return nil, fs.FileStat{}, fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	sb, err := readSuperblockFromDisk(h.DiskID, partStart)
	if err != nil {
		return nil, fs.FileStat{}, fmt.Errorf("error al leer superbloque: %v", err)
	}

	// Separar directorio y nombre de archivo
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, fs.FileStat{}, fmt.Errorf("ruta inválida")
	}

	fileName := parts[len(parts)-1]
	dirPath := ""
	if len(parts) > 1 {
		dirPath = "/" + joinPath(parts[:len(parts)-1])
	}

	// Navegar al directorio padre
	dirInodeIdx, err := navigatePath(h.DiskID, partStart, sb, dirPath, false)
	if err != nil {
		return nil, fs.FileStat{}, fmt.Errorf("directorio no encontrado: %v", err)
	}

	// Leer inodo del directorio
	dirInode, err := readInodeFromDisk(h.DiskID, partStart, sb, dirInodeIdx)
	if err != nil {
		return nil, fs.FileStat{}, fmt.Errorf("error al leer inodo de directorio: %v", err)
	}

	// Buscar el archivo en los bloques del directorio
	fileInodeIdx := int32(-1)
	for _, blockIdx := range dirInode.GetDirectBlocks() {
		folderBlock, err := readFolderBlockFromDisk(h.DiskID, partStart, sb, blockIdx)
		if err != nil {
			continue
		}
		if inodeIdx, ok := folderBlock.FindEntry(fileName); ok {
			fileInodeIdx = inodeIdx
			break
		}
	}

	if fileInodeIdx == -1 {
		return nil, fs.FileStat{}, fs.ErrNotFound
	}

	// Leer inodo del archivo
	fileInode, err := readInodeFromDisk(h.DiskID, partStart, sb, fileInodeIdx)
	if err != nil {
		return nil, fs.FileStat{}, fmt.Errorf("error al leer inodo de archivo: %v", err)
	}

	// Verificar que sea un archivo
	if !fileInode.IsFile() {
		return nil, fs.FileStat{}, fmt.Errorf("no es un archivo")
	}

	// Leer contenido de los bloques
	content := []byte{}
	for _, blockIdx := range fileInode.GetDirectBlocks() {
		fileBlock, err := readFileBlockFromDisk(h.DiskID, partStart, sb, blockIdx)
		if err != nil {
			return nil, fs.FileStat{}, fmt.Errorf("error al leer bloque de archivo: %v", err)
		}
		content = append(content, fileBlock.BContent[:]...)
	}

	// Truncar al tamaño real
	if int32(len(content)) > fileInode.IS {
		content = content[:fileInode.IS]
	}

	stat := fs.FileStat{
		Size:  int64(fileInode.IS),
		Mode:  0644,
		Owner: "root",
		Group: "root",
		IsDir: false,
	}

	return content, stat, nil
}

func (e *FS2) WriteFile(ctx context.Context, h fs.MountHandle, req fs.WriteFileRequest) error {
	e.logger.Printf("Escribiendo archivo: %s (%d bytes)", req.Path, len(req.Content))

	// Obtener información de la partición
	partStart, _, err := getPartitionInfo(h.DiskID, h.PartitionID)
	if err != nil {
		return fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	sb, err := readSuperblockFromDisk(h.DiskID, partStart)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	// Separar directorio y nombre de archivo
	parts := splitPath(req.Path)
	if len(parts) == 0 {
		return fmt.Errorf("ruta inválida")
	}

	fileName := parts[len(parts)-1]
	dirPath := ""
	if len(parts) > 1 {
		dirPath = "/" + joinPath(parts[:len(parts)-1])
	}

	// Navegar al directorio padre (crear si no existe)
	dirInodeIdx, err := navigatePath(h.DiskID, partStart, sb, dirPath, true)
	if err != nil {
		return fmt.Errorf("error al navegar al directorio: %v", err)
	}

	// Releer superbloque después de posibles cambios
	sb, _ = readSuperblockFromDisk(h.DiskID, partStart)

	// Leer inodo del directorio
	dirInode, err := readInodeFromDisk(h.DiskID, partStart, sb, dirInodeIdx)
	if err != nil {
		return fmt.Errorf("error al leer inodo de directorio: %v", err)
	}

	// Buscar si el archivo ya existe
	fileInodeIdx := int32(-1)
	for _, blockIdx := range dirInode.GetDirectBlocks() {
		folderBlock, err := readFolderBlockFromDisk(h.DiskID, partStart, sb, blockIdx)
		if err != nil {
			continue
		}
		if inodeIdx, ok := folderBlock.FindEntry(fileName); ok {
			fileInodeIdx = inodeIdx
			break
		}
	}

	// Si el archivo no existe, crearlo
	if fileInodeIdx == -1 {
		// Asignar nuevo inodo
		fileInodeIdx, err = allocateInode(h.DiskID, partStart, sb)
		if err != nil {
			return fmt.Errorf("error al asignar inodo: %v", err)
		}

		// Releer superbloque
		sb, _ = readSuperblockFromDisk(h.DiskID, partStart)

		// Agregar entrada al directorio padre
		dirInode, _ = readInodeFromDisk(h.DiskID, partStart, sb, dirInodeIdx)
		added := false
		for _, blockIdx := range dirInode.GetDirectBlocks() {
			parentBlock, err := readFolderBlockFromDisk(h.DiskID, partStart, sb, blockIdx)
			if err != nil {
				continue
			}
			if !parentBlock.IsFull() {
				parentBlock.AddEntry(fileName, fileInodeIdx)
				writeFolderBlockToDisk(h.DiskID, partStart, sb, blockIdx, parentBlock)
				added = true
				break
			}
		}

		if !added {
			// Necesitamos un nuevo bloque para el directorio
			newBlockIdx, err := allocateBlock(h.DiskID, partStart, sb)
			if err != nil {
				return fmt.Errorf("error al asignar bloque: %v", err)
			}

			sb, _ = readSuperblockFromDisk(h.DiskID, partStart)
			newBlock := NewFolderBlock()
			newBlock.AddEntry(fileName, fileInodeIdx)
			writeFolderBlockToDisk(h.DiskID, partStart, sb, newBlockIdx, newBlock)

			// Actualizar inodo del directorio
			dirInode.SetBlock(newBlockIdx)
			writeInodeToDisk(h.DiskID, partStart, sb, dirInodeIdx, dirInode)
		}
	}

	// Releer superbloque
	sb, _ = readSuperblockFromDisk(h.DiskID, partStart)

	// Crear inodo de archivo
	fileInode := NewFileInode(1, 1)
	fileInode.IS = int32(len(req.Content))

	// Calcular cuántos bloques necesitamos
	blocksNeeded := (len(req.Content) + int(sb.S_block_size) - 1) / int(sb.S_block_size)
	if blocksNeeded > 12 {
		blocksNeeded = 12 // Máximo 12 bloques directos
	}

	// Asignar y escribir bloques
	offset := 0
	for i := 0; i < blocksNeeded; i++ {
		blockIdx, err := allocateBlock(h.DiskID, partStart, sb)
		if err != nil {
			return fmt.Errorf("error al asignar bloque: %v", err)
		}

		sb, _ = readSuperblockFromDisk(h.DiskID, partStart)

		// Crear bloque de archivo
		fileBlock := NewFileBlock()
		end := offset + int(sb.S_block_size)
		if end > len(req.Content) {
			end = len(req.Content)
		}
		copy(fileBlock.BContent[:], req.Content[offset:end])

		// Escribir bloque al disco
		if err := writeFileBlockToDisk(h.DiskID, partStart, sb, blockIdx, fileBlock); err != nil {
			return fmt.Errorf("error al escribir bloque: %v", err)
		}

		// Agregar bloque al inodo
		fileInode.IBlock[i] = blockIdx
		offset = end
	}

	// Escribir inodo del archivo
	if err := writeInodeToDisk(h.DiskID, partStart, sb, fileInodeIdx, fileInode); err != nil {
		return fmt.Errorf("error al escribir inodo: %v", err)
	}

	e.logger.Printf("✅ Archivo escrito exitosamente: %s (%d bytes)", req.Path, len(req.Content))
	return nil
}

func (e *FS2) Mkdir(ctx context.Context, h fs.MountHandle, req fs.MkdirRequest) error {
	e.logger.Printf("Creando directorio: %s (deep=%v)", req.Path, req.Deep)

	// Obtener información de la partición
	partStart, _, err := getPartitionInfo(h.DiskID, h.PartitionID)
	if err != nil {
		return fmt.Errorf("error al obtener info de partición: %v", err)
	}

	// Leer superbloque
	sb, err := readSuperblockFromDisk(h.DiskID, partStart)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	// Si req.Deep es true, navigatePath creará todos los directorios faltantes
	_, err = navigatePath(h.DiskID, partStart, sb, req.Path, req.Deep)
	if err != nil {
		return fmt.Errorf("error al crear directorio: %v", err)
	}

	e.logger.Printf("✅ Directorio creado exitosamente: %s", req.Path)
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
