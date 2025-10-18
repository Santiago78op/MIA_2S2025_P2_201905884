package users

import "fmt"

// UserService opera sobre users.txt en la partición montada (id).

type FsUsersRepository interface {
	Login(id, user, pass string) (uid, gid int, isRoot bool, err error)
	Mkgrp(id, name string) error
	Rmgrp(id, name string) error
	Mkusr(id, user, pass, grp string) error
	Rmusr(id, user string) error
	Chgrp(id, user, grp string) error
}

type SessionStore interface {
	Login(user string, uid int, gid int, partitionId string)
	Logout()
	Current() (logged bool, user string, uid int, gid int, isRoot bool, partitionId string)
}

type UserService struct {
	repo FsUsersRepository
	sess SessionStore
}

func NewUserService(repo FsUsersRepository, sess SessionStore) *UserService {
	return &UserService{repo: repo, sess: sess}
}

func (u *UserService) Login(id, user, pass string) (string, error) {
	// Verificar si ya existe una sesión activa
	logged, _, _, _, _, _ := u.sess.Current()
	if logged {
		return "", fmt.Errorf("ya existe una sesión activa, debe cerrar sesión primero")
	}

	uid, gid, isRoot, err := u.repo.Login(id, user, pass)
	if err != nil {
		return "", err
	}
	u.sess.Login(user, uid, gid, id) // Guardar también el ID de la partición
	if isRoot {
		return "login (root)", nil
	}
	return "login", nil
}

func (u *UserService) Logout() (string, error) {
	// Verificar que hay una sesión activa antes de cerrarla
	logged, _, _, _, _, _ := u.sess.Current()
	if !logged {
		return "", fmt.Errorf("no hay ninguna sesión activa")
	}

	u.sess.Logout()
	return "logout", nil
}

func (u *UserService) Mkgrp(id, name string) (string, error) {
	// Verificar que hay una sesión activa
	logged, _, _, _, isRoot, partitionId := u.sess.Current()
	if !logged {
		return "", fmt.Errorf("debe iniciar sesión para ejecutar este comando")
	}

	// Verificar que el usuario es root
	if !isRoot {
		return "", fmt.Errorf("solo el usuario root puede crear grupos")
	}

	// Si no se proporciona id, usar el de la sesión actual
	if id == "" {
		id = partitionId
	}

	if err := u.repo.Mkgrp(id, name); err != nil {
		return "", err
	}
	return "grupo creado", nil
}

func (u *UserService) Rmgrp(id, name string) (string, error) {
	// Verificar que hay una sesión activa
	logged, _, _, _, isRoot, partitionId := u.sess.Current()
	if !logged {
		return "", fmt.Errorf("debe iniciar sesión para ejecutar este comando")
	}

	// Verificar que el usuario es root
	if !isRoot {
		return "", fmt.Errorf("solo el usuario root puede eliminar grupos")
	}

	// Si no se proporciona id, usar el de la sesión actual
	if id == "" {
		id = partitionId
	}

	if err := u.repo.Rmgrp(id, name); err != nil {
		return "", err
	}
	return "grupo eliminado", nil
}

func (u *UserService) Mkusr(id, user, pass, grp string) (string, error) {
	// Verificar que hay una sesión activa
	logged, _, _, _, isRoot, partitionId := u.sess.Current()
	if !logged {
		return "", fmt.Errorf("debe iniciar sesión para ejecutar este comando")
	}

	// Verificar que el usuario es root
	if !isRoot {
		return "", fmt.Errorf("solo el usuario root puede crear usuarios")
	}

	// Si no se proporciona id, usar el de la sesión actual
	if id == "" {
		id = partitionId
	}

	if err := u.repo.Mkusr(id, user, pass, grp); err != nil {
		return "", err
	}
	return "usuario creado", nil
}

func (u *UserService) Rmusr(id, user string) (string, error) {
	// Verificar que hay una sesión activa
	logged, _, _, _, isRoot, partitionId := u.sess.Current()
	if !logged {
		return "", fmt.Errorf("debe iniciar sesión para ejecutar este comando")
	}

	// Verificar que el usuario es root
	if !isRoot {
		return "", fmt.Errorf("solo el usuario root puede eliminar usuarios")
	}

	// Si no se proporciona id, usar el de la sesión actual
	if id == "" {
		id = partitionId
	}

	if err := u.repo.Rmusr(id, user); err != nil {
		return "", err
	}
	return "usuario eliminado", nil
}

func (u *UserService) Chgrp(id, user, grp string) (string, error) {
	// Verificar que hay una sesión activa
	logged, _, _, _, isRoot, partitionId := u.sess.Current()
	if !logged {
		return "", fmt.Errorf("debe iniciar sesión para ejecutar este comando")
	}

	// Verificar que el usuario es root
	if !isRoot {
		return "", fmt.Errorf("solo el usuario root puede cambiar grupos de usuarios")
	}

	// Si no se proporciona id, usar el de la sesión actual
	if id == "" {
		id = partitionId
	}

	if err := u.repo.Chgrp(id, user, grp); err != nil {
		return "", err
	}
	return "grupo cambiado", nil
}
