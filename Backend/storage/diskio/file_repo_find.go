package diskio

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ================================================================
// COMANDO FIND - Búsqueda recursiva con patrones glob (* y ?)
// ================================================================

// Find busca recursivamente archivos/directorios que coincidan con un patrón glob.
// Respeta permisos de lectura en cada directorio visitado.
//
// Patrones soportados:
// - '*' : Coincide con cualquier secuencia de caracteres (incluyendo vacío)
// - '?' : Coincide con exactamente un carácter
// - Literales: Caracteres normales deben coincidir exactamente
//
// Reglas:
// - Requiere permiso de LECTURA en cada directorio visitado
// - Omite directorios sin permiso de lectura (no los explora)
// - Busca solo en nombres de archivos/directorios (no en paths completos)
// - Case-insensitive (no distingue mayúsculas/minúsculas)
func (r *FileFsRepository) Find(id string, basePath []string, pattern string, uid int, gid int) ([]string, error) {
	// 1. Resolver montaje y obtener región de la partición
	diskPath, region, err := r.resolve(id)
	if err != nil {
		return nil, err
	}

	// 2. Abrir disco
	f, err := os.OpenFile(diskPath, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 3. Leer superblock (intentar EXT3 primero, luego EXT2)
	sb, _, err := r.readAnySuperblock(f, region)
	if err != nil {
		return nil, err
	}

	// 4. Convertir patrón glob a expresión regular
	regex, err := globToRegex(pattern)
	if err != nil {
		return nil, fmt.Errorf("patrón inválido: %w", err)
	}

	// 5. Navegar al directorio base
	var baseIno InodeUnified
	var basePathStr string

	if len(basePath) == 0 {
		// Buscar desde la raíz
		baseIno, err = r.readInodeByIndex(f, sb, 0, region)
		if err != nil {
			return nil, err
		}
		basePathStr = "/"
	} else {
		baseIno, _, _, err = r.walkToNode(f, sb, basePath, region)
		if err != nil {
			return nil, fmt.Errorf("path base no existe: %w", err)
		}
		basePathStr = "/" + strings.Join(basePath, "/")
	}

	// 6. Verificar que el path base es un directorio
	if !baseIno.IsDir() {
		return nil, fmt.Errorf("path base no es un directorio")
	}

	// 7. Verificar permiso de lectura en el directorio base
	isRoot := uid == 1
	if !canRead(baseIno, uid, gid, isRoot) {
		return nil, fmt.Errorf("sin permiso de lectura en path base")
	}

	// 8. Realizar búsqueda recursiva
	results := []string{}
	err = r.findRecursive(f, sb, baseIno, basePathStr, regex, uid, gid, isRoot, &results, region)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// findRecursive busca recursivamente en un directorio y sus subdirectorios
func (r *FileFsRepository) findRecursive(
	f *os.File,
	sb SuperBlockUnified,
	dirIno InodeUnified,
	currentPath string,
	pattern *regexp.Regexp,
	uid, gid int,
	isRoot bool,
	results *[]string,
	region Region,
) error {
	// 1. Verificar permiso de lectura en el directorio actual
	if !canRead(dirIno, uid, gid, isRoot) {
		// Sin permiso de lectura, omitir este directorio
		return nil
	}

	// 2. Leer entradas del directorio
	entries, err := r.readDirEntriesFromInode(f, sb, dirIno, region)
	if err != nil {
		return err
	}

	// 3. Recorrer entradas
	for _, entry := range entries {
		// Ignorar "." y ".."
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		// Construir path completo
		fullPath := currentPath
		if !strings.HasSuffix(fullPath, "/") {
			fullPath += "/"
		}
		fullPath += entry.Name

		// Verificar si el nombre coincide con el patrón (case-insensitive)
		if pattern.MatchString(strings.ToLower(entry.Name)) {
			*results = append(*results, fullPath)
		}

		// Si es un directorio, buscar recursivamente
		childIno, err := r.readInodeByIndex(f, sb, entry.InodeIdx, region)
		if err != nil {
			// Si no se puede leer el inodo, continuar con la siguiente entrada
			continue
		}

		if childIno.IsDir() {
			// Recursión en subdirectorio
			if err := r.findRecursive(f, sb, childIno, fullPath, pattern, uid, gid, isRoot, results, region); err != nil {
				// Si hay error en subdirectorio, continuar con otros
				continue
			}
		}
	}

	return nil
}

// ================================================================
// HELPER: Conversión de patrón glob a expresión regular
// ================================================================

// globToRegex convierte un patrón glob (con * y ?) a una expresión regular
// que coincida con nombres de archivos (case-insensitive).
//
// Patrones:
// - '*' → '.*' (cualquier secuencia de caracteres)
// - '?' → '.' (exactamente un carácter)
// - Otros caracteres especiales de regex se escapan
//
// Ejemplos:
// - "*.txt" → "^.*\\.txt$" (archivos que terminan en .txt)
// - "file?.log" → "^file.\\.log$" (file1.log, fileA.log, etc.)
// - "test*" → "^test.*$" (test, test1, testing, etc.)
func globToRegex(pattern string) (*regexp.Regexp, error) {
	var regexPattern strings.Builder

	// Anclar al inicio
	regexPattern.WriteString("^")

	// Procesar cada carácter del patrón
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]

		switch c {
		case '*':
			// * → .* (cualquier secuencia)
			regexPattern.WriteString(".*")

		case '?':
			// ? → . (un carácter)
			regexPattern.WriteString(".")

		case '.', '+', '(', ')', '[', ']', '{', '}', '^', '$', '|', '\\':
			// Escapar caracteres especiales de regex
			regexPattern.WriteString("\\")
			regexPattern.WriteByte(c)

		default:
			// Carácter literal
			regexPattern.WriteByte(c)
		}
	}

	// Anclar al final
	regexPattern.WriteString("$")

	// Compilar regex con flag case-insensitive
	regex, err := regexp.Compile("(?i)" + regexPattern.String())
	if err != nil {
		return nil, fmt.Errorf("error compilando patrón: %w", err)
	}

	return regex, nil
}

// ================================================================
// FUNCIÓN ALTERNATIVA: ScanByGlob (alias para compatibilidad)
// ================================================================

// ScanByGlob es un alias de Find para compatibilidad con la interfaz existente
func (r *FileFsRepository) ScanByGlob(id string, basePath []string, pattern string, uid int, gid int) ([]string, error) {
	return r.Find(id, basePath, pattern, uid, gid)
}
