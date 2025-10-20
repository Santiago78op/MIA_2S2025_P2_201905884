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

func (a *SessionAdapter) Login(user string, uid int, gid int, partitionId string) {
	a.sess.Login(user, uid, gid, partitionId)
}

func (a *SessionAdapter) Logout() {
	a.sess.Logout()
}

func (a *SessionAdapter) Current() (logged bool, user string, uid int, gid int, isRoot bool, partitionId string) {
	return a.sess.Current()
}

// Helper para crear directamente desde memoria
func NewSessionAdapterFromMemory() interface {
	users.SessionStore
	fs.SessionStore
} {
	return &SessionAdapter{sess: session.NewMemory()}
}

// NewPortsSessionAdapter adapta SessionAdapter para exponer ports.SessionStore
func NewPortsSessionAdapter(adapter interface {
	users.SessionStore
	fs.SessionStore
}) ports.SessionStore {
	// Si es un SessionAdapter, extraer el ports.SessionStore interno
	if sa, ok := adapter.(*SessionAdapter); ok {
		return sa.sess
	}
	// Si no, crear uno nuevo desde memoria como fallback
	return session.NewMemory()
}
