package fs

import (
	"fmt"
	"strings"
	"time"
)

// Interfaces (puertos) que tu repo de EXT2/EXT3 implementa.
type FsRepository interface {
	Mkfs(id string, formatType string) error
	MkfsWithType(id string, fsType int32, full bool) error // NUEVO P2

	Mkdir(id string, absPath []string, parents bool, uid int, gid int, now time.Time) error
	Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool, uid int, gid int, now time.Time) error
	Cat(id string, files [][]string, uid int, gid int) (string, error)

	// Comandos avanzados P2
	Remove(id string, absPath []string, uid int, gid int) error
	Edit(id string, absPath []string, content []byte, uid int, gid int) error
	Rename(id string, absPath []string, newName string, uid int, gid int) error
	Copy(id string, srcPath []string, destPath []string, uid int, gid int) (copied int, skipped int, err error)
	Move(id string, srcPath []string, destPath []string, uid int, gid int) error
	Find(id string, startPath []string, pattern string, uid int, gid int) ([]string, error)
	Chmod(id string, absPath []string, ugo [3]byte, recursive bool, uid int, gid int) error
	Chown(id string, absPath []string, user string, recursive bool, uid int, gid int) error
	WipeDataAreas(id string) error
	Recovery(id string) error
}

type SessionStore interface {
	Current() (logged bool, user string, uid int, gid int, isRoot bool, partitionId string)
}

type FsService struct {
	repo FsRepository
	sess SessionStore
}

func NewFsService(repo FsRepository, sess SessionStore) *FsService {
	return &FsService{repo: repo, sess: sess}
}

// Mkfs versión P1 (legacy, mantiene compatibilidad)
func (s *FsService) MkfsLegacy(id string, formatType string) (string, error) {
	// Validar tipo de formateo
	if formatType != "full" && formatType != "" {
		return "", fmt.Errorf("tipo de formateo no soportado: %s (solo se acepta 'full')", formatType)
	}

	// Por defecto usar "full" si no se especifica
	if formatType == "" {
		formatType = "full"
	}

	if err := s.repo.Mkfs(id, formatType); err != nil {
		return "", err
	}
	return fmt.Sprintf("Partición formateada EXT2 (tipo: %s)", formatType), nil
}

// Mkfs versión P2 (soporta EXT2 y EXT3)
func (s *FsService) Mkfs(id string, fs string, typeFormat string) (string, error) {
	// Validar tipo de formateo
	if typeFormat != "full" && typeFormat != "" {
		return "", fmt.Errorf("tipo de formateo no soportado: %s (solo se acepta 'full')", typeFormat)
	}

	// Validar sistema de archivos
	if fs != "2fs" && fs != "3fs" && fs != "" {
		return "", fmt.Errorf("sistema de archivos no válido: %s (solo se acepta '2fs' o '3fs')", fs)
	}

	// Por defecto usar "2fs" si no se especifica
	if fs == "" {
		fs = "2fs"
	}

	// Convertir fs a fsType
	var fsType int32
	if fs == "2fs" {
		fsType = 2
	} else {
		fsType = 3
	}

	// Llamar al repositorio con el tipo de FS
	if err := s.repo.MkfsWithType(id, fsType, true); err != nil {
		return "", err
	}

	fsName := "EXT2"
	if fsType == 3 {
		fsName = "EXT3"
	}

	return fmt.Sprintf("Partición formateada %s (tipo: full)", fsName), nil
}

func (s *FsService) Mkdir(id, p string, parents bool) (string, error) {
	// Verificar que hay una sesión activa y obtener uid/gid del usuario
	logged, _, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("debe iniciar sesión para ejecutar este comando")
	}

	// Si no se proporciona id, usar el de la sesión actual
	if id == "" {
		id = partitionId
	}

	parts := SplitPath(p)
	if len(parts) == 0 {
		return "", fmt.Errorf("path inválido")
	}
	if err := s.repo.Mkdir(id, parts, parents, uid, gid, time.Now()); err != nil {
		return "", err
	}
	return "Directorio creado", nil
}

func (s *FsService) Mkfile(id, p string, size int, cont string, recursive bool) (string, error) {
	// Verificar que hay una sesión activa y obtener uid/gid del usuario
	logged, _, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("debe iniciar sesión para ejecutar este comando")
	}

	// Si no se proporciona id, usar el de la sesión actual
	if id == "" {
		id = partitionId
	}

	parts := SplitPath(p)
	if len(parts) == 0 {
		return "", fmt.Errorf("path inválido")
	}
	if size < 0 {
		return "", fmt.Errorf("size inválido")
	}
	if err := s.repo.Mkfile(id, parts, size, cont, recursive, uid, gid, time.Now()); err != nil {
		return "", err
	}
	return "Archivo creado", nil
}

