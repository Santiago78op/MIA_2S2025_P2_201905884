package diskio

import (
	"Backend/core/models"
	"fmt"
	"os"
	"strings"
	"time"
)

// ================================================================
// COMANDO REMOVE - Elimina archivos/directorios con validación de permisos
// ================================================================

// Remove elimina un archivo o directorio
// Reglas:
// - Requiere permiso de ESCRITURA en el elemento a eliminar
// - Si es directorio, valida ESCRITURA recursiva en TODO el subárbol
// - Si algún hijo no tiene permiso, NO elimina NADA (atómica)
// - Libera inodos y bloques en bitmaps
// - Registra en Journal (EXT3)
func (r *FileFsRepository) Remove(id string, path []string, recursive bool, uid int, gid int) error {
	// 1. Resolver montaje
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

	// 3. Leer superblock
	sb, isExt3, err := r.readAnySuperblock(f, region)
	if err != nil {
		return err
	}

	// Write-ahead journal omitido - se registra al final de la operación
	_ = isExt3

	isRoot := uid == 1

	// 5. Navegar al elemento y validar permisos recursivamente
	targetIno, _, _, err := r.walkToNode(f, sb, path, region)
	if err != nil {
		return fmt.Errorf("elemento no encontrado: %w", err)
	}
	
	// Validar permisos recursivamente ANTES de eliminar nada
	if err := r.validateRemovePermissionsRecursive(f, sb, targetIno, uid, gid, isRoot, region); err != nil {
		return err
	}

	// 6. Navegar al elemento a eliminar (reuses targetIno from validation above)
	parentIno, parentIdx, targetName, err := r.walkToNode(f, sb, path[:len(path)-1], region)
	if err != nil && len(path) > 1 {
		return fmt.Errorf("directorio padre no encontrado: %w", err)
	}
	if len(path) == 1 {
		parentIdx = 0 // raíz
		parentIno, err = r.readInodeByIndex(f, sb, 0, region)
		if err != nil {
			return err
		}
	}

	// 7. Leer bitmaps
	bmInode, err := r.readBitmapFromSB(f, sb, true, region)
	if err != nil {
		return err
	}
	bmBlock, err := r.readBitmapFromSB(f, sb, false, region)
	if err != nil {
		return err
	}

	// 8. Eliminar recursivamente (liberar inodos y bloques)
	freedInodes, freedBlocks := 0, 0
	if err := r.removeRecursive(f, sb, targetIno, bmInode, bmBlock, &freedInodes, &freedBlocks, region); err != nil {
		return err
	}

	// 9. Eliminar entrada del directorio padre
	if parentIdx != -1 {
		if err := r.removeEntryFromDir(f, sb, parentIno, targetName, region); err != nil {
			return err
		}

		// Actualizar inodo padre (mtime)
		parentIno.SetMtime(time.Now().Unix())
		if err := r.writeInodeToSB(f, sb, parentIno, region); err != nil {
			return err
		}
	}

	// 10. Escribir bitmaps actualizados
	if err := r.writeBitmapToSB(f, sb, bmInode, true, region); err != nil {
		return err
	}
	if err := r.writeBitmapToSB(f, sb, bmBlock, false, region); err != nil {
		return err
	}

	// 11. Actualizar contadores en superblock
	if sb.IsExt3 {
		sb.Ext3.FreeInodes += int32(freedInodes)
		sb.Ext3.FreeBlocks += int32(freedBlocks)
	} else {
		sb.Ext2.SFreeInodesCount += int32(freedInodes)
		sb.Ext2.SFreeBlocksCount += int32(freedBlocks)
	}

	if err := r.writeSuperblock(f, sb, region); err != nil {
		return err
	}

	// Registrar en journal (ignorar errores para no romper la operación)
	fullPath := "/" + strings.Join(path, "/")
	_ = r.JournalAppendPublic(id, "REMOVE", fullPath, "", time.Now().Unix())

	return nil
}


// validateRemovePermissionsRecursive valida que tiene permisos de escritura en TODO el árbol
func (r *FileFsRepository) validateRemovePermissionsRecursive(
	f *os.File,
	sb SuperBlockUnified,
	ino InodeUnified,
	uid int, gid int, isRoot bool,
	region Region,
) error {
	// Verificar permiso de escritura en este inodo
	if !canWrite(ino, uid, gid, isRoot) {
		return fmt.Errorf("permiso denegado: no tiene permiso de escritura")
	}

	// Si es directorio, validar recursivamente en todos los hijos
	if ino.IsDir() {
		entries, err := r.readDirEntriesFromInode(f, sb, ino, region)
		if err != nil {
			return err
		}

		for _, e := range entries {
			if e.Name == "." || e.Name == ".." {
				continue
			}

			childIno, err := r.readInodeByIndex(f, sb, e.InodeIdx, region)
			if err != nil {
				continue
			}

			// Validar recursivamente
			if err := r.validateRemovePermissionsRecursive(f, sb, childIno, uid, gid, isRoot, region); err != nil {
				return err
			}
		}
	}

	return nil
}

// removeRecursive elimina un inodo y todos sus hijos recursivamente
func (r *FileFsRepository) removeRecursive(
	f *os.File,
	sb SuperBlockUnified,
	ino InodeUnified,
	bmInode, bmBlock []byte,
	freedInodes, freedBlocks *int,
	region Region,
) error {
	// Si es directorio, eliminar hijos primero
	if ino.IsDir() {
		entries, err := r.readDirEntriesFromInode(f, sb, ino, region)
		if err != nil {
			return err
		}

		for _, e := range entries {
			if e.Name == "." || e.Name == ".." {
				continue
			}

			childIno, err := r.readInodeByIndex(f, sb, e.InodeIdx, region)
			if err != nil {
				continue
			}

			if err := r.removeRecursive(f, sb, childIno, bmInode, bmBlock, freedInodes, freedBlocks, region); err != nil {
				return err
			}
		}
	}

	// Liberar bloques del inodo
	for i := 0; i < 12; i++ {
		blkIdx := ino.Block(i)
		if blkIdx == -1 {
			break
		}

		// Marcar bloque como libre
		bmBlock[blkIdx] = 0
		*freedBlocks++
	}

	// Marcar inodo como libre
	bmInode[ino.Index] = 0
	*freedInodes++

	return nil
}

// removeEntryFromDir elimina una entrada de un directorio
func (r *FileFsRepository) removeEntryFromDir(f *os.File, sb SuperBlockUnified, dirIno InodeUnified, name string, region Region) error {
	// Iterar sobre los bloques del directorio
	for i := 0; i < 12; i++ {
		blkIdx := dirIno.Block(i)
		if blkIdx == -1 {
			break
		}

		blk, err := r.readBlockByIndex(f, sb, blkIdx, true, region)
		if err != nil {
			return err
		}

		// Buscar la entrada
		for j := 0; j < models.DirEntriesPerBlk; j++ {
			entry := &blk.FolderBlock.Content[j]
			entryName := strings.TrimRight(string(entry.BName[:]), "\x00")

			if entryName == name {
				// Marcar como libre
				entry.BInodo = -1
				for k := range entry.BName {
					entry.BName[k] = 0
				}

				// Escribir bloque actualizado
				return r.writeBlockToSB(f, sb, blkIdx, blk.FolderBlock, region)
			}
		}
	}

	return fmt.Errorf("entrada '%s' no encontrada en directorio", name)
}

// Las funciones readDirEntriesFromInode, walkToNode, canRead y canWrite
// ya están definidas en unified_helpers.go y file_repo_copy.go
