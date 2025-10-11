package ext2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"MIA_2S2025_P2_201905884/internal/fs"
)

// P1 user/group management - Implementación completa con users.txt

func (f *FS2) AddGroup(ctx context.Context, h fs.MountHandle, name string) error {
	// 1. Leer users.txt
	content, err := f.readUsersFile(h)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// 2. Verificar si el grupo ya existe
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == name {
			return fmt.Errorf("el grupo '%s' ya existe", name)
		}
	}

	// 3. Encontrar el siguiente ID de grupo
	maxID := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 1 {
			id, _ := strconv.Atoi(parts[0])
			if id > maxID {
				maxID = id
			}
		}
	}
	nextID := maxID + 1

	// 4. Agregar nueva línea
	newLine := fmt.Sprintf("%d,G,%s\n", nextID, name)
	newContent := string(content) + newLine

	// 5. Escribir de vuelta a users.txt
	return f.writeUsersFile(h, []byte(newContent))
}

func (f *FS2) RemoveGroup(ctx context.Context, h fs.MountHandle, name string) error {
	// 1. Leer users.txt
	content, err := f.readUsersFile(h)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false
	var newLines []string

	// 2 & 3. Encontrar y marcar como eliminado (cambiar ID a 0)
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == name {
			// Marcar como eliminado (ID = 0)
			newLines = append(newLines, fmt.Sprintf("0,%s", strings.Join(parts[1:], ",")))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		return fmt.Errorf("grupo '%s' no encontrado", name)
	}

	// 4. Escribir de vuelta
	newContent := strings.Join(newLines, "\n") + "\n"
	return f.writeUsersFile(h, []byte(newContent))
}

func (f *FS2) AddUser(ctx context.Context, h fs.MountHandle, user, pass, group string) error {
	// 1. Leer users.txt
	content, err := f.readUsersFile(h)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// 2. Verificar si el usuario ya existe
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 5 && parts[1] == "U" && parts[2] == user {
			return fmt.Errorf("el usuario '%s' ya existe", user)
		}
	}

	// 3. Verificar que el grupo existe
	groupExists := false
	groupID := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == group {
			groupExists = true
			groupID, _ = strconv.Atoi(parts[0])
			break
		}
	}

	if !groupExists {
		return fmt.Errorf("el grupo '%s' no existe", group)
	}

	// 4. Usar el ID del grupo para el usuario
	// Formato: <gid>,U,<user>,<group>,<pass>
	newLine := fmt.Sprintf("%d,U,%s,%s,%s\n", groupID, user, group, pass)
	newContent := string(content) + newLine

	// 5. Escribir de vuelta
	return f.writeUsersFile(h, []byte(newContent))
}

func (f *FS2) RemoveUser(ctx context.Context, h fs.MountHandle, user string) error {
	// 1. Leer users.txt
	content, err := f.readUsersFile(h)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false
	var newLines []string

	// 2 & 3. Encontrar y marcar como eliminado
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 5 && parts[1] == "U" && parts[2] == user {
			// Marcar como eliminado
			newLines = append(newLines, fmt.Sprintf("0,%s", strings.Join(parts[1:], ",")))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		return fmt.Errorf("usuario '%s' no encontrado", user)
	}

	// 4. Escribir de vuelta
	newContent := strings.Join(newLines, "\n") + "\n"
	return f.writeUsersFile(h, []byte(newContent))
}

func (f *FS2) ChangeUserGroup(ctx context.Context, h fs.MountHandle, user, group string) error {
	// 1. Leer users.txt
	content, err := f.readUsersFile(h)
	if err != nil {
		return fmt.Errorf("error leyendo users.txt: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// 2. Verificar que el nuevo grupo existe
	groupExists := false
	groupID := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[1] == "G" && parts[2] == group {
			groupExists = true
			groupID, _ = strconv.Atoi(parts[0])
			break
		}
	}

	if !groupExists {
		return fmt.Errorf("el grupo '%s' no existe", group)
	}

	// 3. Encontrar y actualizar usuario
	found := false
	var newLines []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 5 && parts[1] == "U" && parts[2] == user {
			// Actualizar grupo: <newGid>,U,<user>,<newGroup>,<pass>
			newLines = append(newLines, fmt.Sprintf("%d,U,%s,%s,%s", groupID, user, group, parts[4]))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		return fmt.Errorf("usuario '%s' no encontrado", user)
	}

	// 4. Escribir de vuelta
	newContent := strings.Join(newLines, "\n") + "\n"
	return f.writeUsersFile(h, []byte(newContent))
}

// Helper functions para leer/escribir users.txt desde el disco

func (f *FS2) readUsersFile(h fs.MountHandle) ([]byte, error) {
	// Leer archivo users.txt desde el inodo en el disco
	// Por ahora, usamos una implementación simplificada que lee desde memory/cache

	// Intentar leer desde disco real
	usersPath := "/users.txt"
	content, _, err := f.ReadFile(context.Background(), h, usersPath)
	if err != nil {
		// Si no existe, crear uno inicial
		initialContent := []byte("1,G,root\n1,U,root,root,123\n")
		return initialContent, nil
	}

	return content, nil
}

func (f *FS2) writeUsersFile(h fs.MountHandle, content []byte) error {
	// Escribir archivo users.txt al disco
	// Por ahora, usamos WriteFile que ya está implementado

	req := fs.WriteFileRequest{
		Path:    "/users.txt",
		Content: content,
		Append:  false,
	}

	return f.WriteFile(context.Background(), h, req)
}