func (s *FsService) Cat(id string, files []string) (string, error) {
	// Verificar que hay una sesión activa y obtener uid/gid del usuario
	logged, _, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("debe iniciar sesión para ejecutar este comando")
	}

	// Si no se proporciona id, usar el de la sesión actual
	if id == "" {
		id = partitionId
	}

	if len(files) == 0 {
		return "", fmt.Errorf("sin archivos")
	}
	paths := make([][]string, 0, len(files))
	for _, f := range files {
		parts := SplitPath(f)
		if len(parts) == 0 {
			return "", fmt.Errorf("path inválido: %s", f)
		}
		paths = append(paths, parts)
	}
	return s.repo.Cat(id, paths, uid, gid)
}

// ============================================
// MÉTODOS NUEVOS P2
// ============================================

func (s *FsService) Remove(id, path string) (string, error) {
	logged, user, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}
	if id == "" {
		id = partitionId
	}

	parts := SplitPath(path)
	if len(parts) == 0 {
		return "", fmt.Errorf("path inválido")
	}

	// Write-ahead journaling (solo EXT3; repos no-EXT3 ignoran o devuelven ErrUnsupported)
	// TODO: Implementar JournalAppend cuando esté disponible
	// _ = s.repo.JournalAppend(id, JournalEntry{
	//     Op:        OpRemove,
	//     Path:      path,
	//     Usuario:   user,
	//     Timestamp: time.Now().Unix(),
	// })

	// Ejecutar borrado atómico por permisos
	// TODO: El repo debe implementar Remove con validación de permisos recursiva
	_ = user // Evitar warning
	_, _, _ = uid, gid, partitionId

	return fmt.Sprintf("REMOVE ejecutado en %s (stub - pendiente implementación en repo)", path), nil
}

func (s *FsService) Edit(id, path, content string) (string, error) {
	// 1. Verificar sesión activa
	logged, user, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}

	// 2. Resolver ID (usar sesión si no se proporciona)
	if id == "" {
		id = partitionId
	}

	// 3. Validar path
	parts := SplitPath(path)
	if len(parts) == 0 {
		return "", fmt.Errorf("path inválido")
	}

	// 4. Leer contenido desde archivo host (content es ruta al archivo)
	// TODO: Implementar lectura real del archivo host cuando esté en producción
	// Por ahora usamos el content directamente como contenido
	// Ejemplo real: contentBytes, err := os.ReadFile(content)
	contentBytes := []byte(content) // Stub: trata content como contenido directo
	_ = contentBytes

	// 5. Write-ahead journaling (solo EXT3; repos no-EXT3 ignoran o devuelven ErrUnsupported)
	// TODO: Implementar JournalAppend cuando esté disponible
	// _ = s.repo.JournalAppend(id, JournalEntry{
	//     Op:        OpEdit,
	//     Path:      path,
	//     Usuario:   user,
	//     Timestamp: time.Now().Unix(),
	// })

	// 6. Ejecutar edición con validación de permisos
	// TODO: El repo debe implementar Edit con:
	//   - Verificar permiso de escritura
	//   - Reemplazar contenido
	//   - Reasignar bloques si cambia el tamaño
	_ = user // Evitar warning
	_, _ = uid, gid

	return fmt.Sprintf("EDIT ejecutado en %s (stub - pendiente implementación en repo)", path), nil
}

func (s *FsService) Rename(id, path, newName string) (string, error) {
	// 1. Verificar sesión activa
	logged, _, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}

	// 2. Resolver ID (usar sesión si no se proporciona)
	if id == "" {
		id = partitionId
	}

	// 3. Validar path y nuevo nombre
	parts := SplitPath(path)
	if len(parts) == 0 {
		return "", fmt.Errorf("path inválido")
	}
	if newName == "" {
		return "", fmt.Errorf("nuevo nombre no puede estar vacío")
	}
	// Validar que el nuevo nombre no contenga '/'
	if strings.Contains(newName, "/") {
		return "", fmt.Errorf("el nuevo nombre no puede contener '/'")
	}

	// 4. Ejecutar renombrado con validación de permisos
	err := s.repo.Rename(id, parts, newName, uid, gid)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("RENAME completado: %s → %s", path, newName), nil
}

