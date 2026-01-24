package types

import (
	"context"
	"sync/atomic"
)

// GenericJob representa un job que produce un resultado de tipo T
type GenericJob[T any] interface {
	// Execute procesa el job y retorna el resultado de tipo T
	Execute(ctx context.Context) (T, error)
	// Name retorna el nombre del job para identificación
	Name() string
	// Priority retorna la prioridad del job (mayor = más prioritario)
	Priority() int
}

// GenericJobResult representa el resultado de la ejecución de un job genérico
type GenericJobResult[T any] struct {
	JobName string
	Success bool
	Result  T
	Error   error
}

// GenericJobWrapper envuelve un GenericJob para que pueda ser procesado por el dispatcher
type GenericJobWrapper[T any] struct {
	job             GenericJob[T]
	responseChannel chan<- GenericJobResult[T]
	completedJobs   *atomic.Int64 // Contador opcional para sincronización
}

// NewGenericJobWrapper crea un nuevo wrapper para un job genérico
func NewGenericJobWrapper[T any](job GenericJob[T], responseChan chan<- GenericJobResult[T]) *GenericJobWrapper[T] {
	return &GenericJobWrapper[T]{
		job:             job,
		responseChannel: responseChan,
		completedJobs:   nil, // Sin contador
	}
}

// NewGenericJobWrapperWithCounter crea un wrapper con contador atómico para sincronización
func NewGenericJobWrapperWithCounter[T any](job GenericJob[T], responseChan chan<- GenericJobResult[T], counter *atomic.Int64) *GenericJobWrapper[T] {
	return &GenericJobWrapper[T]{
		job:             job,
		responseChannel: responseChan,
		completedJobs:   counter,
	}
}

// Execute implementa la interfaz Job
func (w *GenericJobWrapper[T]) Execute(ctx context.Context) error {
	result, err := w.job.Execute(ctx)

	jobResult := GenericJobResult[T]{
		JobName: w.job.Name(),
		Success: err == nil,
		Result:  result,
		Error:   err,
	}

	// Enviar el resultado al canal
	select {
	case w.responseChannel <- jobResult:
		// Resultado enviado exitosamente
		// Incrementar contador si está disponible
		if w.completedJobs != nil {
			w.completedJobs.Add(1)
		}
	case <-ctx.Done():
		// Contexto cancelado
		return ctx.Err()
	}

	return err
}

// Name implementa la interfaz Job
func (w *GenericJobWrapper[T]) Name() string {
	return w.job.Name()
}

// Priority implementa la interfaz Job
func (w *GenericJobWrapper[T]) Priority() int {
	return w.job.Priority()
}
