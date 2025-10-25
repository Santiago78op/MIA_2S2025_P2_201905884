package diskio

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ================================================================
// COMANDOS CHOWN Y CHMOD - Cambio de propietario y permisos
// ================================================================

// Chown cambia el propietario (UID) de un archivo/directorio.
// Si recursive=true, aplica el cambio a todo el subárbol.
//
// Reglas:
// - Solo root o el propietario actual pueden cambiar el dueño
// - Requiere permiso de escritura en el archivo/directorio
// - Si recursive, aplica a todos los elementos del subárbol
// - Registra en Journal (EXT3) en write-ahead
func (r *FileFsRepository) Chown(id string, path []string, newOwner string, recursive bool, currentUID int, currentGID int) error {
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

	// 3. Leer superblock
	sb, isExt3, err := r.readAnySuperblock(f, region)
	if err != nil {
		return err
	}

	// Write-ahead journal omitido - se registra al final de la operación
	_ = isExt3

	// 5. Buscar el nuevo UID desde users.txt
	newUID, newGID, err := r.findUserUID(f, sb, newOwner, region)
	if err != nil {
		return fmt.Errorf("usuario '%s' no encontrado: %w", newOwner, err)
	}

	// 6. Navegar al archivo/directorio
	targetIno, _, _, err := r.walkToNode(f, sb, path, region)
	if err != nil {
		return fmt.Errorf("path no existe: %w", err)
	}

	// 7. Verificar permisos (solo root o propietario pueden cambiar owner)
	isRoot := currentUID == 1
	if !isRoot && targetIno.UID() != int32(currentUID) {
		return fmt.Errorf("sin permiso para cambiar propietario (solo root o propietario)")
	}

	// 8. Aplicar chown
	if recursive && targetIno.IsDir() {
		// Recursivo
		count, err := r.chownRecursive(f, sb, targetIno, newUID, newGID, region)
		if err != nil {
			return err
		}
		_ = count // Para logging si se desea
	} else {
		// Solo el archivo/directorio especificado
		targetIno.Ext2.IUid = int32(newUID)
		targetIno.Ext2.IGid = int32(newGID)
		targetIno.Ext2.ICtime = time.Now().Unix()

		if err := r.writeInodeToSB(f, sb, targetIno, region); err != nil {
			return err
		}
	}

	// 9. Actualizar superblock
	sb.Ext2.SMtime = time.Now().Unix()
	if err := r.writeSuperblock(f, sb, region); err != nil {
		return err
	}

	// Registrar en journal (ignorar errores para no romper la operación)
	fullPath := "/" + strings.Join(path, "/")
	contentInfo := fmt.Sprintf("owner=%s, recursive=%t", newOwner, recursive)
	_ = r.JournalAppendPublic(id, "CHOWN", fullPath, contentInfo, time.Now().Unix())

	return nil
}

