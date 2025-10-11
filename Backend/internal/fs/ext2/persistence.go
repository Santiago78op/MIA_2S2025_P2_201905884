package ext2

import (
	"fmt"
	"os"

	"MIA_2S2025_P2_201905884/internal/disk"
)

// getPartitionInfo obtiene información de la partición desde el disco
func getPartitionInfo(diskPath, partitionName string) (start int64, size int64, err error) {
	// Abrir disco
	f, err := os.Open(diskPath)
	if err != nil {
		return 0, 0, fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	// Leer MBR
	var mbr disk.MBR
	if err := disk.ReadStruct(f, 0, &mbr); err != nil {
		return 0, 0, fmt.Errorf("error al leer MBR: %v", err)
	}

	// Buscar partición
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

// writeEXT2ToDisk escribe todas las estructuras EXT2 al disco
func writeEXT2ToDisk(diskPath string, partStart int64, sb *Superblock) error {
	f, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	// 1. Escribir Superbloque
	sbData, err := SerializeSuperblock(sb)
	if err != nil {
		return fmt.Errorf("error al serializar superbloque: %v", err)
	}
	if err := disk.WriteBytesAt(f, partStart, sbData); err != nil {
		return fmt.Errorf("error al escribir superbloque: %v", err)
	}

	// 2. Inicializar bitmap de inodos (todos a 0 = libre, excepto los primeros 2)
	bitmapInodos := make([]byte, sb.S_inodes_count)
	bitmapInodos[0] = 1 // Inodo 0 ocupado (raíz)
	bitmapInodos[1] = 1 // Inodo 1 ocupado (users.txt)

	bitmapInodosOffset := partStart + int64(sb.S_bm_inode_start)
	if err := disk.WriteBytesAt(f, bitmapInodosOffset, bitmapInodos); err != nil {
		return fmt.Errorf("error al escribir bitmap de inodos: %v", err)
	}

	// 3. Inicializar bitmap de bloques (todos a 0 = libre, excepto los primeros 2)
	bitmapBloques := make([]byte, sb.S_blocks_count)
	bitmapBloques[0] = 1 // Bloque 0 ocupado (raíz)
	bitmapBloques[1] = 1 // Bloque 1 ocupado (users.txt)

	bitmapBloquesOffset := partStart + int64(sb.S_bm_block_start)
	if err := disk.WriteBytesAt(f, bitmapBloquesOffset, bitmapBloques); err != nil {
		return fmt.Errorf("error al escribir bitmap de bloques: %v", err)
	}

	// 4. Crear y escribir inodo raíz (inodo 0)
	rootInode := NewFolderInode(1, 1) // uid=1 (root), gid=1 (root)
	rootInode.IBlock[0] = 0            // Apunta al bloque 0

	rootInodeData, err := SerializeInode(rootInode)
	if err != nil {
		return fmt.Errorf("error al serializar inodo raíz: %v", err)
	}
	rootInodeOffset := partStart + int64(sb.S_inode_start)
	if err := disk.WriteBytesAt(f, rootInodeOffset, rootInodeData); err != nil {
		return fmt.Errorf("error al escribir inodo raíz: %v", err)
	}

	// 5. Crear bloque de carpeta raíz (bloque 0)
	rootBlock := NewFolderBlock()
	// Agregar entrada "." (self)
	rootBlock.AddEntry(".", 0)
	// Agregar entrada ".." (parent = self en raíz)
	rootBlock.AddEntry("..", 0)
	// Agregar entrada "users.txt"
	rootBlock.AddEntry("users.txt", 1)

	rootBlockData, err := SerializeFolderBlock(rootBlock)
	if err != nil {
		return fmt.Errorf("error al serializar bloque raíz: %v", err)
	}
	rootBlockOffset := partStart + int64(sb.S_block_start)
	if err := disk.WriteBytesAt(f, rootBlockOffset, rootBlockData); err != nil {
		return fmt.Errorf("error al escribir bloque raíz: %v", err)
	}

	// 6. Crear y escribir inodo de users.txt (inodo 1)
	usersInode := NewFileInode(1, 1)
	usersContent := GetUsersFileContent()
	usersInode.IS = int32(len(usersContent)) // Tamaño del archivo
	usersInode.IBlock[0] = 1                  // Apunta al bloque 1

	usersInodeData, err := SerializeInode(usersInode)
	if err != nil {
		return fmt.Errorf("error al serializar inodo users.txt: %v", err)
	}
	usersInodeOffset := partStart + int64(sb.S_inode_start) + int64(SUPERBLOCK_SIZE_ACTUAL)
	if err := disk.WriteBytesAt(f, usersInodeOffset, usersInodeData); err != nil {
		return fmt.Errorf("error al escribir inodo users.txt: %v", err)
	}

	// 7. Crear bloque de archivo users.txt (bloque 1)
	usersBlock := NewFileBlock()
	copy(usersBlock.BContent[:], usersContent)

	usersBlockData, err := SerializeFileBlock(usersBlock)
	if err != nil {
		return fmt.Errorf("error al serializar bloque users.txt: %v", err)
	}
	usersBlockOffset := partStart + int64(sb.S_block_start) + int64(DEFAULT_BLOCK_SIZE)
	if err := disk.WriteBytesAt(f, usersBlockOffset, usersBlockData); err != nil {
		return fmt.Errorf("error al escribir bloque users.txt: %v", err)
	}

	return nil
}

// readSuperblockFromDisk lee el superbloque desde el disco
func readSuperblockFromDisk(diskPath string, partStart int64) (*Superblock, error) {
	f, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	data, err := disk.ReadBytesAt(f, partStart, SUPERBLOCK_SIZE_ACTUAL)
	if err != nil {
		return nil, fmt.Errorf("error al leer superbloque: %v", err)
	}

	return DeserializeSuperblock(data)
}

// writeSuperblockToDisk escribe el superbloque al disco
func writeSuperblockToDisk(diskPath string, partStart int64, sb *Superblock) error {
	f, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	sbData, err := SerializeSuperblock(sb)
	if err != nil {
		return fmt.Errorf("error al serializar superbloque: %v", err)
	}

	if err := disk.WriteBytesAt(f, partStart, sbData); err != nil {
		return fmt.Errorf("error al escribir superbloque: %v", err)
	}

	return nil
}

// readInodeFromDisk lee un inodo desde el disco
func readInodeFromDisk(diskPath string, partStart int64, sb *Superblock, inodeIndex int32) (*Inode, error) {
	if inodeIndex < 0 || inodeIndex >= sb.S_inodes_count {
		return nil, fmt.Errorf("índice de inodo fuera de rango: %d", inodeIndex)
	}

	f, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	inodeOffset := partStart + int64(sb.S_inode_start) + (int64(inodeIndex) * int64(sb.S_inode_size))
	data, err := disk.ReadBytesAt(f, inodeOffset, int(sb.S_inode_size))
	if err != nil {
		return nil, fmt.Errorf("error al leer inodo: %v", err)
	}

	return DeserializeInode(data)
}

// writeInodeToDisk escribe un inodo al disco
func writeInodeToDisk(diskPath string, partStart int64, sb *Superblock, inodeIndex int32, inode *Inode) error {
	if inodeIndex < 0 || inodeIndex >= sb.S_inodes_count {
		return fmt.Errorf("índice de inodo fuera de rango: %d", inodeIndex)
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	inodeData, err := SerializeInode(inode)
	if err != nil {
		return fmt.Errorf("error al serializar inodo: %v", err)
	}

	inodeOffset := partStart + int64(sb.S_inode_start) + (int64(inodeIndex) * int64(sb.S_inode_size))
	if err := disk.WriteBytesAt(f, inodeOffset, inodeData); err != nil {
		return fmt.Errorf("error al escribir inodo: %v", err)
	}

	return nil
}

// readFolderBlockFromDisk lee un bloque de carpeta desde el disco
func readFolderBlockFromDisk(diskPath string, partStart int64, sb *Superblock, blockIndex int32) (*FolderBlock, error) {
	if blockIndex < 0 || blockIndex >= sb.S_blocks_count {
		return nil, fmt.Errorf("índice de bloque fuera de rango: %d", blockIndex)
	}

	f, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	blockOffset := partStart + int64(sb.S_block_start) + (int64(blockIndex) * int64(sb.S_block_size))
	data, err := disk.ReadBytesAt(f, blockOffset, int(sb.S_block_size))
	if err != nil {
		return nil, fmt.Errorf("error al leer bloque: %v", err)
	}

	return DeserializeFolderBlock(data)
}

// writeFolderBlockToDisk escribe un bloque de carpeta al disco
func writeFolderBlockToDisk(diskPath string, partStart int64, sb *Superblock, blockIndex int32, fb *FolderBlock) error {
	if blockIndex < 0 || blockIndex >= sb.S_blocks_count {
		return fmt.Errorf("índice de bloque fuera de rango: %d", blockIndex)
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	fbData, err := SerializeFolderBlock(fb)
	if err != nil {
		return fmt.Errorf("error al serializar folder block: %v", err)
	}

	blockOffset := partStart + int64(sb.S_block_start) + (int64(blockIndex) * int64(sb.S_block_size))
	if err := disk.WriteBytesAt(f, blockOffset, fbData); err != nil {
		return fmt.Errorf("error al escribir folder block: %v", err)
	}

	return nil
}

// readFileBlockFromDisk lee un bloque de archivo desde el disco
func readFileBlockFromDisk(diskPath string, partStart int64, sb *Superblock, blockIndex int32) (*FileBlock, error) {
	if blockIndex < 0 || blockIndex >= sb.S_blocks_count {
		return nil, fmt.Errorf("índice de bloque fuera de rango: %d", blockIndex)
	}

	f, err := os.Open(diskPath)
	if err != nil {
		return nil, fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	blockOffset := partStart + int64(sb.S_block_start) + (int64(blockIndex) * int64(sb.S_block_size))
	data, err := disk.ReadBytesAt(f, blockOffset, int(sb.S_block_size))
	if err != nil {
		return nil, fmt.Errorf("error al leer bloque: %v", err)
	}

	return DeserializeFileBlock(data)
}

// writeFileBlockToDisk escribe un bloque de archivo al disco
func writeFileBlockToDisk(diskPath string, partStart int64, sb *Superblock, blockIndex int32, fileBlock *FileBlock) error {
	if blockIndex < 0 || blockIndex >= sb.S_blocks_count {
		return fmt.Errorf("índice de bloque fuera de rango: %d", blockIndex)
	}

	f, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	fbData, err := SerializeFileBlock(fileBlock)
	if err != nil {
		return fmt.Errorf("error al serializar file block: %v", err)
	}

	blockOffset := partStart + int64(sb.S_block_start) + (int64(blockIndex) * int64(sb.S_block_size))
	if err := disk.WriteBytesAt(f, blockOffset, fbData); err != nil {
		return fmt.Errorf("error al escribir file block: %v", err)
	}

	return nil
}

// allocateInode busca y marca un inodo libre en el bitmap
func allocateInode(diskPath string, partStart int64, sb *Superblock) (int32, error) {
	f, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return -1, fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	// Leer bitmap de inodos
	bitmapOffset := partStart + int64(sb.S_bm_inode_start)
	bitmap, err := disk.ReadBytesAt(f, bitmapOffset, int(sb.S_inodes_count))
	if err != nil {
		return -1, fmt.Errorf("error al leer bitmap de inodos: %v", err)
	}

	// Buscar primer inodo libre
	for i := int32(0); i < sb.S_inodes_count; i++ {
		if bitmap[i] == 0 {
			// Marcar como ocupado
			bitmap[i] = 1
			if err := disk.WriteBytesAt(f, bitmapOffset, bitmap); err != nil {
				return -1, fmt.Errorf("error al escribir bitmap de inodos: %v", err)
			}

			// Actualizar contador en superbloque
			sb.S_free_inodes_count--
			sb.S_first_ino = i + 1
			if err := writeSuperblockToDisk(diskPath, partStart, sb); err != nil {
				return -1, fmt.Errorf("error al actualizar superbloque: %v", err)
			}

			return i, nil
		}
	}

	return -1, fmt.Errorf("no hay inodos libres")
}

// allocateBlock busca y marca un bloque libre en el bitmap
func allocateBlock(diskPath string, partStart int64, sb *Superblock) (int32, error) {
	f, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return -1, fmt.Errorf("error al abrir disco: %v", err)
	}
	defer f.Close()

	// Leer bitmap de bloques
	bitmapOffset := partStart + int64(sb.S_bm_block_start)
	bitmap, err := disk.ReadBytesAt(f, bitmapOffset, int(sb.S_blocks_count))
	if err != nil {
		return -1, fmt.Errorf("error al leer bitmap de bloques: %v", err)
	}

	// Buscar primer bloque libre
	for i := int32(0); i < sb.S_blocks_count; i++ {
		if bitmap[i] == 0 {
			// Marcar como ocupado
			bitmap[i] = 1
			if err := disk.WriteBytesAt(f, bitmapOffset, bitmap); err != nil {
				return -1, fmt.Errorf("error al escribir bitmap de bloques: %v", err)
			}

			// Actualizar contador en superbloque
			sb.S_free_blocks_count--
			sb.S_first_blo = i + 1
			if err := writeSuperblockToDisk(diskPath, partStart, sb); err != nil {
				return -1, fmt.Errorf("error al actualizar superbloque: %v", err)
			}

			return i, nil
		}
	}

	return -1, fmt.Errorf("no hay bloques libres")
}

// navigatePath navega una ruta y retorna el inodo del último elemento
// Si createMissing es true, crea los directorios faltantes
func navigatePath(diskPath string, partStart int64, sb *Superblock, path string, createMissing bool) (int32, error) {
	if path == "" || path == "/" {
		return 0, nil // Raíz siempre es el inodo 0
	}

	// Limpiar path
	if path[0] == '/' {
		path = path[1:]
	}
	if path == "" {
		return 0, nil
	}

	parts := splitPath(path)
	currentInode := int32(0) // Empezar desde raíz

	for i, part := range parts {
		if part == "" || part == "." {
			continue
		}

		// Leer inodo actual
		inode, err := readInodeFromDisk(diskPath, partStart, sb, currentInode)
		if err != nil {
			return -1, fmt.Errorf("error al leer inodo %d: %v", currentInode, err)
		}

		// Verificar que sea una carpeta
		if !inode.IsFolder() {
			return -1, fmt.Errorf("%s no es un directorio", parts[:i])
		}

		// Buscar la parte del path en los bloques del inodo
		found := false
		nextInode := int32(-1)

		for _, blockIdx := range inode.GetDirectBlocks() {
			folderBlock, err := readFolderBlockFromDisk(diskPath, partStart, sb, blockIdx)
			if err != nil {
				return -1, fmt.Errorf("error al leer bloque %d: %v", blockIdx, err)
			}

			if inodeIdx, ok := folderBlock.FindEntry(part); ok {
				nextInode = inodeIdx
				found = true
				break
			}
		}

		if !found {
			if !createMissing {
				return -1, fmt.Errorf("ruta no encontrada: %s", part)
			}

			// Crear nuevo directorio
			newInodeIdx, err := allocateInode(diskPath, partStart, sb)
			if err != nil {
				return -1, fmt.Errorf("error al asignar inodo: %v", err)
			}

			newBlockIdx, err := allocateBlock(diskPath, partStart, sb)
			if err != nil {
				return -1, fmt.Errorf("error al asignar bloque: %v", err)
			}

			// Crear nuevo inodo de carpeta
			newInode := NewFolderInode(1, 1)
			newInode.IBlock[0] = newBlockIdx
			if err := writeInodeToDisk(diskPath, partStart, sb, newInodeIdx, newInode); err != nil {
				return -1, fmt.Errorf("error al escribir inodo: %v", err)
			}

			// Crear nuevo bloque de carpeta con . y ..
			newFolderBlock := NewFolderBlock()
			newFolderBlock.AddEntry(".", newInodeIdx)
			newFolderBlock.AddEntry("..", currentInode)
			if err := writeFolderBlockToDisk(diskPath, partStart, sb, newBlockIdx, newFolderBlock); err != nil {
				return -1, fmt.Errorf("error al escribir bloque: %v", err)
			}

			// Agregar entrada al directorio padre
			inode, _ = readInodeFromDisk(diskPath, partStart, sb, currentInode)
			added := false
			for _, blockIdx := range inode.GetDirectBlocks() {
				parentBlock, err := readFolderBlockFromDisk(diskPath, partStart, sb, blockIdx)
				if err != nil {
					continue
				}
				if !parentBlock.IsFull() {
					parentBlock.AddEntry(part, newInodeIdx)
					writeFolderBlockToDisk(diskPath, partStart, sb, blockIdx, parentBlock)
					added = true
					break
				}
			}

			if !added {
				// Necesitamos un nuevo bloque para el directorio padre
				newParentBlockIdx, err := allocateBlock(diskPath, partStart, sb)
				if err != nil {
					return -1, fmt.Errorf("error al asignar bloque para padre: %v", err)
				}

				newParentBlock := NewFolderBlock()
				newParentBlock.AddEntry(part, newInodeIdx)
				writeFolderBlockToDisk(diskPath, partStart, sb, newParentBlockIdx, newParentBlock)

				// Actualizar inodo padre
				inode.SetBlock(newParentBlockIdx)
				writeInodeToDisk(diskPath, partStart, sb, currentInode, inode)
			}

			nextInode = newInodeIdx
		}

		currentInode = nextInode
	}

	return currentInode, nil
}

// splitPath divide una ruta en sus componentes
func splitPath(path string) []string {
	if path == "" || path == "/" {
		return []string{}
	}

	// Limpiar path
	if path[0] == '/' {
		path = path[1:]
	}
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	parts := []string{}
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// joinPath une partes de una ruta con '/'
func joinPath(parts []string) string {
	result := ""
	for i, part := range parts {
		result += part
		if i < len(parts)-1 {
			result += "/"
		}
	}
	return result
}
