package ext2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"
)

// Inode representa un inodo en el sistema de archivos EXT2
type Inode struct {
	IUid   int32     `binary:"little"` // UID propietario
	IGid   int32     `binary:"little"` // GID grupo
	IS     int32     `binary:"little"` // Tamaño en bytes
	IAtime int64     `binary:"little"` // Último acceso
	ICtime int64     `binary:"little"` // Creación
	IMtime int64     `binary:"little"` // Modificación
	IBlock [15]int32 `binary:"little"` // 12 directos + 3 indirectos
	IType  byte      `binary:"little"` // 0=carpeta, 1=archivo
	IPerm  [3]byte   `binary:"little"` // Permisos UGO
}

// Constantes para tipos de inodos
const (
	INODE_TYPE_FOLDER = 0
	INODE_TYPE_FILE   = 1
)

// NewInode crea un nuevo inodo con valores iniciales
func NewInode(uid, gid int32, inodeType byte, perm [3]byte) *Inode {
	inode := &Inode{
		IUid:   uid,
		IGid:   gid,
		IS:     0,
		IAtime: time.Now().Unix(),
		ICtime: time.Now().Unix(),
		IMtime: time.Now().Unix(),
		IType:  inodeType,
		IPerm:  perm,
	}

	// Inicializar todos los bloques a -1 (no usados)
	for i := range inode.IBlock {
		inode.IBlock[i] = -1
	}

	return inode
}

// NewFolderInode crea un inodo para carpeta
func NewFolderInode(uid, gid int32) *Inode {
	return NewInode(uid, gid, INODE_TYPE_FOLDER, [3]byte{'7', '5', '5'}) // rwxr-xr-x
}

// NewFileInode crea un inodo para archivo
func NewFileInode(uid, gid int32) *Inode {
	return NewInode(uid, gid, INODE_TYPE_FILE, [3]byte{'6', '6', '4'}) // rw-rw-r--
}

// SerializeInode convierte Inode a bytes
func SerializeInode(inode *Inode) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, inode)
	if err != nil {
		return nil, fmt.Errorf("error al serializar Inode: %v", err)
	}
	return buf.Bytes(), nil
}

// DeserializeInode convierte bytes a Inode
func DeserializeInode(data []byte) (*Inode, error) {
	inodeSize := int(unsafe.Sizeof(Inode{}))
	if len(data) < inodeSize {
		return nil, fmt.Errorf("datos insuficientes para Inode: necesarios %d, recibidos %d", inodeSize, len(data))
	}

	inode := &Inode{}
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, inode)
	if err != nil {
		return nil, fmt.Errorf("error al deserializar Inode: %v", err)
	}
	return inode, nil
}

// IsFolder verifica si el inodo es una carpeta
func (i *Inode) IsFolder() bool {
	return i.IType == INODE_TYPE_FOLDER
}

// IsFile verifica si el inodo es un archivo
func (i *Inode) IsFile() bool {
	return i.IType == INODE_TYPE_FILE
}

// GetDirectBlocks retorna los bloques directos usados
func (i *Inode) GetDirectBlocks() []int32 {
	var blocks []int32
	for j := 0; j < 12; j++ {
		if i.IBlock[j] != -1 {
			blocks = append(blocks, i.IBlock[j])
		}
	}
	return blocks
}

// GetIndirectBlocks retorna los bloques indirectos usados
func (i *Inode) GetIndirectBlocks() []int32 {
	var blocks []int32
	for j := 12; j < 15; j++ {
		if i.IBlock[j] != -1 {
			blocks = append(blocks, i.IBlock[j])
		}
	}
	return blocks
}

// SetBlock asigna un bloque en la primera posición libre
func (i *Inode) SetBlock(blockIndex int32) error {
	for j := 0; j < 15; j++ {
		if i.IBlock[j] == -1 {
			i.IBlock[j] = blockIndex
			return nil
		}
	}
	return fmt.Errorf("no hay espacio para más bloques en el inodo")
}
