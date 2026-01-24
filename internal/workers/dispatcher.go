// Package worker provides a simple fire-and-forget job dispatch system.
//
// The Dispatcher coordinates job execution across a pool of workers without
// collecting results. Use this when you need async execution but don't care
// about return values (e.g., sending emails, logging, cache invalidation).
//
// For jobs that return typed results, use GenericDispatcher[T] instead.
//
// Example usage:
//
//	ctx := context.Background()
//	dispatcher := worker.NewDispatcher(ctx, 5, 100) // 5 workers, queue of 100
//
//	if err := dispatcher.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	defer dispatcher.Stop()
//
//	// Submit jobs
//	job := jobs.NewSimpleJob("task-1", "data", time.Second)
//	dispatcher.EnqueueJob(job)
//
// Thread Safety:
//   - EnqueueJob() is safe to call concurrently
//   - Start() and Stop() must be called from a single goroutine
package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"clean-arq-layout/internal/workers/types"
)

// Dispatcher coordinates fire-and-forget job execution across a worker pool.
//
// Jobs are submitted via EnqueueJob() and executed asynchronously without
// result collection. The dispatcher maintains a background monitoring goroutine
// that periodically logs worker statistics.
//
// Lifecycle:
//  1. Create: NewDispatcher(ctx, maxWorkers, queueSize)
//  2. Start: dispatcher.Start() - initializes workers
//  3. Use: dispatcher.EnqueueJob(job) - submit jobs
//  4. Stop: dispatcher.Stop() or dispatcher.StopGraceful(timeout)
//
// The zero value is NOT usable - always use NewDispatcher.
type Dispatcher struct {
	workerPool *Pool
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.Mutex
	started    bool
}

// DispatcherStats contains current dispatcher statistics.
type DispatcherStats struct {
	Workers     int  // Number of workers in the pool
	PendingJobs int  // Number of jobs waiting in queue
	IsRunning   bool // Whether dispatcher is currently running
}

// NewDispatcher creates a new fire-and-forget job dispatcher.
//
// Parameters:
//   - ctx: Parent context for lifecycle management
//   - maxWorkers: Number of concurrent workers to spawn
//   - queueSize: Maximum number of jobs that can be queued
//
// The dispatcher is created in a stopped state. Call Start() to begin processing.
//
// Example:
//
//	dispatcher := NewDispatcher(context.Background(), 10, 500)
func NewDispatcher(ctx context.Context, maxWorkers int, queueSize int) *Dispatcher {
	dispatcherCtx, cancel := context.WithCancel(ctx)

	return &Dispatcher{
		workerPool: NewWorkerPool(maxWorkers, queueSize),
		ctx:        dispatcherCtx,
		cancel:     cancel,
	}
}

// Start initializes the worker pool and begins job processing.
//
// This method MUST be called before EnqueueJob(). Calling Start() multiple
// times returns an error.
//
// Returns:
//   - error: If dispatcher is already started or pool fails to start
//
// Thread Safety: NOT safe for concurrent Start() calls.
func (d *Dispatcher) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.started {
		return fmt.Errorf("dispatcher already started")
	}

	// Start the worker pool
	if err := d.workerPool.Start(); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	// Start monitoring goroutine
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.monitor()
	}()

	d.started = true
	log.Println("Dispatcher started successfully")
	return nil
}

// EnqueueJob submits a job for async execution without waiting for results.
//
// The job will be executed by an available worker. If the queue is full,
// this method blocks until space becomes available or context is cancelled.
//
// Parameters:
//   - job: The job to execute (must implement types.Job interface)
//
// Returns:
//   - error: If dispatcher not started, queue full, or context cancelled
//
// Thread Safety: Safe for concurrent calls from multiple goroutines.
//
// Example:
//
//	job := jobs.NewSimpleJob("id", "data", 100*time.Millisecond)
//	if err := dispatcher.EnqueueJob(job); err != nil {
//	    log.Printf("Failed to enqueue: %v", err)
//	}
func (d *Dispatcher) EnqueueJob(job types.Job) error {
	if !d.started {
		return fmt.Errorf("dispatcher not started")
	}

	return d.workerPool.Submit(job)
}

// Stop immediately stops the dispatcher and cancels all running jobs.
//
// Jobs currently in the queue are NOT processed. For graceful shutdown
// that completes queued jobs, use StopGraceful() instead.
//
// This method blocks until all worker goroutines exit.
//
// Thread Safety: NOT safe for concurrent Stop() calls.
func (d *Dispatcher) Stop() {
	d.StopGraceful(0)
}

// StopGraceful stops the dispatcher after processing remaining queued jobs.
//
// Parameters:
//   - timeout: Maximum time to wait for queue drainage (0 = immediate stop)
//
// Returns:
//   - error: If timeout expires before queue is drained
//
// This method blocks until either:
//  1. All queued jobs complete successfully
//  2. The timeout expires (remaining jobs are cancelled)
//
// Example:
//
//	if err := dispatcher.StopGraceful(30 * time.Second); err != nil {
//	    log.Printf("Some jobs were cancelled: %v", err)
//	}
func (d *Dispatcher) StopGraceful(timeout time.Duration) error {
	d.mu.Lock()
	if !d.started {
		d.mu.Unlock()
		return nil
	}
	d.mu.Unlock()

	log.Println("Gracefully stopping dispatcher...")

	// If timeout is 0, force immediate stop
	if timeout == 0 {
		d.cancel()
		d.workerPool.Stop()
		d.wg.Wait()
		d.mu.Lock()
		d.started = false
		d.mu.Unlock()
		log.Println("Dispatcher stopped immediately")
		return nil
	}

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		d.workerPool.Stop()
		d.cancel()
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		d.mu.Lock()
		d.started = false
		d.mu.Unlock()
		log.Println("Dispatcher stopped gracefully")
		return nil
	case <-time.After(timeout):
		d.cancel() // Force stop remaining
		d.mu.Lock()
		d.started = false
		d.mu.Unlock()
		return fmt.Errorf("graceful shutdown timeout after %v", timeout)
	}
}

// monitor is a background routine that periodically logs worker statistics.
// It could be extended to implement dynamic scaling based on load.
func (d *Dispatcher) monitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pending := d.workerPool.Pending()
			log.Printf("Worker stats - Workers: %d, Pending jobs: %d",
				d.workerPool.Size(), pending)

			// Future enhancement: Could implement logic to scale worker count
			// based on current load

		case <-d.ctx.Done():
			log.Println("Dispatcher monitor shutting down")
			return
		}
	}
}

// Stats returns current dispatcher statistics.
//
// Thread Safety: Safe for concurrent calls.
func (d *Dispatcher) Stats() DispatcherStats {
	d.mu.Lock()
	defer d.mu.Unlock()

	return DispatcherStats{
		Workers:     d.workerPool.Size(),
		PendingJobs: d.workerPool.Pending(),
		IsRunning:   d.started,
	}
}
