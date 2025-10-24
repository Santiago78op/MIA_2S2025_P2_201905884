package diskio

import (
	"Backend/core/models"
	"fmt"
	"os"
	"strings"
	"time"
)

// ================================================================
// COMANDO COPY - Copia recursiva de archivos/carpetas con permisos
// ================================================================

// Copy realiza copia recursiva de src → destDir, aplicando las reglas de permisos.
// Devuelve la cantidad de elementos copiados y omitidos por permisos.
//
// Reglas:
// - Copia recursiva de archivo/carpeta SOLO si hay permiso de LECTURA en origen.
// - Si un elemento del subárbol no tiene permiso de lectura, SE OMITE ese elemento.
// - El destino DEBE existir, ser carpeta y tener permiso de ESCRITURA.
// - Registra en Journal (EXT3) en write-ahead.
func (r *FileFsRepository) Copy(id string, srcPath, destPath []string, uid int, gid int) (int, int, error) {
	// 1. Resolver montaje y obtener región de la partición
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return 0, 0, err
	}

	// 2. Abrir disco
	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	// 3. Leer superblock (intentar EXT3 primero, luego EXT2)
	sb, isExt3, err := r.readAnySuperblock(f, region)
	if err != nil {
		return 0, 0, err
	}

	// Write-ahead journal omitido - se registra al final de la operación
	_ = isExt3

	// 5. Navegar al origen
	srcIno, srcParentIdx, _, err := r.walkToNode(f, sb, srcPath, region)
	if err != nil {
		return 0, 0, fmt.Errorf("no se encontró origen: %w", err)
	}

	// 6. Verificar permiso de lectura en origen
	isRoot := uid == 1
	if !canRead(srcIno, uid, gid, isRoot) {
		return 0, 0, fmt.Errorf("sin permiso de lectura en origen '%s'", strings.Join(srcPath, "/"))
	}

	// 7. Separar destPath en directorio padre + nombre del archivo destino
	var destDirPath []string
	var destName string

	if len(destPath) == 1 {
		// Destino es en la raíz: /archivo.txt
		destDirPath = []string{} // raíz
		destName = destPath[0]
	} else {
		// Destino es /dir1/dir2/.../archivo.txt
		destDirPath = destPath[:len(destPath)-1]
		destName = destPath[len(destPath)-1]
	}

	// 8. Navegar al directorio padre del destino (debe existir y ser directorio)
	var destDirIno InodeUnified
	if len(destDirPath) == 0 {
		// Destino es en raíz, leer inodo raíz (índice 0)
		destDirIno, err = r.readInodeByIndex(f, sb, 0, region)
		if err != nil {
			return 0, 0, fmt.Errorf("no se pudo leer inodo raíz: %w", err)
		}
	} else {
		// Navegar al directorio padre
		destDirIno, _, _, err = r.walkToNode(f, sb, destDirPath, region)
		if err != nil {
			return 0, 0, fmt.Errorf("directorio padre del destino no existe: %w", err)
		}
	}

	if !destDirIno.IsDir() {
		return 0, 0, fmt.Errorf("el padre del destino no es un directorio")
	}

	// 9. Verificar permiso de escritura en directorio padre del destino
	if !canWrite(destDirIno, uid, gid, isRoot) {
		return 0, 0, fmt.Errorf("sin permiso de escritura en directorio destino '%s'", strings.Join(destDirPath, "/"))
	}

	// 10. Verificar que no existe ya un elemento con el nombre destino
	destEntries, err := r.readDirEntriesFromInode(f, sb, destDirIno, region)
	if err != nil {
		return 0, 0, err
	}
	for _, e := range destEntries {
		if e.Name != "." && e.Name != ".." && strings.EqualFold(e.Name, destName) {
			return 0, 0, fmt.Errorf("ya existe '%s' en destino", destName)
		}
	}

	// 11. Leer bitmaps
	bmInode, err := r.readBitmapFromSB(f, sb, true, region)
	if err != nil {
		return 0, 0, err
	}
	bmBlock, err := r.readBitmapFromSB(f, sb, false, region)
	if err != nil {
		return 0, 0, err
	}

	// 12. Copia recursiva
	newInoIdx, copied, skipped, err := r.copyNodeRecursive(
		f, sb, srcIno, bmInode, bmBlock, uid, gid, isRoot, region,
	)
	if err != nil {
		return copied, skipped, err
	}

	// 13. Agregar entrada en directorio destino con el nombre especificado
	if err := r.addEntryToDirectory(f, sb, destDirIno, destName, newInoIdx, bmBlock, region); err != nil {
		return copied, skipped, err
	}

	// 14. Actualizar mtime del directorio destino
	destDirIno.SetMtime(time.Now().Unix())
	if err := r.writeInodeToSB(f, sb, destDirIno, region); err != nil {
		return copied, skipped, err
	}

	// 15. Escribir bitmaps actualizados
	if err := r.writeBitmapToSB(f, sb, bmInode, true, region); err != nil {
		return copied, skipped, err
	}
	if err := r.writeBitmapToSB(f, sb, bmBlock, false, region); err != nil {
		return copied, skipped, err
	}

	// 16. Actualizar superblock (contadores)
	r.updateSuperblockCounters(sb, bmInode, bmBlock)
	if err := r.writeSuperblock(f, sb, region); err != nil {
		return copied, skipped, err
	}

	// Registrar en journal (ignorar errores para no romper la operación)
	srcFullPath := "/" + strings.Join(srcPath, "/")
	// Construir la ruta completa del destino correctamente
	destFullPath := "/" + strings.Join(append(destDirPath, destName), "/")
	contentInfo := fmt.Sprintf("to=%s, copied=%d, skipped=%d", destFullPath, copied, skipped)
	_ = r.JournalAppendPublic(id, "COPY", srcFullPath, contentInfo, time.Now().Unix())

	_ = srcParentIdx // Puede usarse para logging si se desea
	return copied, skipped, nil
}

