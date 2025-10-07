package ext3

import (
	"encoding/binary"
	"time"
)

// SuperBlock EXT3 con layout correcto según enunciado
type SuperBlock struct {
	// Identificación
	SInodeCount   int32 // n
	SBlockCount   int32 // 3n
	SFreeInodes   int32
	SFreeBlocks   int32
	SCreationTime int64 // timestamp
	SMountTime    int64
	SMagic        int32 // 0xEF53
	SInodeSize    int32
	SBlockSize    int32
	SFirstInode   int32
	SFirstBlock   int32

	// Offsets y tamaños de estructuras
	SBmInodeStart  int64 // Inicio bitmap de inodos
	SBmBlockStart  int64 // Inicio bitmap de bloques
	SInodeStart    int64 // Inicio tabla de inodos
	SBlockStart    int64 // Inicio área de bloques
	SJournalStart  int64 // Inicio del journal
	SJournalCount  int32 // Cantidad de entradas de journal (fijo: 50)

	// Sistema de archivos
	SFsType int32 // 2=EXT2, 3=EXT3
}

// Layout de la partición EXT3:
// 1. SuperBloque (512 bytes)
// 2. Journal (50 * 64 bytes = 3200 bytes)
// 3. Bitmap de Inodos (n bytes o ceil(n/8) si se compacta)
// 4. Bitmap de Bloques (3n bytes o ceil(3n/8) si se compacta)
// 5. Tabla de Inodos (n * 128 bytes)
// 6. Bloques de Datos (3n * blockSize bytes)

// CalcN calcula el número de estructuras según el enunciado
// Ecuación corregida con Journal FIJO de 50 entradas
func CalcN(partSize int64, blockSize int) int64 {
	const (
		superSize    = 512
		journalEntry = 64
		journalFixed = 50 // CONSTANTE según enunciado
		inodeSize    = 128
		bitmapInode  = 1  // 1 byte por inodo (o ceil(n/8) si se compacta a bits)
		bitmapBlock  = 1  // 1 byte por bloque (o ceil(3n/8) si se compacta a bits)
	)

	// Cálculo según enunciado:
	// partSize = superSize + (50 * journalEntry) + n*bitmapInode + 3n*bitmapBlock + n*inodeSize + 3n*blockSize
	//
	// Despejando n:
	// partSize - superSize - 50*journalEntry = n*(bitmapInode + 3*bitmapBlock + inodeSize + 3*blockSize)

	numerator := partSize - superSize - int64(journalFixed*journalEntry)
	denominator := bitmapInode + 3*bitmapBlock + inodeSize + 3*int64(blockSize)

	if denominator <= 0 || numerator <= 0 {
		return 0
	}

	// Aplicar floor(n) según enunciado
	n := numerator / denominator
	return n
}

// CalculateOffsets calcula los offsets de todas las estructuras
func CalculateOffsets(n int64, blockSize int) (sb SuperBlock) {
	const (
		superSize    = 512
		journalEntry = 64
		journalFixed = 50
		inodeSize    = 128
	)

	offset := int64(superSize)

	// 1. SuperBloque (0)
	// Ya está al inicio

	// 2. Journal (fijo: 50 entradas)
	sb.SJournalStart = offset
	sb.SJournalCount = journalFixed
	journalSize := int64(journalFixed * journalEntry)
	offset += journalSize

	// 3. Bitmap de Inodos (n bytes)
	sb.SBmInodeStart = offset
	bmInodeSize := n // 1 byte por inodo
	offset += bmInodeSize

	// 4. Bitmap de Bloques (3n bytes)
	sb.SBmBlockStart = offset
	bmBlockSize := 3 * n // 1 byte por bloque
	offset += bmBlockSize

	// 5. Tabla de Inodos (n * 128)
	sb.SInodeStart = offset
	inodeTableSize := n * inodeSize
	offset += inodeTableSize

	// 6. Bloques de Datos (3n * blockSize)
	sb.SBlockStart = offset

	// Configurar contadores
	sb.SInodeCount = int32(n)
	sb.SBlockCount = int32(3 * n) // Bloques = 3 * inodos
	sb.SFreeInodes = int32(n - 1) // -1 por el root
	sb.SFreeBlocks = int32(3*n - 1) // -1 por el bloque del root

	sb.SInodeSize = inodeSize
	sb.SBlockSize = int32(blockSize)
	sb.SFirstInode = 0 // Root inode
	sb.SFirstBlock = 0 // Root block
	sb.SMagic = 0xEF53
	sb.SFsType = 3 // EXT3
	sb.SCreationTime = time.Now().Unix()
	sb.SMountTime = 0

	return sb
}

