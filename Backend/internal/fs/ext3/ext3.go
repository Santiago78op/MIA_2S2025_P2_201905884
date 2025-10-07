package ext3

import (
	"context"
	"fmt"
	"os"
	"time"

	"MIA_2S2025_P2_201905884/internal/disk"
	"MIA_2S2025_P2_201905884/internal/fs"
	"MIA_2S2025_P2_201905884/internal/fs/ext2"
	"MIA_2S2025_P2_201905884/internal/logger"
)

// getPartitionInfo obtiene información de la partición desde el disco
func getPartitionInfo(diskPath, partitionName string) (start int64, size int64, err error) {
	f, err := os.Open(diskPath)
	if err != nil {
		return 0, 0, fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	var mbr disk.MBR
	if err := disk.ReadStruct(f, 0, &mbr); err != nil {
		return 0, 0, fmt.Errorf("error al leer MBR: %v", err)
	}

	for i := 0; i < disk.MaxPrimaries; i++ {
		p := mbr.Parts[i]
		if p.Status == disk.PartStatusUsed && trimPartName(p.Name) == partitionName {
			return p.Start, p.Size, nil
		}
	}

	return 0, 0, fmt.Errorf("partición %s no encontrada", partitionName)
}

// trimPartName convierte [16]byte a string limpio
func trimPartName(n [16]byte) string {
	for i, b := range n {
		if b == 0 {
			return string(n[:i])
		}
	}
	return string(n[:])
}

type FS3 struct {
	state     *fs.MetaState
	blockSize int
}

func New(state *fs.MetaState, blockSize int, _ interface{}) *FS3 {
	return &FS3{
		state:     state,
		blockSize: blockSize,
	}
}

// Mkfs formatea una partición como EXT3
func (e *FS3) Mkfs(ctx context.Context, req fs.MkfsRequest) error {
	if req.FSKind != "3fs" {
		return fs.ErrUnsupported
	}

	logger.Info("Formateando partición EXT3", map[string]interface{}{
		"mount_id": req.MountID,
	})

	// NOTA: Por ahora usamos valores de ejemplo similar a EXT2
	// En producción esto se obtiene del mount handle/index
	diskPath := "/tmp/test.mia"  // TODO: obtener del adapter
	partitionName := req.MountID // TODO: obtener del adapter

	// 1. Obtener información de la partición
	partStart, partSize, err := getPartitionInfo(diskPath, partitionName)
	if err != nil {
		logger.Warn("No se pudo obtener info de partición", map[string]interface{}{
			"error": err.Error(),
		})
		// Usar valores por defecto para desarrollo
		partStart = 512
		partSize = 10 * 1024 * 1024 // 10MB
	}

	logger.Info("Partición encontrada", map[string]interface{}{
		"name":  partitionName,
		"start": partStart,
		"size":  partSize,
	})

	// 2. Abrir disco para escritura
	f, err := os.OpenFile(diskPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	// 3. Calcular n usando la fórmula CORRECTA
	n := CalcN(partSize, e.blockSize)
	if n < 2 {
		return fmt.Errorf("partición muy pequeña para EXT3: n=%d", n)
	}

	logger.Info("Cálculo de estructuras EXT3", map[string]interface{}{
		"n":      n,
		"inodes": n,
		"blocks": 3 * n,
	})

	// 4. Calcular offsets y crear SuperBlock
	sb := CalculateOffsets(n, e.blockSize)

	// 5. Inicializar Journal vacío
	journal := NewJournal()

	// 6. Inicializar bitmaps
	bmInodes := make([]byte, n)
	bmBlocks := make([]byte, 3*n)

	// 7. Inicializar tabla de inodos
	inodes := make([]ext2.Inode, n)

	// 8. Crear directorio raíz (inodo 0)
	rootInode := ext2.Inode{
		IUid:   0,
		IGid:   0,
		IS:     0,
		IAtime: time.Now().Unix(),
		ICtime: time.Now().Unix(),
		IMtime: time.Now().Unix(),
		IBlock: [15]int32{0, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		IType:  0, // directorio
		IPerm:  [3]byte{7, 5, 5},
	}
	inodes[0] = rootInode

	// Marcar inodo 0 y bloque 0 como usados
	bmInodes[0] = 1
	bmBlocks[0] = 1

	// Actualizar contadores del superblock
	sb.SFreeInodes = int32(n - 1)
	sb.SFreeBlocks = int32(3*n - 1)

	// 9. Crear bloque de directorio raíz
	rootBlock := ext2.NewFolderBlock()
	rootBlock.BContent[0] = ext2.Content{
		BName:  [12]byte{'/', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		BInodo: 0,
	}
	rootBlock.BContent[1] = ext2.Content{
		BName:  [12]byte{'.', '.', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		BInodo: 0,
	}

	// 10. Escribir estructuras al disco
	offset := partStart

	// Escribir SuperBlock
	sbBytes := sb.Serialize()
	if _, err := f.WriteAt(sbBytes, offset); err != nil {
		return fmt.Errorf("error escribiendo superblock: %v", err)
	}
	offset += 512

	// Escribir Journal
	journalBytes := journal.Serialize()
	if _, err := f.WriteAt(journalBytes, offset); err != nil {
		return fmt.Errorf("error escribiendo journal: %v", err)
	}
	offset += int64(len(journalBytes))

	// Escribir Bitmap de Inodos
	if _, err := f.WriteAt(bmInodes, offset); err != nil {
		return fmt.Errorf("error escribiendo bitmap inodos: %v", err)
	}
	offset += int64(len(bmInodes))

	// Escribir Bitmap de Bloques
	if _, err := f.WriteAt(bmBlocks, offset); err != nil {
		return fmt.Errorf("error escribiendo bitmap bloques: %v", err)
	}
	offset += int64(len(bmBlocks))

	// Escribir Tabla de Inodos
	for i := int64(0); i < n; i++ {
		inodeBytes, err := ext2.SerializeInode(&inodes[i])
		if err != nil {
			return fmt.Errorf("error serializando inodo %d: %v", i, err)
		}
		if _, err := f.WriteAt(inodeBytes, offset+i*128); err != nil {
			return fmt.Errorf("error escribiendo inodo %d: %v", i, err)
		}
	}
	offset += n * 128

	// Escribir Bloque raíz
	rootBlockBytes, err := ext2.SerializeFolderBlock(rootBlock)
	if err != nil {
		return fmt.Errorf("error serializando bloque raíz: %v", err)
	}
	if _, err := f.WriteAt(rootBlockBytes, offset); err != nil {
		return fmt.Errorf("error escribiendo bloque raíz: %v", err)
	}

	// 11. Registrar formato en journal
	journal.Append(NewJournalEntry("mkfs", "/", "EXT3 formatted", 0, 0, 0755))

	logger.Info("Formateo EXT3 completado", map[string]interface{}{
		"n":            n,
		"inodes":       n,
		"blocks":       3 * n,
		"journal_size": JournalEntryCount,
	})

	// 12. Guardar metadata
	e.state.Set(req.MountID, fs.Meta{
		FSKind:   "3fs",
		BlockSz:  e.blockSize,
		InodeSz:  128,
		JournalN: JournalEntryCount,
	})

	return nil
}

func (e *FS3) Mount(ctx context.Context, req fs.MountRequest) (fs.MountHandle, error) {
	logger.Info("Montando partición EXT3", map[string]interface{}{
		"disk":      req.DiskPath,
		"partition": req.Partition,
	})

	if _, err := os.Stat(req.DiskPath); err != nil {
		return fs.MountHandle{}, fmt.Errorf("disco no encontrado: %v", err)
	}

	return fs.MountHandle{
		DiskID:      req.DiskPath,
		PartitionID: req.Partition,
		User:        "root",
		Group:       "root",
	}, nil
}

func (e *FS3) Unmount(ctx context.Context, h fs.MountHandle) error {
	logger.Info("Desmontando partición EXT3", map[string]interface{}{
		"partition": h.PartitionID,
	})
	return nil
}

// Las demás funciones (Tree, ReadFile, WriteFile, etc.) se implementarán después
func (e *FS3) Tree(ctx context.Context, h fs.MountHandle, path string) (fs.TreeNode, error) {
	return fs.TreeNode{Path: "/", IsDir: true, Mode: 0755, Owner: "root", Group: "root"}, nil
}

func (e *FS3) ReadFile(ctx context.Context, h fs.MountHandle, path string) ([]byte, fs.FileStat, error) {
	return nil, fs.FileStat{}, fs.ErrNotFound
}

func (e *FS3) WriteFile(ctx context.Context, h fs.MountHandle, req fs.WriteFileRequest) error {
	return fmt.Errorf("not implemented yet")
}

func (e *FS3) Mkdir(ctx context.Context, h fs.MountHandle, req fs.MkdirRequest) error {
	return fmt.Errorf("not implemented yet")
}

func (e *FS3) Remove(ctx context.Context, h fs.MountHandle, path string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *FS3) Rename(ctx context.Context, h fs.MountHandle, from, to string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *FS3) Copy(ctx context.Context, h fs.MountHandle, from, to string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *FS3) Move(ctx context.Context, h fs.MountHandle, from, to string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *FS3) Find(ctx context.Context, h fs.MountHandle, req fs.FindRequest) ([]string, error) {
	return []string{}, nil
}

func (e *FS3) Chown(ctx context.Context, h fs.MountHandle, path, user, group string) error {
	return fmt.Errorf("not implemented yet")
}

func (e *FS3) Chmod(ctx context.Context, h fs.MountHandle, path string, perm uint16) error {
	return fmt.Errorf("not implemented yet")
}

// Métodos específicos EXT3
func (e *FS3) Journaling(ctx context.Context, h fs.MountHandle) ([]fs.JournalEntry, error) {
	// Leer journal desde disco
	_, ok := e.state.Get(h.PartitionID)
	if !ok {
		return nil, fmt.Errorf("partición no encontrada")
	}

	logger.Info("Obteniendo journal", map[string]interface{}{"partition": h.PartitionID})

	// TODO: Leer desde disco y parsear
	return []fs.JournalEntry{}, nil
}

func (e *FS3) Recovery(ctx context.Context, h fs.MountHandle) error {
	logger.Info("Iniciando recovery desde journal", map[string]interface{}{"partition": h.PartitionID})
	// TODO: Implementar recovery
	return nil
}

func (e *FS3) Loss(ctx context.Context, h fs.MountHandle) error {
	logger.Info("Simulando pérdida de datos", map[string]interface{}{"partition": h.PartitionID})
	// TODO: Limpiar bitmaps, inodos y bloques (NO journal ni superblock)
	return nil
}
