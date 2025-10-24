package diskio

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ================================================================
// COMANDO MOVE - Mueve archivos/carpetas mediante re-enlace
// ================================================================

// Move realiza movimiento de src → destDir mediante re-enlace (sin copiar bloques).
// Solo funciona dentro de la misma partición (mismo ID).
//
// Reglas:
// - Mismo id ⇒ misma partición: solo re-enlace (NO copia de bloques).
// - Requiere permiso de ESCRITURA en el directorio origen (para eliminar entrada).
// - Requiere permiso de ESCRITURA en el directorio destino (para agregar entrada).
// - Si hay homónimo en destino, error (no sobrescribe).
// - Registra en Journal (EXT3) en write-ahead.
func (r *FileFsRepository) Move(id string, srcPath, destPath []string, uid int, gid int) error {
	// 1. Resolver montaje y obtener región de la partición
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	// 2. Abrir disco
	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// 3. Leer superblock (intentar EXT3 primero, luego EXT2)
	sb, isExt3, err := r.readAnySuperblock(f, region)
	if err != nil {
		return err
	}

	// Write-ahead journal omitido - se registra al final de la operación
	_ = isExt3

	// 5. Validar que origen y destino no sean vacíos
	if len(srcPath) == 0 {
		return fmt.Errorf("path de origen vacío")
	}
	if len(destPath) == 0 {
		return fmt.Errorf("path de destino vacío")
	}

	// 6. Navegar al directorio padre del origen
	srcParentPath := srcPath[:len(srcPath)-1]
	srcName := srcPath[len(srcPath)-1]

	var srcParentIno InodeUnified
	var srcParentIdx int32

	if len(srcParentPath) == 0 {
		// El origen está en la raíz
		srcParentIno, err = r.readInodeByIndex(f, sb, 0, region)
		if err != nil {
			return fmt.Errorf("error leyendo raíz: %w", err)
		}
		srcParentIdx = 0
	} else {
		srcParentIno, srcParentIdx, _, err = r.walkToNode(f, sb, srcParentPath, region)
		if err != nil {
			return fmt.Errorf("directorio padre de origen no existe: %w", err)
		}
	}

	// 7. Verificar que el padre del origen es un directorio
	if !srcParentIno.IsDir() {
		return fmt.Errorf("el padre del origen no es un directorio")
	}

	// 8. Verificar permiso de escritura en directorio padre del origen
	isRoot := uid == 1
	if !canWrite(srcParentIno, uid, gid, isRoot) {
		return fmt.Errorf("sin permiso de escritura en directorio origen")
	}

	// 9. Buscar el inodo del origen en su directorio padre
	srcEntries, err := r.readDirEntriesFromInode(f, sb, srcParentIno, region)
	if err != nil {
		return err
	}

	srcInodeIdx := int32(-1)
	srcEntryBlockIdx := int32(-1)
	srcEntrySlotIdx := -1

	for _, e := range srcEntries {
		if e.Name == srcName {
			srcInodeIdx = e.InodeIdx
			break
		}
	}

	if srcInodeIdx == -1 {
		return fmt.Errorf("no existe el origen: %s", strings.Join(srcPath, "/"))
	}

	// Encontrar el bloque y slot exacto donde está la entrada del origen
	srcEntryBlockIdx, srcEntrySlotIdx, err = r.findEntryLocation(f, sb, srcParentIno, srcName, region)
	if err != nil {
		return err
	}

	// 10. Leer el inodo del origen (para verificar que existe)
	srcIno, err := r.readInodeByIndex(f, sb, srcInodeIdx, region)
	if err != nil {
		return fmt.Errorf("error leyendo inodo origen: %w", err)
	}

	// 11. Separar destPath en directorio padre + nombre del archivo destino
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

	// 12. Navegar al directorio padre del destino (debe existir y ser directorio)
	var destDirIno InodeUnified
	if len(destDirPath) == 0 {
		// Destino es en raíz, leer inodo raíz (índice 0)
		destDirIno, err = r.readInodeByIndex(f, sb, 0, region)
		if err != nil {
			return fmt.Errorf("no se pudo leer inodo raíz: %w", err)
		}
	} else {
		// Navegar al directorio padre
		destDirIno, _, _, err = r.walkToNode(f, sb, destDirPath, region)
		if err != nil {
			return fmt.Errorf("directorio padre del destino no existe: %w", err)
		}
	}

	// 13. Verificar que el destino es un directorio
	if !destDirIno.IsDir() {
		return fmt.Errorf("el padre del destino no es un directorio")
	}

	// 14. Verificar permiso de escritura en directorio destino
	if !canWrite(destDirIno, uid, gid, isRoot) {
		return fmt.Errorf("sin permiso de escritura en directorio destino")
	}

	// 15. Verificar que no existe ya un elemento con el nombre destino
	destEntries, err := r.readDirEntriesFromInode(f, sb, destDirIno, region)
	if err != nil {
		return err
	}

	for _, e := range destEntries {
		if e.Name != "." && e.Name != ".." && strings.EqualFold(e.Name, destName) {
			return fmt.Errorf("ya existe '%s' en destino", destName)
		}
	}

	// 16. Leer bitmaps (solo para actualizar si es necesario expandir el destino)
	bmBlock, err := r.readBitmapFromSB(f, sb, false, region)
	if err != nil {
		return err
	}

	// 17. OPERACIÓN MOVE: Re-enlazar el inodo en el destino con el nuevo nombre
	// Agregar entrada en directorio destino apuntando al mismo inodo
	if err := r.addEntryToDirectory(f, sb, destDirIno, destName, srcInodeIdx, bmBlock, region); err != nil {
		return err
	}

	// 18. Actualizar timestamp del directorio destino
	destDirIno.SetMtime(time.Now().Unix())
	if err := r.writeInodeToSB(f, sb, destDirIno, region); err != nil {
		return err
	}

	// 19. Eliminar entrada del directorio origen
	if err := r.removeEntryFromDirectory(f, sb, srcParentIno, srcEntryBlockIdx, srcEntrySlotIdx, region); err != nil {
		return err
	}

	// 20. Actualizar timestamp del directorio origen
	srcParentIno.SetMtime(time.Now().Unix())
	if err := r.writeInodeToSB(f, sb, srcParentIno, region); err != nil {
		return err
	}

	// 21. Si el origen es un directorio, actualizar su entrada ".." para apuntar al nuevo padre
	if srcIno.IsDir() {
		if err := r.updateParentPointer(f, sb, srcIno, destDirIno.Index, region); err != nil {
			return err
		}
	}

	// 22. Escribir bitmap actualizado (si se expandió el destino)
	if err := r.writeBitmapToSB(f, sb, bmBlock, false, region); err != nil {
		return err
	}

	// 23. Actualizar superblock
	sb.Ext2.SMtime = time.Now().Unix()
	if err := r.writeSuperblock(f, sb, region); err != nil {
		return err
	}

	// Registrar en journal (ignorar errores para no romper la operación)
	srcFullPath := "/" + strings.Join(srcPath, "/")
	// Construir la ruta completa del destino correctamente
	destFullPath := "/" + strings.Join(append(destDirPath, destName), "/")
	contentInfo := fmt.Sprintf("to=%s", destFullPath)
	_ = r.JournalAppendPublic(id, "MOVE", srcFullPath, contentInfo, time.Now().Unix())

	_ = srcParentIdx // Puede usarse para logging
	return nil
}

