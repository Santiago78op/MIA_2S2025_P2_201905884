package journal

import "errors"

var (
	ErrNotFound  = errors.New("journal: no existe")
	ErrCorrupted = errors.New("journal: archivo corrupto")
	ErrInvalidOp = errors.New("journal: operación inválida")
)
