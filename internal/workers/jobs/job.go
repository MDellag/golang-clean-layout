package jobs

import "context"

// Job is the interface that all jobs must implement
type Job[T any] interface {
	// Execute runs the job and returns the result of type T
	Execute(ctx context.Context) (T, error)
}