// ================================================================
// FUNCIONES AUXILIARES PARA MOVE
// ================================================================

// findEntryLocation busca una entrada por nombre y devuelve el bloque y slot donde está
func (r *FileFsRepository) findEntryLocation(
	f *os.File,
	sb SuperBlockUnified,
	dirIno InodeUnified,
	name string,
	region Region,
) (blockIdx int32, slotIdx int, err error) {
	if !dirIno.IsDir() {
		return -1, -1, fmt.Errorf("no es un directorio")
	}

	// Buscar en bloques directos
	for i := 0; i < 12; i++ {
		blkIdx := dirIno.Block(i)
		if blkIdx == -1 {
			continue
		}

		block, err := r.readBlockByIndex(f, sb, blkIdx, true, region)
		if err != nil {
			return -1, -1, err
		}

		for j := range block.FolderBlock.Content {
			if block.FolderBlock.Content[j].BInodo == -1 {
				continue
			}
			entryName := strings.TrimRight(string(block.FolderBlock.Content[j].BName[:]), "\x00")
			if entryName == name {
				return blkIdx, j, nil
			}
		}
	}

	return -1, -1, fmt.Errorf("entrada no encontrada: %s", name)
}

// removeEntryFromDirectory elimina una entrada de un directorio marcándola como libre
func (r *FileFsRepository) removeEntryFromDirectory(
	f *os.File,
	sb SuperBlockUnified,
	dirIno InodeUnified,
	blockIdx int32,
	slotIdx int,
	region Region,
) error {
	if !dirIno.IsDir() {
		return fmt.Errorf("no es un directorio")
	}

	// Leer el bloque que contiene la entrada
	block, err := r.readBlockByIndex(f, sb, blockIdx, true, region)
	if err != nil {
		return err
	}

	// Marcar la entrada como libre
	block.FolderBlock.Content[slotIdx].BInodo = -1
	for i := range block.FolderBlock.Content[slotIdx].BName {
		block.FolderBlock.Content[slotIdx].BName[i] = 0
	}

	// Escribir el bloque actualizado
	if err := r.writeBlockToSB(f, sb, blockIdx, block, region); err != nil {
		return err
	}

	return nil
}

// updateParentPointer actualiza la entrada ".." de un directorio para apuntar a un nuevo padre
func (r *FileFsRepository) updateParentPointer(
	f *os.File,
	sb SuperBlockUnified,
	dirIno InodeUnified,
	newParentIdx int32,
	region Region,
) error {
	if !dirIno.IsDir() {
		return fmt.Errorf("no es un directorio")
	}

	// Buscar el primer bloque del directorio (que debe contener "." y "..")
	blkIdx := dirIno.Block(0)
	if blkIdx == -1 {
		return fmt.Errorf("directorio sin bloques")
	}

	block, err := r.readBlockByIndex(f, sb, blkIdx, true, region)
	if err != nil {
		return err
	}

	// Buscar la entrada ".."
	found := false
	for i := range block.FolderBlock.Content {
		entryName := strings.TrimRight(string(block.FolderBlock.Content[i].BName[:]), "\x00")
		if entryName == ".." {
			block.FolderBlock.Content[i].BInodo = newParentIdx
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("entrada '..' no encontrada")
	}

	// Escribir el bloque actualizado
	if err := r.writeBlockToSB(f, sb, blkIdx, block, region); err != nil {
		return err
	}

	return nil
}
