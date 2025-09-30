package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"clean-arq-layout/internal/workers/types"
)

// Job es un alias para mantener compatibilidad
type Job = types.Job

// Dispatcher coordina y distribuye trabajos entre trabajadores
type Dispatcher struct {
	workerPool *Pool
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
	started    bool
}

// NewDispatcher crea un nuevo dispatcher con un pool de workers
func NewDispatcher(ctx context.Context, maxWorkers int, queueSize int) *Dispatcher {
	dispatcherCtx, cancel := context.WithCancel(ctx)

	return &Dispatcher{
		workerPool: NewWorkerPool(maxWorkers, queueSize),
		ctx:        dispatcherCtx,
		cancel:     cancel,
	}
}

// Start inicia el dispatcher y sus workers
func (d *Dispatcher) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.started {
		return fmt.Errorf("dispatcher already started")
	}

	// Iniciar el pool de workers
	if err := d.workerPool.Start(); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	// Iniciar rutina de monitoreo usando waitgroup.Go
	d.wg.Go(d.monitor)

	d.started = true
	log.Println("Dispatcher started successfully")
	return nil
}

// EnqueueJob encola un trabajo para su procesamiento
func (d *Dispatcher) EnqueueJob(job Job) error {
	if !d.started {
		return fmt.Errorf("dispatcher not started")
	}

	return d.workerPool.Submit(job)
}

// Stop detiene el dispatcher y todos sus workers de manera ordenada
func (d *Dispatcher) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.started {
		return
	}

	log.Println("Stopping dispatcher...")

	// Señalar cancelación
	d.cancel()

	// Detener el pool de workers
	d.workerPool.Stop()

	// Esperar a que todas las goroutines de monitoreo terminen
	d.wg.Wait()

	d.started = false
	log.Println("Dispatcher stopped")
}

// monitor es una rutina que podría monitorear el estado del sistema
// y proporcionar métricas o ajustar dinámicamente el tamaño del pool
func (d *Dispatcher) monitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pending := d.workerPool.Pending()
			log.Printf("Worker stats - Workers: %d, Pending jobs: %d",
				d.workerPool.Size(), pending)

			// Aquí podrías implementar lógica para escalar el número de workers
			// basándote en la carga actual

		case <-d.ctx.Done():
			log.Println("Dispatcher monitor shutting down")
			return
		}
	}
}

func (d *Dispatcher) Stats() map[string]interface{} {
	return map[string]interface{}{
		"workers":      d.workerPool.Size(),
		"pending_jobs": d.workerPool.Pending(),
		"is_running":   d.started,
	}
}
