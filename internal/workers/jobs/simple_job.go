package jobs

import (
	"context"
	"fmt"
	"log"
	"time"
)

// TIP SimpleJob is a basic job for processing without dependencies
type SimpleJob struct {
	ID    string
	Data  string
	Delay time.Duration
}

// NewSimpleJob creates a simple job (it's for example purpose)
func NewSimpleJob(id string, data string, delay time.Duration) *SimpleJob {
	return &SimpleJob{
		ID:    id,
		Data:  data,
		Delay: delay,
	}
}

// Execute Implements Job interface and it customs logic
func (j *SimpleJob) Execute(ctx context.Context) error {
	log.Printf("Starting job %s with data: %s", j.ID, j.Data)

	timer := time.NewTimer(j.Delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		result := fmt.Sprintf("Processed: %s", j.Data)
		log.Printf("Completed job %s with result: %s", j.ID, result)
		return nil

	case <-ctx.Done():
		log.Printf("Job %s was canceled: %v", j.ID, ctx.Err())
		return ctx.Err()
	}
}

// Name returns the Job name for identification
func (j *SimpleJob) Name() string {
	return fmt.Sprintf("simple-job-%s", j.ID)
}

// Priority implements a simple priority (always normal = 1)
func (j *SimpleJob) Priority() int {
	return 1
}