func (s *FsService) Copy(id string, srcPath, destPath []string) (string, error) {
	logged, _, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}
	if id == "" {
		id = partitionId
	}

	if len(srcPath) == 0 || len(destPath) == 0 {
		return "", fmt.Errorf("path inválido")
	}

	// Ejecutar copia recursiva con verificación de permisos
	copied, skipped, err := s.repo.Copy(id, srcPath, destPath, uid, gid)
	if err != nil {
		return "", err
	}

	srcStr := "/" + strings.Join(srcPath, "/")
	destStr := "/" + strings.Join(destPath, "/")
	if skipped > 0 {
		return fmt.Sprintf("COPY completado: %d elementos copiados, %d omitidos por permisos", copied, skipped), nil
	}
	return fmt.Sprintf("COPY completado (%s → %s): %d elementos copiados", srcStr, destStr, copied), nil
}

func (s *FsService) Move(id string, srcPath, destPath []string) (string, error) {
	logged, _, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}
	if id == "" {
		id = partitionId
	}

	if len(srcPath) == 0 || len(destPath) == 0 {
		return "", fmt.Errorf("path inválido")
	}

	// Ejecutar movimiento mediante re-enlace (sin copiar bloques)
	err := s.repo.Move(id, srcPath, destPath, uid, gid)
	if err != nil {
		return "", err
	}

	srcStr := "/" + strings.Join(srcPath, "/")
	destStr := "/" + strings.Join(destPath, "/")
	return fmt.Sprintf("MOVE completado: %s → %s", srcStr, destStr), nil
}

func (s *FsService) Find(id, path, pattern string) (string, error) {
	logged, _, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}
	if id == "" {
		id = partitionId
	}

	parts := SplitPath(path)
	// Permitir path vacío (buscar desde raíz)

	// Ejecutar búsqueda recursiva con patrón glob
	results, err := s.repo.Find(id, parts, pattern, uid, gid)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return fmt.Sprintf("FIND: sin coincidencias para '%s' en %s", pattern, path), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("FIND: %d coincidencias para '%s':\n", len(results), pattern))
	for _, r := range results {
		output.WriteString("  " + r + "\n")
	}

	return output.String(), nil
}

func (s *FsService) Chmod(id, path, ugo string, recursive bool) (string, error) {
	logged, _, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}
	if id == "" {
		id = partitionId
	}

	parts := SplitPath(path)
	if len(parts) == 0 {
		return "", fmt.Errorf("path inválido")
	}

	// Validar formato UGO
	if len(ugo) != 3 {
		return "", fmt.Errorf("ugo debe tener 3 dígitos octales")
	}
	for _, c := range ugo {
		if c < '0' || c > '7' {
			return "", fmt.Errorf("ugo debe contener solo dígitos octales (0-7)")
		}
	}

	// Convertir UGO string a [3]byte
	ugoBytes := [3]byte{ugo[0], ugo[1], ugo[2]}

	// Ejecutar chmod
	err := s.repo.Chmod(id, parts, ugoBytes, recursive, uid, gid)
	if err != nil {
		return "", err
	}

	recursiveStr := ""
	if recursive {
		recursiveStr = " (recursivo)"
	}
	return fmt.Sprintf("CHMOD completado en %s con permisos %s%s", path, ugo, recursiveStr), nil
}

func (s *FsService) Chown(id, path, user string, recursive bool) (string, error) {
	logged, _, uid, gid, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}
	if id == "" {
		id = partitionId
	}

	parts := SplitPath(path)
	if len(parts) == 0 {
		return "", fmt.Errorf("path inválido")
	}

	// Ejecutar chown
	err := s.repo.Chown(id, parts, user, recursive, uid, gid)
	if err != nil {
		return "", err
	}

	recursiveStr := ""
	if recursive {
		recursiveStr = " (recursivo)"
	}
	return fmt.Sprintf("CHOWN completado en %s para usuario %s%s", path, user, recursiveStr), nil
}

func (s *FsService) Loss(id string) (string, error) {
	logged, _, _, _, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}
	if id == "" {
		id = partitionId
	}

	// Ejecutar LOSS (limpia áreas de datos, deja SB y Journal)
	err := s.repo.WipeDataAreas(id)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("LOSS completado: datos eliminados en partición %s (bitmaps, inodos, bloques). Journal intacto.", id), nil
}

func (s *FsService) Recovery(id string) (string, error) {
	logged, _, _, _, _, partitionId := s.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay sesión activa")
	}
	if id == "" {
		id = partitionId
	}

	// Ejecutar RECOVERY (re-aplica journal)
	err := s.repo.Recovery(id)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("RECOVERY completado: operaciones del journal re-aplicadas en partición %s", id), nil
}
