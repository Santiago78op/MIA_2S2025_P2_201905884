package utils

import "errors"

// Sentencias de error comunes para la app
var (
	ErrOutOfBounds      = errors.New("operación fuera de los límites de la partición")
	ErrNoSpace          = errors.New("espacio insuficiente")
	ErrDuplicateName    = errors.New("nombre duplicado")
	ErrNotFound         = errors.New("no encontrado")
	ErrAlreadyExists    = errors.New("ya existe")
	ErrPathInvalid      = errors.New("path inválido")
	ErrIsDirectory      = errors.New("es un directorio")
	ErrIsFile           = errors.New("es un archivo")
	ErrPermissionDenied = errors.New("permiso denegado")
	ErrLoginRequired    = errors.New("requiere sesión iniciada")
	ErrOnlyRoot         = errors.New("solo root puede ejecutar esta operación")
	ErrNotMounted       = errors.New("id no montado")
)
