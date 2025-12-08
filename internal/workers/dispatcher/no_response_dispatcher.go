package dispatcher

import (
	"context"
	"sync"

	"clean-arq-layout/internal/workers/jobs"
)

// NoResponseDispatcher handles jobs that don't need to return responses
type NoResponseDispatcher[T any] struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	metrics    MetricsTracker
}

// NewNoResponseDispatcher creates a new dispatcher for jobs that don't return responses
func NewNoResponseDispatcher[T any](ctx context.Context, metrics MetricsTracker) *NoResponseDispatcher[T] {
	ctx, cancel := context.WithCancel(ctx)
	return &NoResponseDispatcher[T]{
		ctx:        ctx,
		cancelFunc: cancel,
		metrics:    metrics,
	}
}

// Dispatch dispatches multiple jobs without expecting responses
func (d *NoResponseDispatcher[T]) Dispatch(numWorkers int, jobList []jobs.Job[T]) <-chan JobResult[T] {
	// Create a dummy result channel just to signal completion
	resultChan := make(chan JobResult[T], 1)
	jobChan := make(chan jobs.Job[T], len(jobList))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go processJobsNoResponse(d.ctx, jobChan, d.metrics, &wg)
	}

	// Send jobs to the job channel
	go func() {
		for _, job := range jobList {
			d.metrics.IncrementTotal()
			select {
			case <-d.ctx.Done():
				close(jobChan)
				return
			case jobChan <- job:
			}
		}
		close(jobChan)
	}()

	// Signal completion when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}

// DispatchSingle dispatches a single job without expecting a response
func (d *NoResponseDispatcher[T]) DispatchSingle(numWorkers int, job jobs.Job[T]) <-chan JobResult[T] {
	return d.Dispatch(numWorkers, []jobs.Job[T]{job})
}

// Cancel cancels all running jobs
func (d *NoResponseDispatcher[T]) Cancel() {
	if d.cancelFunc != nil {
		d.cancelFunc()
	}
}
