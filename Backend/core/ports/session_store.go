package ports

// Sesión simple: usuario logueado y si es root
type SessionStore interface {
	Login(user string, uid int, gid int)
	Logout()
	Current() (logged bool, user string, uid int, gid int, isRoot bool)
}