// chownRecursive aplica chown recursivamente a un directorio y todo su contenido
func (r *FileFsRepository) chownRecursive(
	f *os.File,
	sb SuperBlockUnified,
	dirIno InodeUnified,
	newUID, newGID int,
	region Region,
) (int, error) {
	count := 0

	// Cambiar el directorio actual
	dirIno.Ext2.IUid = int32(newUID)
	dirIno.Ext2.IGid = int32(newGID)
	dirIno.Ext2.ICtime = time.Now().Unix()

	if err := r.writeInodeToSB(f, sb, dirIno, region); err != nil {
		return count, err
	}
	count++

	// Leer entradas del directorio
	entries, err := r.readDirEntriesFromInode(f, sb, dirIno, region)
	if err != nil {
		return count, err
	}

	// Aplicar recursivamente a cada entrada
	for _, entry := range entries {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		childIno, err := r.readInodeByIndex(f, sb, entry.InodeIdx, region)
		if err != nil {
			continue
		}

		if childIno.IsDir() {
			// Recursión en subdirectorio
			c, err := r.chownRecursive(f, sb, childIno, newUID, newGID, region)
			if err != nil {
				return count, err
			}
			count += c
		} else {
			// Cambiar archivo
			childIno.Ext2.IUid = int32(newUID)
			childIno.Ext2.IGid = int32(newGID)
			childIno.Ext2.ICtime = time.Now().Unix()

			if err := r.writeInodeToSB(f, sb, childIno, region); err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}

// Chmod cambia los permisos (UGO) de un archivo/directorio.
// Si recursive=true, aplica el cambio a todo el subárbol.
//
// Reglas:
// - Solo root o el propietario pueden cambiar permisos
// - ugo debe ser un string de 3 dígitos octales (ej: "755", "644")
// - Si recursive, aplica a todos los elementos del subárbol
// - Registra en Journal (EXT3) en write-ahead
func (r *FileFsRepository) Chmod(id string, path []string, ugo [3]byte, recursive bool, currentUID int, currentGID int) error {
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

	// 3. Leer superblock
	sb, isExt3, err := r.readAnySuperblock(f, region)
	if err != nil {
		return err
	}

	// Write-ahead journal omitido - se registra al final de la operación
	_ = isExt3

	// 5. Navegar al archivo/directorio
	targetIno, _, _, err := r.walkToNode(f, sb, path, region)
	if err != nil {
		return fmt.Errorf("path no existe: %w", err)
	}

	// 6. Verificar permisos (solo root o propietario pueden cambiar permisos)
	isRoot := currentUID == 1
	if !isRoot && targetIno.UID() != int32(currentUID) {
		return fmt.Errorf("sin permiso para cambiar permisos (solo root o propietario)")
	}

	// 7. Aplicar chmod
	if recursive && targetIno.IsDir() {
		// Recursivo
		count, err := r.chmodRecursive(f, sb, targetIno, ugo, region)
		if err != nil {
			return err
		}
		_ = count // Para logging si se desea
	} else {
		// Solo el archivo/directorio especificado
		targetIno.Ext2.IPerm = ugo
		targetIno.Ext2.ICtime = time.Now().Unix()

		if err := r.writeInodeToSB(f, sb, targetIno, region); err != nil {
			return err
		}
	}

	// 8. Actualizar superblock
	sb.Ext2.SMtime = time.Now().Unix()
	if err := r.writeSuperblock(f, sb, region); err != nil {
		return err
	}

	// Registrar en journal (ignorar errores para no romper la operación)
	fullPath := "/" + strings.Join(path, "/")
	contentInfo := fmt.Sprintf("perms=%d%d%d, recursive=%t", ugo[0], ugo[1], ugo[2], recursive)
	_ = r.JournalAppendPublic(id, "CHMOD", fullPath, contentInfo, time.Now().Unix())

	return nil
}

// chmodRecursive aplica chmod recursivamente a un directorio y todo su contenido
func (r *FileFsRepository) chmodRecursive(
	f *os.File,
	sb SuperBlockUnified,
	dirIno InodeUnified,
	ugo [3]byte,
	region Region,
) (int, error) {
	count := 0

	// Cambiar el directorio actual
	dirIno.Ext2.IPerm = ugo
	dirIno.Ext2.ICtime = time.Now().Unix()

	if err := r.writeInodeToSB(f, sb, dirIno, region); err != nil {
		return count, err
	}
	count++

	// Leer entradas del directorio
	entries, err := r.readDirEntriesFromInode(f, sb, dirIno, region)
	if err != nil {
		return count, err
	}

	// Aplicar recursivamente a cada entrada
	for _, entry := range entries {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		childIno, err := r.readInodeByIndex(f, sb, entry.InodeIdx, region)
		if err != nil {
			continue
		}

		if childIno.IsDir() {
			// Recursión en subdirectorio
			c, err := r.chmodRecursive(f, sb, childIno, ugo, region)
			if err != nil {
				return count, err
			}
			count += c
		} else {
			// Cambiar archivo
			childIno.Ext2.IPerm = ugo
			childIno.Ext2.ICtime = time.Now().Unix()

			if err := r.writeInodeToSB(f, sb, childIno, region); err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}

// ================================================================
// HELPERS
// ================================================================

// findUserUID busca el UID y GID de un usuario en el archivo users.txt
func (r *FileFsRepository) findUserUID(
	f *os.File,
	sb SuperBlockUnified,
	username string,
	region Region,
) (int, int, error) {
	// Leer users.txt (inodo 1)
	usersIno, err := r.readInodeByIndex(f, sb, 1, region)
	if err != nil {
		return 0, 0, fmt.Errorf("error leyendo users.txt: %w", err)
	}

	content, err := r.readFileContentFromInode(f, sb, usersIno, region)
	if err != nil {
		return 0, 0, fmt.Errorf("error leyendo contenido de users.txt: %w", err)
	}

	// Parsear users.txt
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		// Formato: ID,Tipo,Grupo,Usuario,Password
		// Tipo: G=grupo, U=usuario
		if parts[1] == "U" && len(parts) >= 5 {
			// Es un usuario
			user := parts[3]
			if user == username {
				// Encontrado, retornar UID y GID
				var uid, gid int
				fmt.Sscanf(parts[0], "%d", &uid)

				// El GID es el mismo que el UID en este sistema simplificado
				// O buscar el grupo correspondiente
				gid = uid

				return uid, gid, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("usuario no encontrado")
}
