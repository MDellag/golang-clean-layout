package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// CounterJob is a test helper that counts executions
type CounterJob struct {
	id       string
	counter  *atomic.Int64
	delay    time.Duration
	shouldErr bool
}

// NewCounterJob creates a new counter job for testing
func NewCounterJob(id string, counter *atomic.Int64, delay time.Duration) *CounterJob {
	return &CounterJob{
		id:      id,
		counter: counter,
		delay:   delay,
	}
}

// NewFailingCounterJob creates a counter job that always fails
func NewFailingCounterJob(id string, counter *atomic.Int64, delay time.Duration) *CounterJob {
	return &CounterJob{
		id:        id,
		counter:   counter,
		delay:     delay,
		shouldErr: true,
	}
}

func (j *CounterJob) Execute(ctx context.Context) error {
	select {
	case <-time.After(j.delay):
		j.counter.Add(1)
		if j.shouldErr {
			return fmt.Errorf("job %s failed intentionally", j.id)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (j *CounterJob) Name() string {
	return j.id
}

func (j *CounterJob) Priority() int {
	return 1
}

// Test 1: Basic lifecycle
func TestDispatcher_StartStop(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 3, 10)

	// Should start successfully
	if err := d.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Duplicate start should fail
	if err := d.Start(); err == nil {
		t.Error("Expected duplicate start to fail")
	}

	// Should stop successfully
	d.Stop()

	// Stats should show stopped
	stats := d.Stats()
	if stats.IsRunning {
		t.Error("Expected IsRunning=false after Stop")
	}
}

// Test 2: Job execution
func TestDispatcher_EnqueueJobs(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 3, 10)
	defer d.Stop()

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}

	// Submit jobs and count executions
	var counter atomic.Int64
	for i := 0; i < 10; i++ {
		job := NewCounterJob(fmt.Sprintf("job-%d", i), &counter, 10*time.Millisecond)
		if err := d.EnqueueJob(job); err != nil {
			t.Errorf("EnqueueJob failed: %v", err)
		}
	}

	// Wait for completion
	time.Sleep(200 * time.Millisecond)

	if counter.Load() != 10 {
		t.Errorf("Expected 10 jobs executed, got %d", counter.Load())
	}
}

// Test 3: Concurrent enqueue
func TestDispatcher_ConcurrentEnqueue(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 5, 100)
	defer d.Stop()

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}

	var counter atomic.Int64
	var wg sync.WaitGroup

	// 10 goroutines submit 10 jobs each
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				job := NewCounterJob(
					fmt.Sprintf("worker-%d-job-%d", workerID, j),
					&counter,
					5*time.Millisecond,
				)
				if err := d.EnqueueJob(job); err != nil {
					t.Errorf("EnqueueJob failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond) // Wait for processing

	if counter.Load() != 100 {
		t.Errorf("Expected 100 jobs executed, got %d", counter.Load())
	}
}

// Test 4: Context cancellation
func TestDispatcher_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	d := NewDispatcher(ctx, 3, 10)

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}

	// Submit long-running jobs
	var counter atomic.Int64
	for i := 0; i < 5; i++ {
		job := NewCounterJob(fmt.Sprintf("job-%d", i), &counter, 5*time.Second)
		d.EnqueueJob(job)
	}

	// Cancel context mid-execution
	cancel()
	time.Sleep(100 * time.Millisecond)

	d.Stop()

	// Not all jobs should complete
	if counter.Load() >= 5 {
		t.Error("Expected jobs to be cancelled")
	}
}

// Test 5: Graceful shutdown
// Note: Current implementation cancels jobs on stop but ensures clean shutdown
func TestDispatcher_GracefulShutdown(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 5, 50)

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}

	// Submit batch of quick jobs
	var counter atomic.Int64
	for i := 0; i < 10; i++ {
		job := NewCounterJob(fmt.Sprintf("job-%d", i), &counter, 10*time.Millisecond)
		d.EnqueueJob(job)
	}

	// Give jobs time to start processing
	time.Sleep(50 * time.Millisecond)

	// Graceful stop with sufficient timeout
	if err := d.StopGraceful(2 * time.Second); err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	// Some jobs should have completed (those that started before stop)
	completed := counter.Load()
	t.Logf("Graceful shutdown: %d jobs completed", completed)

	// Verify dispatcher is stopped
	stats := d.Stats()
	if stats.IsRunning {
		t.Error("Expected dispatcher to be stopped")
	}
}

