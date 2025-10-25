package diskio

import (
	"Backend/core/models"
	"Backend/core/ports"
	"fmt"
	"os"
	"strings"
)

// ListDirectory lista el contenido de un directorio
func (r *FileFsRepository) ListDirectory(id string, path []string) ([]ports.DirectoryEntry, error) {
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Leer superblock
	sb, _, err := r.readAnySuperblock(f, region)
	if err != nil {
		return nil, err
	}

	// Resolver la ruta al inodo
	ino, err := r.resolvePath(f, sb, path, region)
	if err != nil {
		return nil, fmt.Errorf("ruta no encontrada: %w", err)
	}

	// Verificar que sea un directorio
	if !ino.IsDir() {
		return nil, fmt.Errorf("la ruta no es un directorio")
	}

	// Leer las entradas del directorio
	entries := make([]ports.DirectoryEntry, 0)

	// Iterar sobre los bloques directos del inodo
	for i := 0; i < 12; i++ {
		blkIdx := ino.Block(i)
		if blkIdx == -1 {
			break
		}

		// Leer el bloque como directorio
		blk, err := r.readBlockByIndex(f, sb, blkIdx, true, region)
		if err != nil {
			return nil, fmt.Errorf("error leyendo bloque %d: %w", blkIdx, err)
		}

		// Iterar sobre las entradas del bloque
		for j := 0; j < models.DirEntriesPerBlk; j++ {
			entry := blk.FolderBlock.Content[j]
			if entry.BInodo == -1 {
				continue
			}

			name := strings.TrimRight(string(entry.BName[:]), "\x00")
			if name == "" || name == "." || name == ".." {
				continue
			}

			// Leer el inodo de la entrada
			entryIno, err := r.readInodeByIndex(f, sb, entry.BInodo, region)
			if err != nil {
				continue
			}

			// Determinar el tipo
			entryType := "file"
			if entryIno.IsDir() {
				entryType = "dir"
			}

			// Convertir permisos octales a formato rwx
			perm := convertOctalPermToRWX(entryIno.Perm())

			// Obtener usuario y grupo
			owner := fmt.Sprintf("%d", entryIno.UID())
			group := fmt.Sprintf("%d", entryIno.GID())

			entries = append(entries, ports.DirectoryEntry{
				Name:  name,
				Type:  entryType,
				Size:  int64(entryIno.Size()),
				Perm:  perm,
				Owner: owner,
				Group: group,
				Mtime: entryIno.Ext2.IMtime,
			})
		}
	}

	return entries, nil
}

// resolvePath resuelve una ruta a su inodo correspondiente
func (r *FileFsRepository) resolvePath(f *os.File, sb SuperBlockUnified, path []string, region Region) (InodeUnified, error) {
	// Empezar desde el inodo raíz (0)
	current, err := r.readInodeByIndex(f, sb, 0, region)
	if err != nil {
		return InodeUnified{}, err
	}

	// Si la ruta está vacía o es solo "/", devolver la raíz
	if len(path) == 0 || (len(path) == 1 && path[0] == "") {
		return current, nil
	}

	// Recorrer cada componente de la ruta
	for _, component := range path {
		if component == "" || component == "." {
			continue
		}

		// Verificar que current es un directorio
		if !current.IsDir() {
			return InodeUnified{}, fmt.Errorf("'%s' no es un directorio", component)
		}

		// Buscar el componente en los bloques del directorio
		found := false
		var nextIno InodeUnified

		for i := 0; i < 12; i++ {
			blkIdx := current.Block(i)
			if blkIdx == -1 {
				break
			}

			// Leer el bloque como directorio
			blk, err := r.readBlockByIndex(f, sb, blkIdx, true, region)
			if err != nil {
				return InodeUnified{}, err
			}

			// Buscar el componente en las entradas
			for j := 0; j < models.DirEntriesPerBlk; j++ {
				entry := blk.FolderBlock.Content[j]
				if entry.BInodo == -1 {
					continue
				}

				name := strings.TrimRight(string(entry.BName[:]), "\x00")
				if name == component {
					// Encontrado, leer el inodo
					nextIno, err = r.readInodeByIndex(f, sb, entry.BInodo, region)
					if err != nil {
						return InodeUnified{}, err
					}
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return InodeUnified{}, fmt.Errorf("'%s' no encontrado", component)
		}

		current = nextIno
	}

	return current, nil
}

// convertOctalPermToRWX convierte permisos octales a formato rwx
// Ejemplo: [3]byte{'6','6','4'} -> "rw-rw-r--"
func convertOctalPermToRWX(octPerm [3]byte) string {
	result := ""

	for i := 0; i < 3; i++ {
		// Convertir carácter ASCII a número
		digit := int(octPerm[i] - '0')

		// Validar que esté en rango 0-7
		if digit < 0 || digit > 7 {
			result += "---"
			continue
		}

		// Convertir a rwx
		// bit 2 (4): read
		// bit 1 (2): write
		// bit 0 (1): execute
		if digit&4 != 0 {
			result += "r"
		} else {
			result += "-"
		}
		if digit&2 != 0 {
			result += "w"
		} else {
			result += "-"
		}
		if digit&1 != 0 {
			result += "x"
		} else {
			result += "-"
		}
	}

	return result
}
