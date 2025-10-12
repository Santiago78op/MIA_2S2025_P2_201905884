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
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}})
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

		username := parts[2]
		groupname := parts[3]
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
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}})
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Verificar que el grupo no existe
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == name {
			return fmt.Errorf("el grupo ya existe: %s", name)
		}
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

	// Agregar nueva línea
	newLine := fmt.Sprintf("%d,G,%s\n", newID, name)
	newContent := content + newLine

	// TODO: Implementar escritura de archivo (por ahora solo validamos)
	// Necesitaríamos implementar un método Write en FileFsRepository
	_ = newContent
	return fmt.Errorf("TODO: implementar escritura de archivo para Mkgrp")
}

func (a *UsersAdapter) Rmgrp(id, name string) error {
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}})
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

	// Verificar que no hay usuarios usando este grupo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 4 && parts[1] == "U" && parts[3] == name {
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
	_ = newContent
	return fmt.Errorf("TODO: implementar escritura de archivo para Rmgrp")
}

func (a *UsersAdapter) Mkusr(id, user, pass, grp string) error {
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}})
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Verificar que el usuario no existe
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "U" && parts[2] == user {
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

	// Agregar nueva línea
	newLine := fmt.Sprintf("%d,U,%s,%s,%s\n", newID, user, grp, pass)
	newContent := content + newLine

	_ = newContent
	return fmt.Errorf("TODO: implementar escritura de archivo para Mkusr")
}

func (a *UsersAdapter) Rmusr(id, user string) error {
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}})
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Verificar que el usuario existe y no es root
	found := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "U" && parts[2] == user {
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

	// Marcar usuario como eliminado (cambiar a 0,U,nombre,grupo,pass)
	var newLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 5 && parts[1] == "U" && parts[2] == user {
			newLines = append(newLines, fmt.Sprintf("0,U,%s,%s,%s", parts[2], parts[3], parts[4]))
		} else {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n") + "\n"
	_ = newContent
	return fmt.Errorf("TODO: implementar escritura de archivo para Rmusr")
}

func (a *UsersAdapter) Chgrp(id, user, grp string) error {
	content, err := a.repo.Cat(id, [][]string{{"users.txt"}})
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Verificar que el usuario existe
	found := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "U" && parts[2] == user {
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
	var newLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 5 && parts[1] == "U" && parts[2] == user {
			newLines = append(newLines, fmt.Sprintf("%s,U,%s,%s,%s", parts[0], user, grp, parts[4]))
		} else {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n") + "\n"
	_ = newContent
	return fmt.Errorf("TODO: implementar escritura de archivo para Chgrp")
}
