package journal

import (
	"context"
	"time"
)

// Entry — representación de alto nivel que usa tu FS.
type Entry struct {
	Op        string    // mkdir | mkfile | edit | remove | rename | copy | move | chown | chmod ...
	Path      string    // path destino; para rename/copy/move puedes loguear "from -> to" en Content o usar otro campo si extiendes
	Content   []byte    // payload (p.ej. contenido nuevo de un archivo)
	Timestamp time.Time // cuando se registra
}

// Store — contrato para la persistencia del journal.
type Store interface {
	// Append agrega una transacción; debe ser atómico.
	Append(ctx context.Context, partID string, e Entry) error

	// List devuelve las entradas en orden lógico (más antigua → más reciente).
	// Si hay rotación (circular buffer), debe recomponer el orden.
	List(ctx context.Context, partID string) ([]Entry, error)

	// Replay recorre las entradas en orden y llama a apply por cada una.
	// Si apply retorna error, corta y devuelve ese error.
	Replay(ctx context.Context, partID string, apply func(Entry) error) error

	// ClearAll borra el journal completo (útil para "loss" o re-formato).
	ClearAll(ctx context.Context, partID string) error
}
