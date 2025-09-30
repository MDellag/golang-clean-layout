package worker

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"clean-arq-layout/internal/workers/types"
)

// Worker representa un trabajador individual que procesa jobs
type Worker struct {
	ID         int
	jobChannel chan Job
	workerPool chan chan Job
	ctx        context.Context
	metrics    *WorkerMetrics
}

// WorkerMetrics almacena métricas del worker
type WorkerMetrics struct {
	JobsProcessed int64
	Errors        int64
	LastJobTime   time.Duration
}

// NewWorker crea un nuevo worker
func NewWorker(id int, workerPool chan chan Job, ctx context.Context) *Worker {
	return &Worker{
		ID:         id,
		jobChannel: make(chan Job),
		workerPool: workerPool,
		ctx:        ctx,
		metrics:    &WorkerMetrics{},
	}
}

// run ejecuta el worker (usado con waitgroup.Go)
func (w *Worker) run() {
	log.Printf("Worker %d starting", w.ID)

	for {
		// Registrar este worker como disponible para trabajos
		select {
		case w.workerPool <- w.jobChannel:
			// El worker está ahora en el pool, esperando trabajo
		case <-w.ctx.Done():
			log.Printf("Worker %d shutting down", w.ID)
			return
		}

		// Esperar a recibir un trabajo o señal de cierre
		select {
		case job := <-w.jobChannel:
			w.processJob(job)
		case <-w.ctx.Done():
			log.Printf("Worker %d shutting down", w.ID)
			return
		}
	}
}

// processJob procesa un trabajo individual
func (w *Worker) processJob(job Job) {
	log.Printf("Worker %d processing job: %s", w.ID, job.Name())

	startTime := time.Now()

	// Crear un contexto derivado con timeout para el job
	jobCtx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	defer cancel()

	// Ejecutar el trabajo
	err := job.Execute(jobCtx)

	duration := time.Since(startTime)
	w.metrics.LastJobTime = duration

	// Si el job implementa JobWithResponse, enviar el resultado
	if jobWithResponse, ok := job.(types.JobWithResponse); ok {
		result := types.JobResult{
			JobID:     jobWithResponse.ID(),
			JobName:   job.Name(),
			Success:   err == nil,
			Error:     err,
			Duration:  duration,
			Timestamp: time.Now(),
		}

		// Intentar enviar resultado al canal
		select {
		case jobWithResponse.ResponseChannel() <- result:
			// Resultado enviado exitosamente
		default:
			// Canal lleno o cerrado, log de advertencia
			log.Printf("Worker %d: failed to send result for job %s", w.ID, job.Name())
		}
	}

	if err != nil {
		log.Printf("Worker %d error processing job %s: %v", w.ID, job.Name(), err)
		atomic.AddInt64(&w.metrics.Errors, 1)
	} else {
		log.Printf("Worker %d completed job %s in %v", w.ID, job.Name(), duration)
		atomic.AddInt64(&w.metrics.JobsProcessed, 1)
	}
}

// Metrics devuelve las métricas del worker
func (w *Worker) Metrics() WorkerMetrics {
	return *w.metrics
}
