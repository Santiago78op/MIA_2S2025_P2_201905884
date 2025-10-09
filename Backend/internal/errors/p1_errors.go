package errors

import "errors"

// Errores P1 - TEXTOS EXACTOS requeridos por el calificador
// NO modificar estos textos, el calificador espera estos mensajes exactos

var (
	// Parámetros y validación
	ErrParams = errors.New("ERROR PARAMETROS")

	// Rutas y archivos
	ErrPathNotFound     = errors.New("ERROR RUTA NO ENCONTRADA")
	ErrPathDoesNotExist = errors.New("ERROR NO EXISTE RUTA")
	ErrDiskNotExist     = errors.New("ERROR DISCO NO EXISTE")

	// Particiones y discos
	ErrAlreadyExists     = errors.New("ERROR YA EXISTE")
	ErrPartitionLimit    = errors.New("ERROR LIMITE PARTICION")
	ErrNoSpace           = errors.New("ERROR FALTA ESPACIO")
	ErrAlreadyMounted    = errors.New("ERROR PARTICION YA MONTADA")
	ErrPartitionNotFound = errors.New("ERROR PARTICION NO EXISTE")
	ErrIDNotFound        = errors.New("ERROR ID NO ENCONTRADO")

	// Sesión
	ErrNoSession     = errors.New("ERROR NO HAY SESION INICIADA")
	ErrSessionExists = errors.New("ERROR SESION INICIADA")

	// Usuarios y grupos
	ErrGroupExists    = errors.New("ERROR YA EXISTE EL GRUPO")
	ErrUserExists     = errors.New("ERROR EL USUARIO YA EXISTE")
	ErrGroupNotExist  = errors.New("ERROR GRUPO NO EXISTE")
	ErrUserNotExist   = errors.New("ERROR USUARIO NO EXISTE")
	ErrInvalidCredentials = errors.New("ERROR CREDENCIALES INVALIDAS")

	// Archivos y directorios
	ErrNoParentFolders = errors.New("ERROR NO EXISTEN LAS CARPETAS PADRES")
	ErrFileNotFound    = errors.New("ERROR ARCHIVO NO ENCONTRADO")
	ErrDirNotFound     = errors.New("ERROR DIRECTORIO NO ENCONTRADO")

	// Validaciones numéricas
	ErrNegative = errors.New("ERROR NEGATIVO")
	ErrInvalidSize = errors.New("ERROR TAMAÑO INVALIDO")
)

// IsP1Error verifica si un error es uno de los errores P1 estándar
func IsP1Error(err error) bool {
	if err == nil {
		return false
	}

	p1Errors := []error{
		ErrParams,
		ErrPathNotFound,
		ErrPathDoesNotExist,
		ErrDiskNotExist,
		ErrAlreadyExists,
		ErrPartitionLimit,
		ErrNoSpace,
		ErrAlreadyMounted,
		ErrPartitionNotFound,
		ErrIDNotFound,
		ErrNoSession,
		ErrSessionExists,
		ErrGroupExists,
		ErrUserExists,
		ErrGroupNotExist,
		ErrUserNotExist,
		ErrNoParentFolders,
		ErrFileNotFound,
		ErrDirNotFound,
		ErrNegative,
		ErrInvalidSize,
		ErrInvalidCredentials,
	}

	for _, p1Err := range p1Errors {
		if errors.Is(err, p1Err) {
			return true
		}
	}

	return false
}
