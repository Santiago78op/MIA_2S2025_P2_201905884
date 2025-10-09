package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MIA_2S2025_P2_201905884/internal/fs"
)

// Session representa una sesión activa de usuario
type Session struct {
	User      string
	MountID   string
	Timestamp time.Time
}

// SessionManager gestiona la sesión activa (solo una a la vez en P1)
type SessionManager struct {
	mu      sync.RWMutex
	current *Session
	fs      fs.FS // Para validar credenciales en /users.txt
}

// NewSessionManager crea un nuevo gestor de sesiones
func NewSessionManager(filesystem fs.FS) *SessionManager {
	return &SessionManager{
		fs: filesystem,
	}
}

// IsActive verifica si hay una sesión activa
func (sm *SessionManager) IsActive() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current != nil
}

// Login inicia sesión para un usuario
func (sm *SessionManager) Login(ctx context.Context, user, pass, mountID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validar que no hay sesión activa
	if sm.current != nil {
		return fmt.Errorf("ERROR SESION INICIADA")
	}

	// TODO: Validar credenciales contra /users.txt en el FS montado
	// Por ahora, aceptar cualquier credencial para desarrollo
	// En producción, debería:
	// 1. Leer /users.txt del FS montado
	// 2. Buscar línea con formato: <gid>,U,<user>,<grp>,<pass>
	// 3. Verificar que pass coincide

	sm.current = &Session{
		User:      user,
		MountID:   mountID,
		Timestamp: time.Now(),
	}

	return nil
}

// Logout cierra la sesión activa
func (sm *SessionManager) Logout() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.current = nil
}

// CurrentUser retorna el usuario de la sesión activa
func (sm *SessionManager) CurrentUser() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if sm.current == nil {
		return ""
	}
	return sm.current.User
}

// CurrentMountID retorna el ID de montaje de la sesión activa
func (sm *SessionManager) CurrentMountID() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if sm.current == nil {
		return ""
	}
	return sm.current.MountID
}
