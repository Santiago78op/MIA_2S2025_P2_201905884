package ext2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"
)

// Superblock contiene información sobre el sistema de archivos EXT2
type Superblock struct {
	S_filesystem_type   int32 `binary:"little"` // Número que identifica el sistema de archivos utilizado (2)
	S_inodes_count      int32 `binary:"little"` // Número total de inodos
	S_blocks_count      int32 `binary:"little"` // Número total de bloques
	S_free_blocks_count int32 `binary:"little"` // Número de bloques libres
	S_free_inodes_count int32 `binary:"little"` // Número de inodos libres
	S_mtime             int64 `binary:"little"` // Fecha y hora de la última vez que se montó
	S_umtime            int64 `binary:"little"` // Fecha y hora de la última vez que se desmontó
	S_mnt_count         int32 `binary:"little"` // Número de veces que se ha montado
	S_magic             int32 `binary:"little"` // Identificador del sistema de archivos (0xEF53)
	S_inode_size        int32 `binary:"little"` // Tamaño del inodo
	S_block_size        int32 `binary:"little"` // Tamaño del bloque
	S_first_ino         int32 `binary:"little"` // Primer inodo libre
	S_first_blo         int32 `binary:"little"` // Primer bloque libre
	S_bm_inode_start    int32 `binary:"little"` // Inicio del bitmap de inodos
	S_bm_block_start    int32 `binary:"little"` // Inicio del bitmap de bloques
	S_inode_start       int32 `binary:"little"` // Inicio de la tabla de inodos
	S_block_start       int32 `binary:"little"` // Inicio de la tabla de bloques
}

// Constantes actualizadas con sizeof
var (
	SUPERBLOCK_SIZE_ACTUAL = int(unsafe.Sizeof(Superblock{}))
	INODE_SIZE_ACTUAL      = 0 // Se actualizará cuando definamos Inode
	BLOCK_SIZE_ACTUAL      = 64
)

// NewSuperblock crea un nuevo superbloque con valores iniciales
func NewSuperblock(partitionSize int64, blockSize int) *Superblock {
	// Tamaño del inodo se calculará cuando esté definido
	inodeSize := 128 // Placeholder, debe ser sizeof(Inode{})

	// Calcular número de estructuras usando la fórmula:
	// tamaño_particion = superblock + n + 3*n + n*sizeof(inodos) + 3*n*sizeof(block)
	// Simplificando: tamaño_particion = superblock + n*(1 + 3 + sizeof(inodos) + 3*sizeof(block))

	available := partitionSize - int64(SUPERBLOCK_SIZE_ACTUAL)
	denominator := int64(1 + 3 + inodeSize + 3*blockSize)
	n := available / denominator

	if n < MIN_INODES {
		n = MIN_INODES
	}

	inodesCount := int32(n)
	blocksCount := int32(3 * n) // Triple de inodos

	now := time.Now().Unix()

	return &Superblock{
		S_filesystem_type:   EXT2_FILESYSTEM_TYPE,
		S_inodes_count:      inodesCount,
		S_blocks_count:      blocksCount,
		S_free_blocks_count: blocksCount - 2, // -2 por raíz y users.txt
		S_free_inodes_count: inodesCount - 2, // -2 por raíz y users.txt
		S_mtime:             now,
		S_umtime:            0,
		S_mnt_count:         1,
		S_magic:             EXT2_MAGIC,
		S_inode_size:        int32(inodeSize),
		S_block_size:        int32(blockSize),
		S_first_ino:         3, // Primeros 2 ocupados (raíz y users.txt)
		S_first_blo:         3,
		S_bm_inode_start:    int32(SUPERBLOCK_SIZE_ACTUAL),
		S_bm_block_start:    int32(SUPERBLOCK_SIZE_ACTUAL) + inodesCount,
		S_inode_start:       int32(SUPERBLOCK_SIZE_ACTUAL) + inodesCount + blocksCount,
		S_block_start:       int32(SUPERBLOCK_SIZE_ACTUAL) + inodesCount + blocksCount + inodesCount*int32(inodeSize),
	}
}

// SerializeSuperblock convierte Superblock a bytes
func SerializeSuperblock(sb *Superblock) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, sb)
	if err != nil {
		return nil, fmt.Errorf("error al serializar Superblock: %v", err)
	}
	return buf.Bytes(), nil
}

// DeserializeSuperblock convierte bytes a Superblock
func DeserializeSuperblock(data []byte) (*Superblock, error) {
	if len(data) < SUPERBLOCK_SIZE_ACTUAL {
		return nil, fmt.Errorf("datos insuficientes para Superblock: necesarios %d, recibidos %d", SUPERBLOCK_SIZE_ACTUAL, len(data))
	}

	sb := &Superblock{}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, sb)
	if err != nil {
		return nil, fmt.Errorf("error al deserializar Superblock: %v", err)
	}
	return sb, nil
}

// CalculateEXT2Structures calcula el número de inodos y bloques para una partición
func CalculateEXT2Structures(partitionSize int64, blockSize int) (int32, int32) {
	inodeSize := 128 // Placeholder
	available := partitionSize - int64(SUPERBLOCK_SIZE_ACTUAL)
	denominator := int64(1 + 3 + inodeSize + 3*blockSize)
	n := available / denominator

	inodesCount := int32(n)
	blocksCount := int32(3 * n)

	return inodesCount, blocksCount
}

// ValidateEXT2Structures valida que las estructuras EXT2 sean consistentes
func ValidateEXT2Structures(sb *Superblock) error {
	if sb.S_filesystem_type != EXT2_FILESYSTEM_TYPE {
		return fmt.Errorf("tipo de sistema de archivos incorrecto: %d (esperado: %d)", sb.S_filesystem_type, EXT2_FILESYSTEM_TYPE)
	}

	if sb.S_magic != EXT2_MAGIC {
		return fmt.Errorf("número mágico incorrecto: 0x%X (esperado: 0x%X)", sb.S_magic, EXT2_MAGIC)
	}

	if sb.S_inodes_count <= 0 || sb.S_blocks_count <= 0 {
		return fmt.Errorf("número de inodos (%d) o bloques (%d) inválido", sb.S_inodes_count, sb.S_blocks_count)
	}

	if sb.S_blocks_count != 3*sb.S_inodes_count {
		return fmt.Errorf("la relación bloques/inodos debe ser 3:1, encontrado: %d bloques para %d inodos", sb.S_blocks_count, sb.S_inodes_count)
	}

	if sb.S_free_inodes_count < 0 || sb.S_free_inodes_count > sb.S_inodes_count {
		return fmt.Errorf("número de inodos libres inválido: %d (debe estar entre 0 y %d)", sb.S_free_inodes_count, sb.S_inodes_count)
	}

	if sb.S_free_blocks_count < 0 || sb.S_free_blocks_count > sb.S_blocks_count {
		return fmt.Errorf("número de bloques libres inválido: %d (debe estar entre 0 y %d)", sb.S_free_blocks_count, sb.S_blocks_count)
	}

	return nil
}

// GetUsersFileContent genera el contenido inicial del archivo users.txt
func GetUsersFileContent() string {
	return "1,G,root\n1,U,root,root,123\n"
}
