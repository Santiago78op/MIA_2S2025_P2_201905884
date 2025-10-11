package commands

import (
	"context"
	"fmt"

	"MIA_2S2025_P2_201905884/internal/disk"
	"MIA_2S2025_P2_201905884/internal/fs"
	"MIA_2S2025_P2_201905884/internal/reports"
)

// SessionManager define la interfaz para gestión de sesiones
type SessionManager interface {
	IsActive() bool
	Login(ctx context.Context, user, pass, mountID string) error
	Logout()
	CurrentUser() string
	CurrentMountID() string
}

// Adapter conecta el parser/validador de comandos con los servicios reales.
type Adapter struct {
	FS2     fs.FS              // implementación EXT2
	FS3     fs.FS              // implementación EXT3
	DM      disk.Manager       // gestor de discos/particiones
	Index   MountIndex         // índice de montajes (id -> refs/handles)
	State   *fs.MetaState      // estado de metadatos de filesystems
	Session SessionManager     // gestor de sesiones
	Reports reports.Generator  // generador de reportes
}

// Run ejecuta una línea de comando (parsea, valida y ejecuta).
func (a *Adapter) Run(ctx context.Context, line string) (string, error) {
	// 1. Parsear el comando
	handler, err := ParseCommand(line)
	if err != nil {
		return "", err
	}

	// 2. Inyectar ID de sesión si no está presente y hay sesión activa
	if a.Session != nil && a.Session.IsActive() {
		a.injectSessionID(handler)
	}

	// 3. Validar el comando
	if err := handler.Validate(); err != nil {
		return "", fmt.Errorf("%v\n\n%s", err, Usage(handler.Name()))
	}

	// 4. Ejecutar el comando
	return handler.Execute(ctx, a)
}

// injectSessionID inyecta el ID de sesión activa en comandos que lo requieren
// si no tienen uno especificado
func (a *Adapter) injectSessionID(handler CommandHandler) {
	sessionID := a.Session.CurrentMountID()
	if sessionID == "" {
		return
	}

	// Inyectar ID en comandos de archivo que lo requieren
	switch cmd := handler.(type) {
	case *MkdirCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *MkfileCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *RemoveCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *EditCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *RenameCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *CopyCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *MoveCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *FindCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *ChownCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *ChmodCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *JournalingCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *RecoveryCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *LossCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	case *MkfsCommand:
		if cmd.ID == "" {
			cmd.ID = sessionID
		}
	}
}

// pickFS selecciona el filesystem apropiado basado en el handle
// Consulta el superbloque/metadatos para determinar si es 2fs o 3fs
func (a *Adapter) pickFS(h fs.MountHandle) fs.FS {
	// Buscar el ID de montaje completo en el índice
	// (ejemplo: 841A en lugar de solo Part1)
	var mountID string
	ids := a.Index.List()
	for _, id := range ids {
		ref, ok := a.Index.GetRef(id)
		if ok && ref.DiskPath == h.DiskID && ref.PartitionID == h.PartitionID {
			mountID = id
			break
		}
	}

	// Consultar el estado de metadatos
	if mountID != "" && a.State != nil {
		meta, ok := a.State.Get(mountID)
		if ok {
			// Seleccionar según el tipo de filesystem guardado
			switch meta.FSKind {
			case "3fs":
				if a.FS3 != nil {
					return a.FS3
				}
			case "2fs":
				if a.FS2 != nil {
					return a.FS2
				}
			}
		}
	}

	// Fallback: usar FS2 por defecto (más compatible con P1)
	return a.FS2
}
