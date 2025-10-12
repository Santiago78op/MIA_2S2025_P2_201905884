package models

// Representación en memoria (útil para parsear/crear users.txt)
type UserRecord struct {
	UID   int
	Kind  byte   // 'G' o 'U'
	Group string // grupo (para G y U)
	User  string // solo para U
	Pass  string // solo para U
	// puedes añadir un flag "Deleted" si usas bajas lógicas
}

// Convenciones del proyecto para el contenido inicial:
const (
	UsersBootGroup = "1,G,root"
	UsersBootUser  = "1,U,root,root,123"
)
