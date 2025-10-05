package ext2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

// Content dentro de FolderBlock
type Content struct {
	BName  [12]byte `binary:"little"` // Nombre archivo/carpeta
	BInodo int32    `binary:"little"` // Índice del inodo
}

// FolderBlock representa un bloque de carpeta (directorio)
type FolderBlock struct {
	BContent [4]Content `binary:"little"` // 4 entradas de 16 bytes = 64 bytes total
}

// FileBlock representa un bloque de archivo
type FileBlock struct {
	BContent [64]byte `binary:"little"` // 64 bytes de contenido
}

// PointerBlock representa un bloque de punteros
type PointerBlock struct {
	BPointers [16]int32 `binary:"little"` // 16 punteros de 4 bytes = 64 bytes
}

// NewFolderBlock crea un bloque de carpeta vacío
func NewFolderBlock() *FolderBlock {
	fb := &FolderBlock{}
	// Inicializar todas las entradas con inodo -1 (vacío)
	for i := range fb.BContent {
		fb.BContent[i].BInodo = -1
	}
	return fb
}

// NewFileBlock crea un bloque de archivo vacío
func NewFileBlock() *FileBlock {
	return &FileBlock{}
}

// NewPointerBlock crea un bloque de punteros vacío
func NewPointerBlock() *PointerBlock {
	pb := &PointerBlock{}
	// Inicializar todos los punteros a -1 (vacío)
	for i := range pb.BPointers {
		pb.BPointers[i] = -1
	}
	return pb
}

// SerializeFolderBlock convierte FolderBlock a bytes
func SerializeFolderBlock(fb *FolderBlock) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, fb)
	if err != nil {
		return nil, fmt.Errorf("error al serializar FolderBlock: %v", err)
	}
	return buf.Bytes(), nil
}

// DeserializeFolderBlock convierte bytes a FolderBlock
func DeserializeFolderBlock(data []byte) (*FolderBlock, error) {
	blockSize := int(unsafe.Sizeof(FolderBlock{}))
	if len(data) < blockSize {
		return nil, fmt.Errorf("datos insuficientes para FolderBlock: necesarios %d, recibidos %d", blockSize, len(data))
	}

	fb := &FolderBlock{}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, fb)
	if err != nil {
		return nil, fmt.Errorf("error al deserializar FolderBlock: %v", err)
	}
	return fb, nil
}

// SerializeFileBlock convierte FileBlock a bytes
func SerializeFileBlock(fb *FileBlock) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, fb)
	if err != nil {
		return nil, fmt.Errorf("error al serializar FileBlock: %v", err)
	}
	return buf.Bytes(), nil
}

// DeserializeFileBlock convierte bytes a FileBlock
func DeserializeFileBlock(data []byte) (*FileBlock, error) {
	blockSize := int(unsafe.Sizeof(FileBlock{}))
	if len(data) < blockSize {
		return nil, fmt.Errorf("datos insuficientes para FileBlock: necesarios %d, recibidos %d", blockSize, len(data))
	}

	fb := &FileBlock{}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, fb)
	if err != nil {
		return nil, fmt.Errorf("error al deserializar FileBlock: %v", err)
	}
	return fb, nil
}

// SerializePointerBlock convierte PointerBlock a bytes
func SerializePointerBlock(pb *PointerBlock) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, pb)
	if err != nil {
		return nil, fmt.Errorf("error al serializar PointerBlock: %v", err)
	}
	return buf.Bytes(), nil
}

// DeserializePointerBlock convierte bytes a PointerBlock
func DeserializePointerBlock(data []byte) (*PointerBlock, error) {
	blockSize := int(unsafe.Sizeof(PointerBlock{}))
	if len(data) < blockSize {
		return nil, fmt.Errorf("datos insuficientes para PointerBlock: necesarios %d, recibidos %d", blockSize, len(data))
	}

	pb := &PointerBlock{}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, pb)
	if err != nil {
		return nil, fmt.Errorf("error al deserializar PointerBlock: %v", err)
	}
	return pb, nil
}

// AddEntry agrega una entrada al FolderBlock
func (fb *FolderBlock) AddEntry(name string, inodeIndex int32) error {
	for i := range fb.BContent {
		if fb.BContent[i].BInodo == -1 {
			// Convertir nombre a [12]byte
			var nameBytes [12]byte
			copy(nameBytes[:], name)
			fb.BContent[i].BName = nameBytes
			fb.BContent[i].BInodo = inodeIndex
			return nil
		}
	}
	return fmt.Errorf("folder block lleno, no se puede agregar más entradas")
}

// GetEntries retorna todas las entradas válidas del FolderBlock
func (fb *FolderBlock) GetEntries() []Content {
	var entries []Content
	for i := range fb.BContent {
		if fb.BContent[i].BInodo != -1 {
			entries = append(entries, fb.BContent[i])
		}
	}
	return entries
}

// FindEntry busca una entrada por nombre
func (fb *FolderBlock) FindEntry(name string) (int32, bool) {
	var nameBytes [12]byte
	copy(nameBytes[:], name)

	for i := range fb.BContent {
		if fb.BContent[i].BInodo != -1 && fb.BContent[i].BName == nameBytes {
			return fb.BContent[i].BInodo, true
		}
	}
	return -1, false
}

// RemoveEntry elimina una entrada del FolderBlock
func (fb *FolderBlock) RemoveEntry(name string) bool {
	var nameBytes [12]byte
	copy(nameBytes[:], name)

	for i := range fb.BContent {
		if fb.BContent[i].BInodo != -1 && fb.BContent[i].BName == nameBytes {
			fb.BContent[i].BInodo = -1
			fb.BContent[i].BName = [12]byte{}
			return true
		}
	}
	return false
}

// IsFull verifica si el FolderBlock está lleno
func (fb *FolderBlock) IsFull() bool {
	for i := range fb.BContent {
		if fb.BContent[i].BInodo == -1 {
			return false
		}
	}
	return true
}

// GetName convierte el nombre de bytes a string
func (c *Content) GetName() string {
	// Encontrar el primer byte 0 o recorrer todo
	for i, b := range c.BName {
		if b == 0 {
			return string(c.BName[:i])
		}
	}
	return string(c.BName[:])
}
