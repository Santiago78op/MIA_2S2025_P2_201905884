package ports

// Sesión simple: usuario logueado, si es root, y partición activa
type SessionStore interface {
	Login(user string, uid int, gid int, partitionId string)
	Logout()
	Current() (logged bool, user string, uid int, gid int, isRoot bool, partitionId string)
}
