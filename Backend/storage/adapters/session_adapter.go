package adapters

import (
	"Backend/command/fs"
	"Backend/command/users"
	"Backend/core/ports"
	"Backend/storage/session"
)

// SessionAdapter adapta session para cumplir con users.SessionStore y fs.SessionStore
type SessionAdapter struct {
	sess ports.SessionStore
}

func NewSessionAdapter(sess ports.SessionStore) interface {
	users.SessionStore
	fs.SessionStore
} {
	return &SessionAdapter{sess: sess}
}

func (a *SessionAdapter) Login(user string, uid int, gid int) {
	isRoot := user == "root"
	a.sess.Login(user, uid, gid)
	_ = isRoot // La interfaz ports.SessionStore no requiere isRoot en Login
}

func (a *SessionAdapter) Logout() {
	a.sess.Logout()
}

func (a *SessionAdapter) Current() (logged bool, user string, uid int, gid int, isRoot bool) {
	return a.sess.Current()
}

// Helper para crear directamente desde memoria
func NewSessionAdapterFromMemory() interface {
	users.SessionStore
	fs.SessionStore
} {
	return &SessionAdapter{sess: session.NewMemory()}
}
