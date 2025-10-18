package adapters

import (
	"fmt"
	"strconv"
	"strings"

	"Backend/command/users"
	"Backend/storage/diskio"
)

// UsersAdapter adapta FileFsRepository para cumplir con users.FsUsersRepository
type UsersAdapter struct {
	repo *diskio.FileFsRepository
}

func NewUsersAdapter(repo *diskio.FileFsRepository) users.FsUsersRepository {
	return &UsersAdapter{repo: repo}
}

func (a *UsersAdapter) Login(id, user, pass string) (uid, gid int, isRoot bool, err error) {
	// Leer users.txt como root (uid=1, gid=1) ya que es una operación del sistema
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}}, 1, 1)
	if err != nil {
		return 0, 0, false, fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		idStr := parts[0]
		typ := parts[1]

		// Solo nos interesan usuarios
		if typ != "U" {
			continue
		}

		if len(parts) < 5 {
			continue
		}

		// Formato correcto: UID,Tipo,Grupo,Usuario,Contraseña
		groupname := parts[2]
		username := parts[3]
		password := parts[4]

		if username == user && password == pass {
			// Parsear UID
			parsedUID, err := strconv.Atoi(idStr)
			if err != nil {
				return 0, 0, false, fmt.Errorf("error parseando UID: %w", err)
			}

			// Buscar GID del grupo
			groupGID := 0
			for _, gline := range lines {
				gline = strings.TrimSpace(gline)
				if gline == "" {
					continue
				}
				gparts := strings.Split(gline, ",")
				if len(gparts) >= 3 && gparts[1] == "G" && gparts[2] == groupname {
					groupGID, _ = strconv.Atoi(gparts[0])
					break
				}
			}

			isRoot := username == "root"
			return parsedUID, groupGID, isRoot, nil
		}
	}

	return 0, 0, false, fmt.Errorf("usuario o contraseña incorrectos")
}

func (a *UsersAdapter) Mkgrp(id, name string) error {
	// Validar longitud del nombre del grupo (máximo 10 caracteres)
	if len(name) > 10 {
		return fmt.Errorf("el nombre del grupo excede el máximo de 10 caracteres")
	}

	// Leer users.txt como root (uid=1, gid=1) ya que es una operación del sistema
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}}, 1, 1)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Verificar que el grupo no existe (ignorando grupos eliminados con GID=0)
	// y buscar si hay un espacio reutilizable con GID=0 y el mismo nombre
	reuseLineIndex := -1
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == name {
			gid, err := strconv.Atoi(parts[0])
			if err == nil && gid > 0 {
				// El grupo existe y está activo
				return fmt.Errorf("el grupo ya existe: %s", name)
			}
			// Encontramos un grupo eliminado (GID=0) con el mismo nombre, podemos reutilizarlo
			if gid == 0 {
				reuseLineIndex = i
			}
		}
	}

	// Si encontramos un espacio reutilizable, actualizar esa línea
	if reuseLineIndex >= 0 {
		// Encontrar el próximo ID disponible
		maxID := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, ",")
			if len(parts) >= 1 {
				idNum, err := strconv.Atoi(parts[0])
				if err == nil && idNum > maxID {
					maxID = idNum
				}
			}
		}
		newID := maxID + 1

		// Actualizar la línea reutilizable
		var newLines []string
		for i, line := range lines {
			if i == reuseLineIndex {
				newLines = append(newLines, fmt.Sprintf("%d,G,%s", newID, name))
			} else {
				newLines = append(newLines, strings.TrimSpace(line))
			}
		}
		newContent := strings.Join(newLines, "\n") + "\n"
		return a.repo.UpdateFile(id, []string{"users.txt"}, newContent)
	}

	// No hay espacio reutilizable, encontrar el próximo ID disponible y agregar al final
	maxID := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 1 {
			idNum, err := strconv.Atoi(parts[0])
			if err == nil && idNum > maxID {
				maxID = idNum
			}
		}
	}
	newID := maxID + 1

	// Agregar nueva línea
	newLine := fmt.Sprintf("%d,G,%s\n", newID, name)
	newContent := content + newLine

	// Escribir el archivo actualizado
	return a.repo.UpdateFile(id, []string{"users.txt"}, newContent)
}

func (a *UsersAdapter) Rmgrp(id, name string) error {
	// Leer users.txt como root (uid=1, gid=1) ya que es una operación del sistema
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}}, 1, 1)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Verificar que el grupo existe y no es root
	found := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == name {
			// Verificar que el ID no sea 0 (ya eliminado)
			if parts[0] == "0" {
				return fmt.Errorf("el grupo no existe: %s", name)
			}
			if name == "root" {
				return fmt.Errorf("no se puede eliminar el grupo root")
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("el grupo no existe: %s", name)
	}

	// Verificar que no hay usuarios activos usando este grupo
	// Formato: UID,Tipo,Grupo,Usuario,Contraseña - verificar parts[2] (Grupo)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 4 && parts[1] == "U" && parts[2] == name && parts[0] != "0" {
			return fmt.Errorf("no se puede eliminar el grupo, hay usuarios asociados: %s", name)
		}
	}

	// Marcar grupo como eliminado (cambiar a 0,G,nombre)
	var newLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == name {
			newLines = append(newLines, "0,G,"+name)
		} else {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n") + "\n"
	return a.repo.UpdateFile(id, []string{"users.txt"}, newContent)
}