// Serialize convierte el SuperBlock a bytes
func (sb *SuperBlock) Serialize() []byte {
	buf := make([]byte, 512)

	binary.LittleEndian.PutUint32(buf[0:], uint32(sb.SInodeCount))
	binary.LittleEndian.PutUint32(buf[4:], uint32(sb.SBlockCount))
	binary.LittleEndian.PutUint32(buf[8:], uint32(sb.SFreeInodes))
	binary.LittleEndian.PutUint32(buf[12:], uint32(sb.SFreeBlocks))
	binary.LittleEndian.PutUint64(buf[16:], uint64(sb.SCreationTime))
	binary.LittleEndian.PutUint64(buf[24:], uint64(sb.SMountTime))
	binary.LittleEndian.PutUint32(buf[32:], uint32(sb.SMagic))
	binary.LittleEndian.PutUint32(buf[36:], uint32(sb.SInodeSize))
	binary.LittleEndian.PutUint32(buf[40:], uint32(sb.SBlockSize))
	binary.LittleEndian.PutUint32(buf[44:], uint32(sb.SFirstInode))
	binary.LittleEndian.PutUint32(buf[48:], uint32(sb.SFirstBlock))

	binary.LittleEndian.PutUint64(buf[52:], uint64(sb.SBmInodeStart))
	binary.LittleEndian.PutUint64(buf[60:], uint64(sb.SBmBlockStart))
	binary.LittleEndian.PutUint64(buf[68:], uint64(sb.SInodeStart))
	binary.LittleEndian.PutUint64(buf[76:], uint64(sb.SBlockStart))
	binary.LittleEndian.PutUint64(buf[84:], uint64(sb.SJournalStart))
	binary.LittleEndian.PutUint32(buf[92:], uint32(sb.SJournalCount))
	binary.LittleEndian.PutUint32(buf[96:], uint32(sb.SFsType))

	return buf
}

// Deserialize lee el SuperBlock desde bytes
func DeserializeSuperBlock(data []byte) SuperBlock {
	var sb SuperBlock

	sb.SInodeCount = int32(binary.LittleEndian.Uint32(data[0:]))
	sb.SBlockCount = int32(binary.LittleEndian.Uint32(data[4:]))
	sb.SFreeInodes = int32(binary.LittleEndian.Uint32(data[8:]))
	sb.SFreeBlocks = int32(binary.LittleEndian.Uint32(data[12:]))
	sb.SCreationTime = int64(binary.LittleEndian.Uint64(data[16:]))
	sb.SMountTime = int64(binary.LittleEndian.Uint64(data[24:]))
	sb.SMagic = int32(binary.LittleEndian.Uint32(data[32:]))
	sb.SInodeSize = int32(binary.LittleEndian.Uint32(data[36:]))
	sb.SBlockSize = int32(binary.LittleEndian.Uint32(data[40:]))
	sb.SFirstInode = int32(binary.LittleEndian.Uint32(data[44:]))
	sb.SFirstBlock = int32(binary.LittleEndian.Uint32(data[48:]))

	sb.SBmInodeStart = int64(binary.LittleEndian.Uint64(data[52:]))
	sb.SBmBlockStart = int64(binary.LittleEndian.Uint64(data[60:]))
	sb.SInodeStart = int64(binary.LittleEndian.Uint64(data[68:]))
	sb.SBlockStart = int64(binary.LittleEndian.Uint64(data[76:]))
	sb.SJournalStart = int64(binary.LittleEndian.Uint64(data[84:]))
	sb.SJournalCount = int32(binary.LittleEndian.Uint32(data[92:]))
	sb.SFsType = int32(binary.LittleEndian.Uint32(data[96:]))

	return sb
}
