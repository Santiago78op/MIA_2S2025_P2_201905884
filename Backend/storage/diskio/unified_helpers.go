package diskio

import (
	"Backend/core/models"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// ================================================================
// ESTRUCTURAS UNIFICADAS PARA EXT2/EXT3
// ================================================================

// SuperBlockUnified representa un superblock que puede ser EXT2 o EXT3
type SuperBlockUnified struct {
	Ext2   models.SuperBlock
	Ext3   models.SuperBlockExt3
	IsExt3 bool
}

// InodeUnified representa un inodo que puede ser de EXT2 o EXT3
type InodeUnified struct {
	Ext2   models.Inode
	IsExt3 bool
	Index  int32 // Índice del inodo en la tabla
}

// BlockUnified representa un bloque que puede ser de directorio o archivo
type BlockUnified struct {
	FolderBlock models.FolderBlock
	FileBlock   models.FileBlock
	IsFolder    bool
}

// DirEntry representa una entrada de directorio
type DirEntry struct {
	Name     string
	InodeIdx int32
}

// ================================================================
// MÉTODOS DE InodeUnified
// ================================================================

func (i InodeUnified) IsDir() bool {
	return i.Ext2.IType == models.FileTypeFolder
}

func (i InodeUnified) UID() int32 {
	return i.Ext2.IUid
}

func (i InodeUnified) GID() int32 {
	return i.Ext2.IGid
}

func (i InodeUnified) Perm() [3]byte {
	return i.Ext2.IPerm
}

func (i InodeUnified) Size() int32 {
	return i.Ext2.ISize
}

func (i InodeUnified) Block(idx int) int32 {
	if idx < 0 || idx >= 12 {
		return -1
	}
	return i.Ext2.IBlock[idx]
}

func (i *InodeUnified) SetBlock(idx int, blkIdx int32) {
	if idx >= 0 && idx < 12 {
		i.Ext2.IBlock[idx] = blkIdx
	}
}

func (i *InodeUnified) SetMtime(ts int64) {
	i.Ext2.IMtime = ts
}

// ================================================================
// MÉTODOS DE BlockUnified
// ================================================================

func (b *BlockUnified) SetEntry(idx int, name string, inodeIdx int32) {
	if !b.IsFolder || idx < 0 || idx >= models.DirEntriesPerBlk {
		return
	}
	copy(b.FolderBlock.Content[idx].BName[:], name)
	b.FolderBlock.Content[idx].BInodo = inodeIdx
}

// ================================================================
// FUNCIONES DE LECTURA/ESCRITURA UNIFICADAS
// ================================================================

// readAnySuperblock lee el superblock detectando automáticamente EXT2 o EXT3
func (r *FileFsRepository) readAnySuperblock(f *os.File, region Region) (SuperBlockUnified, bool, error) {
	// Intentar leer como EXT3 primero
	sb3, err := r.readSuperExt3(f, region.Start)
	if err == nil && sb3.FsType == 3 {
		return SuperBlockUnified{Ext3: sb3, IsExt3: true}, true, nil
	}

	// Intentar leer como EXT2
	sb2, err := readSuperBlock(f, region.Start)
	if err != nil {
		return SuperBlockUnified{}, false, err
	}

	return SuperBlockUnified{Ext2: sb2, IsExt3: false}, false, nil
}

// writeSuperblock escribe el superblock apropiado
func (r *FileFsRepository) writeSuperblock(f *os.File, sb SuperBlockUnified, region Region) error {
	if sb.IsExt3 {
		return r.writeSuperExt3(f, sb.Ext3)
	}
	return writeSuperBlock(f, region.Start, &sb.Ext2)
}

// readBitmapFromSB lee el bitmap de inodos o bloques según el superblock
func (r *FileFsRepository) readBitmapFromSB(f *os.File, sb SuperBlockUnified, isInode bool, region Region) ([]byte, error) {
	if sb.IsExt3 {
		var start int64
		var size int
		if isInode {
			start = sb.Ext3.BmInodeStart
			size = int(sb.Ext3.InodesCount)
		} else {
			start = sb.Ext3.BmBlockStart
			size = int(sb.Ext3.BlocksCount)
		}
		return readBitmap(f, start, size, region)
	}

	// EXT2
	var start int64
	var size int
	if isInode {
		start = sb.Ext2.SBmInodeStart
		size = int(sb.Ext2.SInodesCount)
	} else {
		start = sb.Ext2.SBmBlockStart
		size = int(sb.Ext2.SBlocksCount)
	}
	return readBitmap(f, start, size, region)
}

// writeBitmapToSB escribe el bitmap de inodos o bloques según el superblock
func (r *FileFsRepository) writeBitmapToSB(f *os.File, sb SuperBlockUnified, bm []byte, isInode bool, region Region) error {
	if sb.IsExt3 {
		var start int64
		if isInode {
			start = sb.Ext3.BmInodeStart
		} else {
			start = sb.Ext3.BmBlockStart
		}
		_, err := SafeWriteAt(f, bm, start, region)
		return err
	}

	// EXT2
	var start int64
	if isInode {
		start = sb.Ext2.SBmInodeStart
	} else {
		start = sb.Ext2.SBmBlockStart
	}
	_, err := SafeWriteAt(f, bm, start, region)
	return err
}

// readInodeByIndex lee un inodo por su índice
func (r *FileFsRepository) readInodeByIndex(f *os.File, sb SuperBlockUnified, idx int32, region Region) (InodeUnified, error) {
	var start int64
	if sb.IsExt3 {
		start = sb.Ext3.InodeStart
	} else {
		start = sb.Ext2.SInodeStart
	}

	ino, err := readInodeAt(f, start, int(idx), region)
	if err != nil {
		return InodeUnified{}, err
	}

	return InodeUnified{Ext2: ino, IsExt3: sb.IsExt3, Index: idx}, nil
}

// writeInodeToSB escribe un inodo
func (r *FileFsRepository) writeInodeToSB(f *os.File, sb SuperBlockUnified, ino InodeUnified, region Region) error {
	var start int64
	if sb.IsExt3 {
		start = sb.Ext3.InodeStart
	} else {
		start = sb.Ext2.SInodeStart
	}

	return writeInodeAt(f, start, int(ino.Index), &ino.Ext2, region)
}

// readBlockToSB lee un bloque por su índice
func (r *FileFsRepository) readBlockByIndex(f *os.File, sb SuperBlockUnified, idx int32, isFolder bool, region Region) (BlockUnified, error) {
	var start int64
	if sb.IsExt3 {
		start = sb.Ext3.BlockStart
	} else {
		start = sb.Ext2.SBlockStart
	}

	offset := start + int64(idx*64)
	if _, err := f.Seek(offset, 0); err != nil {
		return BlockUnified{}, err
	}

	if isFolder {
		var fb models.FolderBlock
		if err := binary.Read(f, binary.LittleEndian, &fb); err != nil {
			return BlockUnified{}, err
		}
		return BlockUnified{FolderBlock: fb, IsFolder: true}, nil
	}

	// Archivo
	var fileBlock models.FileBlock
	if err := binary.Read(f, binary.LittleEndian, &fileBlock); err != nil {
		return BlockUnified{}, err
	}
	return BlockUnified{FileBlock: fileBlock, IsFolder: false}, nil
}

// writeBlockToSB escribe un bloque (puede ser []byte para archivos o BlockUnified para directorios)
func (r *FileFsRepository) writeBlockToSB(f *os.File, sb SuperBlockUnified, idx int32, data interface{}, region Region) error {
	var start int64
	if sb.IsExt3 {
		start = sb.Ext3.BlockStart
	} else {
		start = sb.Ext2.SBlockStart
	}

	offset := start + int64(idx*64)
	if _, err := f.Seek(offset, 0); err != nil {
		return err
	}

	switch v := data.(type) {
	case BlockUnified:
		if v.IsFolder {
			return binary.Write(f, binary.LittleEndian, &v.FolderBlock)
		}
		return binary.Write(f, binary.LittleEndian, &v.FileBlock)
	case []byte:
		// Para archivos: escribir bytes directamente (hasta BlockSizeFile bytes)
		block := make([]byte, models.BlockSizeFile)
		copy(block, v)
		_, err := f.Write(block)
		return err
	default:
		return fmt.Errorf("tipo de bloque no soportado")
	}
}

// walkToNode navega un path y devuelve el inodo final, índice del padre y nombre del último componente
func (r *FileFsRepository) walkToNode(f *os.File, sb SuperBlockUnified, path []string, region Region) (InodeUnified, int32, string, error) {
	if len(path) == 0 {
		// Retornar raíz
		root, err := r.readInodeByIndex(f, sb, 0, region)
		return root, -1, "", err
	}

	currentIdx := int32(0)
	parentIdx := int32(-1)
	lastName := ""

	for i, name := range path {
		lastName = name
		currentIno, err := r.readInodeByIndex(f, sb, currentIdx, region)
		if err != nil {
			return InodeUnified{}, -1, "", err
		}

		if !currentIno.IsDir() {
			return InodeUnified{}, -1, "", fmt.Errorf("'%s' no es un directorio", strings.Join(path[:i], "/"))
		}

		// Buscar entrada
		entries, err := r.readDirEntriesFromInode(f, sb, currentIno, region)
		if err != nil {
			return InodeUnified{}, -1, "", err
		}

		found := false
		for _, e := range entries {
			if e.Name == name {
				parentIdx = currentIdx
				currentIdx = e.InodeIdx
				found = true
				break
			}
		}

		if !found {
			return InodeUnified{}, -1, "", fmt.Errorf("no existe: %s", strings.Join(path[:i+1], "/"))
		}
	}

	finalIno, err := r.readInodeByIndex(f, sb, currentIdx, region)
	return finalIno, parentIdx, lastName, err
}

// readDirEntriesFromInode lee todas las entradas de un directorio
func (r *FileFsRepository) readDirEntriesFromInode(f *os.File, sb SuperBlockUnified, ino InodeUnified, region Region) ([]DirEntry, error) {
	if !ino.IsDir() {
		return nil, fmt.Errorf("no es un directorio")
	}

	var entries []DirEntry

	// Leer bloques directos
	for i := 0; i < 12; i++ {
		blkIdx := ino.Block(i)
		if blkIdx == -1 {
			break
		}

		block, err := r.readBlockByIndex(f, sb, blkIdx, true, region)
		if err != nil {
			return nil, err
		}

		for _, e := range block.FolderBlock.Content {
			if e.BInodo != -1 {
				name := strings.TrimRight(string(e.BName[:]), "\x00")
				entries = append(entries, DirEntry{Name: name, InodeIdx: e.BInodo})
			}
		}
	}

	return entries, nil
}

// readFileContentFromInode lee el contenido completo de un archivo
func (r *FileFsRepository) readFileContentFromInode(f *os.File, sb SuperBlockUnified, ino InodeUnified, region Region) ([]byte, error) {
	if ino.IsDir() {
		return nil, fmt.Errorf("es un directorio, no un archivo")
	}

	content := make([]byte, 0, ino.Size())

	for i := 0; i < 12; i++ {
		blkIdx := ino.Block(i)
		if blkIdx == -1 {
			break
		}

		block, err := r.readBlockByIndex(f, sb, blkIdx, false, region)
		if err != nil {
			return nil, err
		}

		// Agregar contenido del bloque
		remaining := int(ino.Size()) - len(content)
		if remaining > 64 {
			content = append(content, block.FileBlock.Content[:]...)
		} else {
			content = append(content, block.FileBlock.Content[:remaining]...)
			break
		}
	}

	return content, nil
}

// addEntryToDirectory agrega una entrada a un directorio existente
func (r *FileFsRepository) addEntryToDirectory(f *os.File, sb SuperBlockUnified, dirIno InodeUnified, name string, inodeIdx int32, bmBlock []byte, region Region) error {
	if !dirIno.IsDir() {
		return fmt.Errorf("no es un directorio")
	}

	// Buscar slot libre en bloques existentes
	for i := 0; i < 12; i++ {
		blkIdx := dirIno.Block(i)
		if blkIdx == -1 {
			// Necesitamos reservar un nuevo bloque
			newBlkIdx := r.allocateBlock(bmBlock)
			if newBlkIdx < 0 {
				return fmt.Errorf("no hay bloques libres para expandir directorio")
			}

			// Crear nuevo bloque de directorio vacío
			dirBlock := createEmptyDirBlock()
			dirBlock.SetEntry(0, name, inodeIdx)

			// Escribir bloque
			if err := r.writeBlockToSB(f, sb, newBlkIdx, dirBlock, region); err != nil {
				return err
			}

			// Actualizar inodo del directorio
			dirIno.SetBlock(i, newBlkIdx)
			if err := r.writeInodeToSB(f, sb, dirIno, region); err != nil {
				return err
			}

			return nil
		}

		// Leer bloque existente
		block, err := r.readBlockByIndex(f, sb, blkIdx, true, region)
		if err != nil {
			return err
		}

		// Buscar slot libre
		for j := range block.FolderBlock.Content {
			if block.FolderBlock.Content[j].BInodo == -1 {
				block.SetEntry(j, name, inodeIdx)

				// Escribir bloque actualizado
				if err := r.writeBlockToSB(f, sb, blkIdx, block, region); err != nil {
					return err
				}
				return nil
			}
		}
	}

	return fmt.Errorf("directorio lleno (no hay slots libres)")
}

// updateSuperblockCounters actualiza los contadores de inodos/bloques libres
func (r *FileFsRepository) updateSuperblockCounters(sb SuperBlockUnified, bmInode, bmBlock []byte) {
	freeInodes := int32(0)
	for _, b := range bmInode {
		if b == 0 {
			freeInodes++
		}
	}

	freeBlocks := int32(0)
	for _, b := range bmBlock {
		if b == 0 {
			freeBlocks++
		}
	}

	if sb.IsExt3 {
		sb.Ext3.FreeInodes = freeInodes
		sb.Ext3.FreeBlocks = freeBlocks
	} else {
		sb.Ext2.SFreeInodesCount = freeInodes
		sb.Ext2.SFreeBlocksCount = freeBlocks
	}
}
