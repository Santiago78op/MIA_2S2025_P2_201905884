package utils

import "strings"

// ParseSMIA procesa un script .smia: elimina comentarios, líneas vacías, y divide en comandos ejecutables.
// Los comentarios comienzan con # y van hasta el fin de línea.
func ParseSMIA(script string) []string {
	lines := strings.Split(script, "\n")
	var commands []string

	for _, line := range lines {
		// Eliminar comentarios
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}

		// Trim espacios
		line = strings.TrimSpace(line)

		// Ignorar líneas vacías
		if line == "" {
			continue
		}

		commands = append(commands, line)
	}

	return commands
}
