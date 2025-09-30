package jobs

import (
	"context"
	"fmt"
	"log"
	"time"
)

// EmailJob represents an email sending job with retry logic
type EmailJob struct {
	ID        string
	Recipient string
	Subject   string
	Body      string
	MaxRetries int
	currentRetry int
}

// NewEmailJob creates a new email job
func NewEmailJob(id, recipient, subject, body string) *EmailJob {
	return &EmailJob{
		ID:         id,
		Recipient:  recipient,
		Subject:    subject,
		Body:       body,
		MaxRetries: 3,
	}
}

// Execute implements Job interface with retry logic
func (j *EmailJob) Execute(ctx context.Context) error {
	log.Printf("Starting email job %s to %s", j.ID, j.Recipient)

	for j.currentRetry <= j.MaxRetries {
		select {
		case <-ctx.Done():
			log.Printf("Email job %s was canceled: %v", j.ID, ctx.Err())
			return ctx.Err()
		default:
		}

		// Simulate email sending (with potential failure)
		if err := j.sendEmail(ctx); err != nil {
			j.currentRetry++
			if j.currentRetry <= j.MaxRetries {
				log.Printf("Email job %s failed (attempt %d/%d): %v, retrying...", 
					j.ID, j.currentRetry, j.MaxRetries, err)
				
				// Wait before retry with exponential backoff
				backoff := time.Duration(j.currentRetry*j.currentRetry) * time.Second
				timer := time.NewTimer(backoff)
				
				select {
				case <-timer.C:
					continue
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
			} else {
				log.Printf("Email job %s failed permanently after %d attempts", j.ID, j.MaxRetries)
				return fmt.Errorf("email job failed after %d attempts: %w", j.MaxRetries, err)
			}
		} else {
			log.Printf("Email job %s completed successfully", j.ID)
			return nil
		}
	}

	return fmt.Errorf("email job %s exhausted all retry attempts", j.ID)
}

// sendEmail simulates sending an email
func (j *EmailJob) sendEmail(ctx context.Context) error {
	// Simulate processing time
	select {
	case <-time.After(500 * time.Millisecond):
		// Simulate 70% success rate
		if time.Now().UnixNano()%10 < 7 {
			return nil
		}
		return fmt.Errorf("temporary email service error")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Name returns the job name for identification
func (j *EmailJob) Name() string {
	return fmt.Sprintf("email-job-%s", j.ID)
}

// Priority returns priority (higher number = higher priority)
func (j *EmailJob) Priority() int {
	return 2 // Higher priority than simple jobs
}