package session

import "backend/core/ports"

type memorySession struct {
	logged bool
	user   string
	uid    int
	gid    int
}

func NewMemory() ports.SessionStore {
	return &memorySession{}
}

func (m *memorySession) Login(user string, uid int, gid int) {
	m.logged = true
	m.user = user
	m.uid = uid
	m.gid = gid
}

func (m *memorySession) Logout() {
	m.logged = false
	m.user = ""
	m.uid = 0
	m.gid = 0
}

func (m *memorySession) Current() (bool, string, int, int, bool) {
	// root bypass si user == "root"
	return m.logged, m.user, m.uid, m.gid, m.logged && m.user == "root"
}
