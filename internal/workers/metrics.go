package workers

import (
	"sync"
	"time"
)

// Metrics tracks worker performance metrics
type Metrics struct {
	mu sync.RWMutex

	TotalJobs      int64
	CompletedJobs  int64
	FailedJobs     int64
	ActiveJobs     int64
	TotalDuration  time.Duration
	AverageDuration time.Duration
}

// IncrementTotal increments the total jobs counter
func (m *Metrics) IncrementTotal() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalJobs++
}

// IncrementActive increments the active jobs counter
func (m *Metrics) IncrementActive() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveJobs++
}

// DecrementActive decrements the active jobs counter
func (m *Metrics) DecrementActive() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveJobs--
}

// IncrementCompleted increments the completed jobs counter and updates duration metrics
func (m *Metrics) IncrementCompleted(duration int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CompletedJobs++
	m.TotalDuration += time.Duration(duration)
	if m.CompletedJobs > 0 {
		m.AverageDuration = m.TotalDuration / time.Duration(m.CompletedJobs)
	}
}

// IncrementFailed increments the failed jobs counter
func (m *Metrics) IncrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailedJobs++
}

// GetMetrics returns a snapshot of the current metrics
func (m *Metrics) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Metrics{
		TotalJobs:       m.TotalJobs,
		CompletedJobs:   m.CompletedJobs,
		FailedJobs:      m.FailedJobs,
		ActiveJobs:      m.ActiveJobs,
		TotalDuration:   m.TotalDuration,
		AverageDuration: m.AverageDuration,
	}
}
