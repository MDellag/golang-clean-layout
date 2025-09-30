package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Pool maneja un conjunto de workers para procesar jobs
type Pool struct {
	workers    []*Worker
	workerPool chan chan Job
	maxWorkers int
	jobQueue   chan Job
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	started    bool
}

// NewWorkerPool crea un nuevo pool de workers
func NewWorkerPool(maxWorkers int, jobQueueSize int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		maxWorkers: maxWorkers,
		workerPool: make(chan chan Job, maxWorkers),
		jobQueue:   make(chan Job, jobQueueSize),
		workers:    make([]*Worker, 0, maxWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start inicia el pool de workers
func (p *Pool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return fmt.Errorf("worker pool already started")
	}

	// Inicializar y arrancar los workers usando waitgroup.Go
	for i := 0; i < p.maxWorkers; i++ {
		worker := NewWorker(i, p.workerPool, p.ctx)
		p.workers = append(p.workers, worker)
		p.wg.Go(worker.run)
	}

	// Iniciar el distribuidor de trabajos usando waitgroup.Go
	p.wg.Go(p.dispatch)

	p.started = true
	log.Printf("Worker pool started with %d workers", p.maxWorkers)
	return nil
}

// dispatch distribuye los trabajos entre los workers disponibles
func (p *Pool) dispatch() {
	for {
		select {
		case <-p.ctx.Done():
			log.Println("Dispatcher shutting down")
			return

		case job := <-p.jobQueue:
			// Esperar un worker disponible del pool
			select {
			case jobChannel := <-p.workerPool:
				// Enviar el trabajo al worker
				select {
				case jobChannel <- job:
					// Trabajo enviado al worker
				case <-p.ctx.Done():
					return
				}

			case <-p.ctx.Done():
				return
			}
		}
	}
}

// Submit encola un nuevo trabajo para ser procesado
func (p *Pool) Submit(job Job) error {
	if !p.started {
		return fmt.Errorf("worker pool not started")
	}

	select {
	case p.jobQueue <- job:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	}
}

// Stop detiene el pool y todos sus workers de manera ordenada
func (p *Pool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return
	}

	log.Println("Stopping worker pool...")
	// Cancelar el contexto para notificar a todos los workers
	p.cancel()

	// Esperar a que todos los workers terminen
	p.wg.Wait()

	// Limpiar recursos
	close(p.jobQueue)
	close(p.workerPool)

	p.started = false
	log.Println("Worker pool stopped")
}

// Size devuelve el número de workers en el pool
func (p *Pool) Size() int {
	return p.maxWorkers
}

// Pending devuelve el número aproximado de trabajos pendientes
func (p *Pool) Pending() int {
	return len(p.jobQueue)
}
