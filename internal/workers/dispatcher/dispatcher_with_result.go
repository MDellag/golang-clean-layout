package dispatcher

import (
	"context"
	"sync"

	"clean-arq-layout/internal/workers/jobs"
)

// DispatcherWithResult handles jobs that return responses through a channel
type DispatcherWithResult[T any] struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	metrics    MetricsTracker
}

// NewDispatcherWithResult creates a new dispatcher for jobs that return responses
func NewDispatcherWithResult[T any](ctx context.Context, metrics MetricsTracker) *DispatcherWithResult[T] {
	ctx, cancel := context.WithCancel(ctx)
	return &DispatcherWithResult[T]{
		ctx:        ctx,
		cancelFunc: cancel,
		metrics:    metrics,
	}
}

// Dispatch dispatches multiple jobs and returns a channel to receive results
func (d *DispatcherWithResult[T]) Dispatch(numWorkers int, jobList []jobs.Job[T]) <-chan JobResult[T] {
	resultChan := make(chan JobResult[T], len(jobList))
	jobChan := make(chan jobs.Job[T], len(jobList))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go processJobs(d.ctx, jobChan, resultChan, d.metrics, &wg)
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

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}

// DispatchSingle dispatches a single job and returns a channel to receive the result
func (d *DispatcherWithResult[T]) DispatchSingle(numWorkers int, job jobs.Job[T]) <-chan JobResult[T] {
	return d.Dispatch(numWorkers, []jobs.Job[T]{job})
}

// Cancel cancels all running jobs
func (d *DispatcherWithResult[T]) Cancel() {
	if d.cancelFunc != nil {
		d.cancelFunc()
	}
}
