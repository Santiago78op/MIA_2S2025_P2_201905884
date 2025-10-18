package session

import "Backend/core/ports"

type memorySession struct {
	logged      bool
	user        string
	uid         int
	gid         int
	partitionId string // ID de la partición donde se hizo login
}

func NewMemory() ports.SessionStore {
	return &memorySession{}
}

func (m *memorySession) Login(user string, uid int, gid int, partitionId string) {
	m.logged = true
	m.user = user
	m.uid = uid
	m.gid = gid
	m.partitionId = partitionId
}

func (m *memorySession) Logout() {
	m.logged = false
	m.user = ""
	m.uid = 0
	m.gid = 0
	m.partitionId = ""
}

func (m *memorySession) Current() (bool, string, int, int, bool, string) {
	// root bypass si user == "root"
	return m.logged, m.user, m.uid, m.gid, m.logged && m.user == "root", m.partitionId
}
