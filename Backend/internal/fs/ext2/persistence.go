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
