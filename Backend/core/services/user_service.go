package services

type UserService interface {
	Login(id, user, pass string) (string, error)
	Logout() string
	Mkgrp(id, name string) (string, error)
	Rmgrp(id, name string) (string, error)
	Mkusr(id, user, pass, grp string) (string, error)
	Rmusr(id, user string) (string, error)
	Chgrp(id, user, grp string) (string, error)
}
