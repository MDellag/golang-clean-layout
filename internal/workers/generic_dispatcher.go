package worker

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"clean-arq-layout/internal/workers/types"
)

// newTicker crea un nuevo ticker - wrapper para facilitar testing
func newTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}

// GenericDispatcher permite procesar jobs genéricos y recibir respuestas tipadas
type GenericDispatcher[T any] struct {
	pool            *Pool
	ctx             context.Context
	cancel          context.CancelFunc
	responseChannel chan types.GenericJobResult[T]
	completedJobs   atomic.Int64
}

// NewGenericDispatcher crea un nuevo dispatcher genérico
// maxWorkers: número de workers a crear
// bufferSize: tamaño del buffer del canal de respuestas
func NewGenericDispatcher[T any](maxWorkers int, bufferSize int) *GenericDispatcher[T] {
	ctx, cancel := context.WithCancel(context.Background())

	return &GenericDispatcher[T]{
		pool:            NewWorkerPool(maxWorkers, maxWorkers*2),
		ctx:             ctx,
		cancel:          cancel,
		responseChannel: make(chan types.GenericJobResult[T], bufferSize),
	}
}

// Dispatch procesa un slice de jobs y retorna el canal de respuestas
// Los resultados se enviarán al canal a medida que los jobs se completen
func (gd *GenericDispatcher[T]) Dispatch(jobs []types.GenericJob[T]) (<-chan types.GenericJobResult[T], error) {
	// Iniciar el pool si no está iniciado
	if err := gd.pool.Start(); err != nil {
		return nil, fmt.Errorf("failed to start worker pool: %w", err)
	}

	// Resetear contador
	gd.completedJobs.Store(0)

	// Enviar todos los jobs al pool
	for _, job := range jobs {
		// Envolver el job genérico con el wrapper, pasando el contador atómico
		wrapper := types.NewGenericJobWrapperWithCounter(job, gd.responseChannel, &gd.completedJobs)

		// Enviar el wrapper al pool
		if err := gd.pool.Submit(wrapper); err != nil {
			log.Printf("Failed to submit job %s: %v", job.Name(), err)
			// Continuar con los demás jobs incluso si uno falla
		}
	}

	// Iniciar goroutine para cerrar el canal cuando todos los jobs terminen
	go gd.waitAndClose(len(jobs))

	return gd.responseChannel, nil
}

// waitAndClose espera a que todos los jobs se procesen y cierra el canal de respuestas
func (gd *GenericDispatcher[T]) waitAndClose(totalJobs int) {
	// Usar un ticker para verificar periódicamente sin consumir del canal
	ticker := newTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-gd.ctx.Done():
			log.Println("GenericDispatcher.waitAndClose: context canceled, closing response channel")
			close(gd.responseChannel)
			return

		case <-ticker.C:
			// Verificar el contador atómico sin consumir mensajes del canal
			completed := gd.completedJobs.Load()
			if completed >= int64(totalJobs) {
				log.Printf("GenericDispatcher.waitAndClose: all %d jobs completed, closing response channel", totalJobs)
				close(gd.responseChannel)
				return
			}
		}
	}
}

// Stop detiene el dispatcher y limpia recursos
func (gd *GenericDispatcher[T]) Stop() {
	gd.cancel()
	gd.pool.Stop()
}

// DispatchAndWait es una función de conveniencia que procesa jobs y espera a que todos terminen
// Retorna un slice con todos los resultados
func (gd *GenericDispatcher[T]) DispatchAndWait(ctx context.Context, jobs []types.GenericJob[T]) ([]types.GenericJobResult[T], error) {
	resultChan, err := gd.Dispatch(jobs)
	if err != nil {
		return nil, err
	}

	results := make([]types.GenericJobResult[T], 0, len(jobs))

	for result := range resultChan {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
			results = append(results, result)
		}
	}

	return results, nil
}
