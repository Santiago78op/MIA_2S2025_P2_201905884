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

	// 2. Validar el comando
	if err := handler.Validate(); err != nil {
		return "", fmt.Errorf("%v\n\n%s", err, Usage(handler.Name()))
	}

	// 3. Ejecutar el comando
	return handler.Execute(ctx, a)
}

// pickFS selecciona el filesystem apropiado basado en el handle
// Consulta el superbloque/metadatos para determinar si es 2fs o 3fs
func (a *Adapter) pickFS(h fs.MountHandle) fs.FS {
	// Construir el ID de montaje a partir del handle
	mountID := h.PartitionID

	// Consultar el estado de metadatos
	if a.State != nil {
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

	// Fallback: si no hay metadata o no se pudo determinar,
	// usar FS3 si existe, caso contrario FS2
	if a.FS3 != nil {
		return a.FS3
	}
	return a.FS2
}
