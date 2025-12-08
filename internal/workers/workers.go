package workers

import (
	"context"

	"clean-arq-layout/internal/workers/dispatcher"
	"clean-arq-layout/internal/workers/jobs"
)

// Dispatcher is the interface for dispatching jobs to workers
type Dispatcher[T any] interface {
	// Dispatch dispatches a slice of jobs and returns a channel to receive results
	Dispatch(numWorkers int, jobs []jobs.Job[T]) <-chan dispatcher.JobResult[T]

	// DispatchSingle dispatches a single job and returns a channel to receive the result
	DispatchSingle(numWorkers int, job jobs.Job[T]) <-chan dispatcher.JobResult[T]

	// Cancel cancels all running jobs
	Cancel()
}

// dispatcherAny is a type-erased version of Dispatcher for internal storage
type dispatcherAny interface {
	Cancel()
}

// Workers manages a pool of workers to process jobs concurrently
// Workers is NOT generic - type safety is provided by the dispatcher passed in the constructor
type Workers struct {
	numWorkers int
	dispatcher dispatcherAny
	metrics    *Metrics
}

// NewWorkers creates a new Workers instance with the specified number of workers
// Metrics are initialized automatically and can be accessed via GetMetrics()
func NewWorkers[T any](numWorkers int, ctx context.Context, dispatcherFactory func(context.Context, dispatcher.MetricsTracker) Dispatcher[T]) *Workers {
	metrics := &Metrics{}
	disp := dispatcherFactory(ctx, metrics)

	return &Workers{
		numWorkers: numWorkers,
		dispatcher: disp,
		metrics:    metrics,
	}
}

// Dispatch dispatches multiple jobs using the configured dispatcher
func Dispatch[T any](w *Workers, jobs []jobs.Job[T]) <-chan dispatcher.JobResult[T] {
	disp := w.dispatcher.(Dispatcher[T])
	return disp.Dispatch(w.numWorkers, jobs)
}

// DispatchSingle dispatches a single job using the configured dispatcher
func DispatchSingle[T any](w *Workers, job jobs.Job[T]) <-chan dispatcher.JobResult[T] {
	disp := w.dispatcher.(Dispatcher[T])
	return disp.DispatchSingle(w.numWorkers, job)
}

// GetMetrics returns the current metrics for the workers
func (w *Workers) GetMetrics() Metrics {
	return w.metrics.GetMetrics()
}

// Cancel cancels all running jobs
func (w *Workers) Cancel() {
	w.dispatcher.Cancel()
}
