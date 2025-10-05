package disk

import "errors"

var (
	ErrInvalidParam   = errors.New("disk: parámetro inválido")
	ErrNotFound       = errors.New("disk: no encontrado")
	ErrExists         = errors.New("disk: ya existe")
	ErrNoSpace        = errors.New("disk: sin espacio suficiente")
	ErrBadLayout      = errors.New("disk: layout inválido")
	ErrUnsupported    = errors.New("disk: operación no soportada")
	ErrNotMounted     = errors.New("disk: no montado")
	ErrAlreadyMounted = errors.New("disk: ya montado")
	ErrNoExtended     = errors.New("disk: no existe partición extendida")
)
