package dispatcher

import (
	"context"
	"sync"

	"clean-arq-layout/internal/workers/jobs"
)

// JobResult represents the result of a job execution
type JobResult[T any] struct {
	Result T
	Error  error
}

// MetricsTracker defines methods for tracking job metrics
type MetricsTracker interface {
	IncrementTotal()
	IncrementActive()
	DecrementActive()
	IncrementCompleted(duration int64)
	IncrementFailed()
}

// executeJob executes a single job and sends the result to the result channel
func executeJob[T any](ctx context.Context, job jobs.Job[T], resultChan chan<- JobResult[T], metrics MetricsTracker) {
	metrics.IncrementActive()
	defer metrics.DecrementActive()

	result, err := job.Execute(ctx)

	if err != nil {
		metrics.IncrementFailed()
		resultChan <- JobResult[T]{Error: err}
	} else {
		metrics.IncrementCompleted(0)
		resultChan <- JobResult[T]{Result: result}
	}
}

// executeJobNoResponse executes a job without sending the result to a channel
func executeJobNoResponse[T any](ctx context.Context, job jobs.Job[T], metrics MetricsTracker) {
	metrics.IncrementActive()
	defer metrics.DecrementActive()

	_, err := job.Execute(ctx)

	if err != nil {
		metrics.IncrementFailed()
	} else {
		metrics.IncrementCompleted(0)
	}
}

// processJobs is the worker function that processes jobs from the job channel
func processJobs[T any](ctx context.Context, jobChan <-chan jobs.Job[T], resultChan chan<- JobResult[T], metrics MetricsTracker, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobChan:
			if !ok {
				return
			}
			executeJob(ctx, job, resultChan, metrics)
		}
	}
}

// processJobsNoResponse processes jobs without sending results to a channel
func processJobsNoResponse[T any](ctx context.Context, jobChan <-chan jobs.Job[T], metrics MetricsTracker, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobChan:
			if !ok {
				return
			}
			executeJobNoResponse(ctx, job, metrics)
		}
	}
}