func (a *UsersAdapter) Mkusr(id, user, pass, grp string) error {
	// Validar longitudes (máximo 10 caracteres cada uno)
	if len(user) > 10 {
		return fmt.Errorf("el nombre del usuario excede el máximo de 10 caracteres")
	}
	if len(pass) > 10 {
		return fmt.Errorf("la contraseña excede el máximo de 10 caracteres")
	}
	if len(grp) > 10 {
		return fmt.Errorf("el nombre del grupo excede el máximo de 10 caracteres")
	}

	// Leer users.txt como root (uid=1, gid=1) ya que es una operación del sistema
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}}, 1, 1)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Verificar que el usuario no existe (incluyendo eliminados)
	// Formato: UID,Tipo,Grupo,Usuario,Contraseña
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 4 && parts[1] == "U" && parts[3] == user && parts[0] != "0" {
			return fmt.Errorf("el usuario ya existe: %s", user)
		}
	}

	// Verificar que el grupo existe
	groupExists := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == grp {
			idNum, _ := strconv.Atoi(parts[0])
			if idNum > 0 {
				groupExists = true
			}
			break
		}
	}

	if !groupExists {
		return fmt.Errorf("el grupo no existe: %s", grp)
	}

	// Encontrar el próximo ID disponible
	maxID := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 1 {
			idNum, err := strconv.Atoi(parts[0])
			if err == nil && idNum > maxID {
				maxID = idNum
			}
		}
	}
	newID := maxID + 1

	// Agregar nueva línea con formato correcto: UID,Tipo,Grupo,Usuario,Contraseña
	newLine := fmt.Sprintf("%d,U,%s,%s,%s\n", newID, grp, user, pass)
	newContent := content + newLine

	return a.repo.UpdateFile(id, []string{"users.txt"}, newContent)
}

func (a *UsersAdapter) Rmusr(id, user string) error {
	// Leer users.txt como root (uid=1, gid=1) ya que es una operación del sistema
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}}, 1, 1)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Verificar que el usuario existe y no es root
	// Formato: UID,Tipo,Grupo,Usuario,Contraseña
	found := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 4 && parts[1] == "U" && parts[3] == user {
			// Verificar que el ID no sea 0 (ya eliminado)
			if parts[0] == "0" {
				return fmt.Errorf("el usuario no existe: %s", user)
			}
			if user == "root" {
				return fmt.Errorf("no se puede eliminar el usuario root")
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("el usuario no existe: %s", user)
	}

	// Marcar usuario como eliminado (cambiar a 0,U,grupo,nombre,pass)
	// Formato: UID,Tipo,Grupo,Usuario,Contraseña
	var newLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 5 && parts[1] == "U" && parts[3] == user {
			newLines = append(newLines, fmt.Sprintf("0,U,%s,%s,%s", parts[2], parts[3], parts[4]))
		} else {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n") + "\n"
	return a.repo.UpdateFile(id, []string{"users.txt"}, newContent)
}

func (a *UsersAdapter) Chgrp(id, user, grp string) error {
	// Validar longitud del nombre del grupo (máximo 10 caracteres)
	if len(grp) > 10 {
		return fmt.Errorf("el nombre del grupo excede el máximo de 10 caracteres")
	}

	// Leer users.txt como root (uid=1, gid=1) ya que es una operación del sistema
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}}, 1, 1)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Verificar que el usuario existe
	// Formato: UID,Tipo,Grupo,Usuario,Contraseña
	found := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 4 && parts[1] == "U" && parts[3] == user {
			// Verificar que el ID no sea 0 (ya eliminado)
			if parts[0] == "0" {
				return fmt.Errorf("el usuario no existe: %s", user)
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("el usuario no existe: %s", user)
	}

	// Verificar que el grupo existe
	groupExists := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == grp {
			idNum, _ := strconv.Atoi(parts[0])
			if idNum > 0 {
				groupExists = true
			}
			break
		}
	}

	if !groupExists {
		return fmt.Errorf("el grupo no existe: %s", grp)
	}

	// Cambiar grupo del usuario
	// Formato: UID,Tipo,Grupo,Usuario,Contraseña
	var newLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 5 && parts[1] == "U" && parts[3] == user {
			newLines = append(newLines, fmt.Sprintf("%s,U,%s,%s,%s", parts[0], grp, user, parts[4]))
		} else {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n") + "\n"
	return a.repo.UpdateFile(id, []string{"users.txt"}, newContent)
}