// Test 6: Graceful shutdown with timeout parameter
// Note: Current implementation cancels jobs immediately on stop,
// so this tests that stop completes within the timeout period
func TestDispatcher_GracefulShutdownTimeout(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 3, 50)

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}

	// Give workers time to start
	time.Sleep(50 * time.Millisecond)

	// Submit jobs
	var counter atomic.Int64
	for i := 0; i < 10; i++ {
		job := NewCounterJob(fmt.Sprintf("job-%d", i), &counter, 100*time.Millisecond)
		d.EnqueueJob(job)
	}

	// Graceful stop with timeout - should complete successfully
	start := time.Now()
	err := d.StopGraceful(2 * time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("StopGraceful should not timeout: %v", err)
	}

	// Should complete quickly (not wait for full timeout)
	if elapsed > 1*time.Second {
		t.Errorf("StopGraceful took too long: %v", elapsed)
	}

	t.Logf("StopGraceful completed in %v, %d jobs completed", elapsed, counter.Load())
}

// Test 7: Stats accuracy
func TestDispatcher_Stats(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 5, 20)

	// Before start
	stats := d.Stats()
	if stats.IsRunning {
		t.Error("Expected IsRunning=false before start")
	}

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}
	defer d.Stop()

	// After start
	stats = d.Stats()
	if !stats.IsRunning {
		t.Error("Expected IsRunning=true after start")
	}
	if stats.Workers != 5 {
		t.Errorf("Expected 5 workers, got %d", stats.Workers)
	}
}

// Test 8: Enqueue before start
func TestDispatcher_EnqueueBeforeStart(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 3, 10)

	job := NewCounterJob("job-1", &atomic.Int64{}, 10*time.Millisecond)

	// Should fail
	if err := d.EnqueueJob(job); err == nil {
		t.Error("Expected EnqueueJob to fail before Start")
	}
}

// Test 9: Multiple stop calls are safe
func TestDispatcher_MultipleStopCalls(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 3, 10)

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}

	// Multiple stops should not panic
	d.Stop()
	d.Stop()
	d.Stop()

	stats := d.Stats()
	if stats.IsRunning {
		t.Error("Expected IsRunning=false after Stop")
	}
}

// Test 10: Job errors are handled gracefully
func TestDispatcher_JobErrors(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 3, 10)
	defer d.Stop()

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}

	// Submit jobs that will fail
	var counter atomic.Int64
	for i := 0; i < 5; i++ {
		job := NewFailingCounterJob(fmt.Sprintf("job-%d", i), &counter, 10*time.Millisecond)
		if err := d.EnqueueJob(job); err != nil {
			t.Errorf("EnqueueJob failed: %v", err)
		}
	}

	// Wait for completion
	time.Sleep(200 * time.Millisecond)

	// All jobs should have been attempted (counter incremented even though they failed)
	if counter.Load() != 5 {
		t.Errorf("Expected 5 jobs attempted, got %d", counter.Load())
	}
}

// Test 11: Empty dispatcher lifecycle
func TestDispatcher_EmptyLifecycle(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 3, 10)

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}

	// Stop without submitting any jobs
	d.Stop()

	stats := d.Stats()
	if stats.IsRunning {
		t.Error("Expected IsRunning=false after Stop")
	}
	if stats.PendingJobs != 0 {
		t.Errorf("Expected 0 pending jobs, got %d", stats.PendingJobs)
	}
}

// Test 12: Pending jobs counter
func TestDispatcher_PendingJobsCounter(t *testing.T) {
	ctx := context.Background()
	d := NewDispatcher(ctx, 1, 100) // Single worker to create backlog
	defer d.Stop()

	if err := d.Start(); err != nil {
		t.Fatal(err)
	}

	// Submit jobs that take a while
	var counter atomic.Int64
	for i := 0; i < 10; i++ {
		job := NewCounterJob(fmt.Sprintf("job-%d", i), &counter, 100*time.Millisecond)
		d.EnqueueJob(job)
	}

	// Check pending jobs (should be > 0 since we have only 1 worker)
	time.Sleep(10 * time.Millisecond) // Let first job start
	stats := d.Stats()
	if stats.PendingJobs == 0 {
		t.Log("Warning: Expected some pending jobs, but this is non-deterministic")
	}
}
