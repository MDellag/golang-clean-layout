package types

import (
	"context"
	"time"
)

// Job representa una tarea que debe ser ejecutada de forma asíncrona
type Job interface {
	Execute(ctx context.Context) error
	Name() string
	Priority() int
}

// JobResult representa el resultado de la ejecución de un job
type JobResult struct {
	JobID     string
	JobName   string
	Success   bool
	Error     error
	Data      interface{}
	Duration  time.Duration
	Timestamp time.Time
}

// JobWithResponse extiende Job para incluir canal de respuesta
type JobWithResponse interface {
	Job
	// ResponseChannel devuelve el canal donde enviar el resultado
	ResponseChannel() chan<- JobResult
	// ID devuelve un identificador único para el job
	ID() string
}