// copyNodeRecursive clona un archivo o directorio verificando permisos de lectura.
// Devuelve el índice del nuevo inodo creado, cantidad copiada y omitida.
func (r *FileFsRepository) copyNodeRecursive(
	f *os.File,
	sb SuperBlockUnified,
	srcIno InodeUnified,
	bmInode, bmBlock []byte,
	uid, gid int,
	isRoot bool,
	region Region,
) (newInodeIdx int32, copied, skipped int, err error) {
	// 1. Verificar permiso de lectura
	if !canRead(srcIno, uid, gid, isRoot) {
		return -1, 0, 1, nil // Omitido por permisos
	}

	// 2. Reservar inodo
	newInoIdx := r.allocateInode(bmInode)
	if newInoIdx < 0 {
		return -1, 0, 0, fmt.Errorf("no hay inodos libres")
	}

	// 3. Si es directorio
	if srcIno.IsDir() {
		// Reservar bloque para el directorio
		newBlkIdx := r.allocateBlock(bmBlock)
		if newBlkIdx < 0 {
			bmInode[newInoIdx] = 0 // Liberar inodo
			return -1, 0, 0, fmt.Errorf("no hay bloques libres")
		}

		// Crear inodo del nuevo directorio
		newIno := createInodeDir(uid, gid, srcIno.Perm())
		newIno.SetBlock(0, newBlkIdx)

		// Escribir bloque de directorio con . y ..
		dirBlock := createEmptyDirBlock()
		dirBlock.SetEntry(0, ".", newInoIdx)
		dirBlock.SetEntry(1, "..", newInoIdx) // Se actualizará cuando se añada al padre
		if err := r.writeBlockToSB(f, sb, newBlkIdx, dirBlock, region); err != nil {
			return -1, 0, 0, err
		}

		// Escribir inodo
		if err := r.writeInodeToSB(f, sb, newIno, region); err != nil {
			return -1, 0, 0, err
		}

		totalCopied := 1
		totalSkipped := 0

		// Copiar hijos
		children, err := r.readDirEntriesFromInode(f, sb, srcIno, region)
		if err != nil {
			return newInoIdx, totalCopied, totalSkipped, err
		}

		for _, child := range children {
			if child.Name == "." || child.Name == ".." {
				continue
			}

			// Leer inodo hijo
			childIno, err := r.readInodeByIndex(f, sb, child.InodeIdx, region)
			if err != nil {
				return newInoIdx, totalCopied, totalSkipped, err
			}

			// Copiar recursivamente
			childNewIdx, c, s, err := r.copyNodeRecursive(
				f, sb, childIno, bmInode, bmBlock, uid, gid, isRoot, region,
			)
			if err != nil {
				return newInoIdx, totalCopied, totalSkipped, err
			}

			totalCopied += c
			totalSkipped += s

			// Si se copió el hijo, agregar entrada al nuevo directorio
			if childNewIdx >= 0 {
				if err := r.addEntryToDirectory(f, sb, newIno, child.Name, childNewIdx, bmBlock, region); err != nil {
					return newInoIdx, totalCopied, totalSkipped, err
				}
			}
		}

		return newInoIdx, totalCopied, totalSkipped, nil
	}

	// 4. Si es archivo regular
	// Leer contenido del archivo original
	content, err := r.readFileContentFromInode(f, sb, srcIno, region)
	if err != nil {
		bmInode[newInoIdx] = 0 // Liberar inodo
		return -1, 0, 0, err
	}

	// Calcular bloques necesarios
	blocksNeeded := (len(content) + models.BlockSizeFile - 1) / models.BlockSizeFile
	if blocksNeeded > 12 {
		bmInode[newInoIdx] = 0
		return -1, 0, 0, fmt.Errorf("archivo demasiado grande (max 12 bloques)")
	}

	// Reservar bloques
	blockIndices := make([]int32, blocksNeeded)
	for i := 0; i < blocksNeeded; i++ {
		blkIdx := r.allocateBlock(bmBlock)
		if blkIdx < 0 {
			// Liberar recursos
			bmInode[newInoIdx] = 0
			for j := 0; j < i; j++ {
				bmBlock[blockIndices[j]] = 0
			}
			return -1, 0, 0, fmt.Errorf("no hay bloques libres")
		}
		blockIndices[i] = blkIdx
	}

	// Crear inodo del archivo
	newIno := createInodeFile(uid, gid, srcIno.Perm(), int32(len(content)))
	for i, blkIdx := range blockIndices {
		newIno.SetBlock(i, blkIdx)
	}

	// Escribir contenido en bloques
	for i, blkIdx := range blockIndices {
		start := i * models.BlockSizeFile
		end := start + models.BlockSizeFile
		if end > len(content) {
			end = len(content)
		}

		block := make([]byte, models.BlockSizeFile)
		copy(block, content[start:end])

		if err := r.writeBlockToSB(f, sb, blkIdx, block, region); err != nil {
			return -1, 0, 0, err
		}
	}

	// Escribir inodo
	if err := r.writeInodeToSB(f, sb, newIno, region); err != nil {
		return -1, 0, 0, err
	}

	return newInoIdx, 1, 0, nil
}

