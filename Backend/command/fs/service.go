package fs

import (
	"fmt"
	"time"
)

// Interfaces (puertos) que tu repo de EXT2 implementa.
type FsRepository interface {
	Mkfs(id string, formatType string) error

	Mkdir(id string, absPath []string, parents bool, uid int, gid int, now time.Time) error
	Mkfile(id string, absPath []string, size int, contentHostPath string, recursive bool, uid int, gid int, now time.Time) error
	Cat(id string, files [][]string, uid int, gid int) (string, error)
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

func (s *FsService) Mkfs(id string, formatType string) (string, error) {
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
