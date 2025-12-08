package jobs

import (
	"context"
	"fmt"
)

// ExampleJob is a simple example implementation of the Job interface
type ExampleJob struct {
	ID   int
	Data string
}

// Execute implements the Job interface
func (e *ExampleJob) Execute(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// Simulate some work
		result := fmt.Sprintf("Job %d processed: %s", e.ID, e.Data)
		return result, nil
	}
}

// APIJob is an example of a job that makes an API request and returns a response
type APIJob[T any] struct {
	URL string
	// Add other fields as needed for API requests
}

// Execute implements the Job interface for API calls
func (a *APIJob[T]) Execute(ctx context.Context) (T, error) {
	var result T

	select {
	case <-ctx.Done():
		return result, ctx.Err()
	default:
		// Here you would make the actual API request
		// For now, this is just a placeholder
		return result, nil
	}
}