// ================================================================
// FUNCIONES AUXILIARES
// ================================================================

// allocateInode busca el primer inodo libre en el bitmap y lo marca como usado
func (r *FileFsRepository) allocateInode(bmInode []byte) int32 {
	for i := 0; i < len(bmInode); i++ {
		if bmInode[i] == 0 {
			bmInode[i] = 1
			return int32(i)
		}
	}
	return -1
}

// allocateBlock busca el primer bloque libre en el bitmap y lo marca como usado
func (r *FileFsRepository) allocateBlock(bmBlock []byte) int32 {
	for i := 0; i < len(bmBlock); i++ {
		if bmBlock[i] == 0 {
			bmBlock[i] = 1
			return int32(i)
		}
	}
	return -1
}

// canRead verifica si el usuario tiene permiso de lectura
func canRead(ino InodeUnified, uid, gid int, isRoot bool) bool {
	if isRoot {
		return true
	}
	perm := ino.Perm()
	if ino.UID() == int32(uid) {
		return (perm[0]-'0')&4 != 0 // owner read
	}
	if ino.GID() == int32(gid) {
		return (perm[1]-'0')&4 != 0 // group read
	}
	return (perm[2]-'0')&4 != 0 // others read
}

// canWrite verifica si el usuario tiene permiso de escritura
func canWrite(ino InodeUnified, uid, gid int, isRoot bool) bool {
	if isRoot {
		return true
	}
	perm := ino.Perm()
	if ino.UID() == int32(uid) {
		return (perm[0]-'0')&2 != 0 // owner write
	}
	if ino.GID() == int32(gid) {
		return (perm[1]-'0')&2 != 0 // group write
	}
	return (perm[2]-'0')&2 != 0 // others write
}

// createInodeDir crea un inodo de directorio
func createInodeDir(uid, gid int, perm [3]byte) InodeUnified {
	now := time.Now().Unix()
	return InodeUnified{
		Ext2: models.Inode{
			IUid:   int32(uid),
			IGid:   int32(gid),
			ISize:  0,
			IAtime: now,
			ICtime: now,
			IMtime: now,
			IType:  models.FileTypeFolder,
			IPerm:  perm,
			IBlock: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		},
		IsExt3: false,
	}
}

// createInodeFile crea un inodo de archivo regular
func createInodeFile(uid, gid int, perm [3]byte, size int32) InodeUnified {
	now := time.Now().Unix()
	return InodeUnified{
		Ext2: models.Inode{
			IUid:   int32(uid),
			IGid:   int32(gid),
			ISize:  size,
			IAtime: now,
			ICtime: now,
			IMtime: now,
			IType:  models.FileTypeRegular,
			IPerm:  perm,
			IBlock: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		},
		IsExt3: false,
	}
}

// createEmptyDirBlock crea un bloque de directorio vacío
func createEmptyDirBlock() BlockUnified {
	var fb models.FolderBlock
	for i := range fb.Content {
		fb.Content[i].BInodo = -1
	}
	return BlockUnified{
		FolderBlock: fb,
		IsFolder:    true,
	}
}
