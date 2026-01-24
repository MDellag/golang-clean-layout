package types

import (
	"context"
)

// Job representa una tarea que debe ser ejecutada de forma asíncrona
// Esta es la interfaz base que deben implementar todos los jobs.
// Para jobs con resultados tipados, usa GenericJob[T] en su lugar.
type Job interface {
	Execute(ctx context.Context) error
	Name() string
	Priority() int
}