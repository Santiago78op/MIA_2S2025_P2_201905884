package diskio

import (
	"Backend/core/models"
	"fmt"
	"os"
	"strings"
	"time"
)

// ================================================================
// COMANDO RENAME - Renombra un archivo o directorio
// ================================================================

// Rename renombra un archivo o directorio
// Reglas:
// - Requiere permiso de ESCRITURA en el directorio padre
// - No puede existir otro elemento con el mismo nombre en el mismo directorio
// - Registra en Journal (EXT3)
func (r *FileFsRepository) Rename(id string, path []string, newName string) error {
	// 1. Validar que newName no esté vacío y no contenga '/'
	if newName == "" {
		return fmt.Errorf("el nuevo nombre no puede estar vacío")
	}
	if strings.Contains(newName, "/") {
		return fmt.Errorf("el nuevo nombre no puede contener '/'")
	}
	if len(newName) > 12 {
		return fmt.Errorf("el nuevo nombre no puede exceder 12 caracteres")
	}

	// 2. Resolver montaje
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return err
	}

	// 3. Abrir disco
	f, err := os.OpenFile(diskPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// 4. Leer superblock
	sb, isExt3, err := r.readAnySuperblock(f, region)
	if err != nil {
		return err
	}

	// 5. Write-ahead journal para EXT3
	if isExt3 {
		pathStr := "/" + strings.Join(path, "/")
		j := models.Journal{
			Count: 0,
			Content: models.Information{
				Date: float64(time.Now().Unix()),
			},
		}
		copy(j.Content.Op[:], "RENAME")
		copy(j.Content.Path[:], pathStr)
		copy(j.Content.Content[:], fmt.Sprintf("new_name=%s", newName))

		_ = r.AppendJournal(f, sb.Ext3, j)
	}

	// 6. Navegar al elemento
	targetIno, parentIdx, oldName, err := r.walkToNode(f, sb, path, region)
	if err != nil {
		return fmt.Errorf("elemento no encontrado: %w", err)
	}

	// 7. Verificar que no es la raíz
	if parentIdx == -1 {
		return fmt.Errorf("no se puede renombrar el directorio raíz")
	}

	// 8. Leer el directorio padre
	parentIno, err := r.readInodeByIndex(f, sb, parentIdx, region)
	if err != nil {
		return err
	}

	// TODO: Verificar permiso de escritura en el padre

	// 9. Verificar que no existe otro elemento con el nuevo nombre
	entries, err := r.readDirEntriesFromInode(f, sb, parentIno, region)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.Name != "." && e.Name != ".." && strings.EqualFold(e.Name, newName) && e.Name != oldName {
			return fmt.Errorf("ya existe un elemento con el nombre '%s'", newName)
		}
	}

	// 10. Actualizar la entrada en el directorio padre
	updated := false
	for i := 0; i < 12; i++ {
		blkIdx := parentIno.Block(i)
		if blkIdx == -1 {
			break
		}

		blk, err := r.readBlockByIndex(f, sb, blkIdx, true, region)
		if err != nil {
			return err
		}

		for j := 0; j < models.DirEntriesPerBlk; j++ {
			entry := &blk.FolderBlock.Content[j]
			entryName := strings.TrimRight(string(entry.BName[:]), "\x00")

			if entry.BInodo == targetIno.Index && entryName == oldName {
				// Actualizar el nombre
				for k := range entry.BName {
					entry.BName[k] = 0
				}
				copy(entry.BName[:], newName)

				// Escribir bloque actualizado
				if err := r.writeBlockToSB(f, sb, blkIdx, blk.FolderBlock, region); err != nil {
					return err
				}
				updated = true
				break
			}
		}

		if updated {
			break
		}
	}

	if !updated {
		return fmt.Errorf("no se pudo actualizar la entrada del directorio")
	}

	return nil
}
