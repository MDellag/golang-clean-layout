package examples_test

import (
	"context"
	"fmt"
	"log"
	"time"

	worker "clean-arq-layout/internal/workers"
	"clean-arq-layout/internal/workers/jobs"
)

// Example 1: Basic fire-and-forget job execution
func ExampleDispatcher_basic() {
	ctx := context.Background()
	dispatcher := worker.NewDispatcher(ctx, 5, 100)

	if err := dispatcher.Start(); err != nil {
		log.Fatal(err)
	}
	defer dispatcher.Stop()

	// Submit simple jobs
	for i := 0; i < 10; i++ {
		job := jobs.NewSimpleJob(
			fmt.Sprintf("task-%d", i),
			fmt.Sprintf("data-%d", i),
			100*time.Millisecond,
		)

		if err := dispatcher.EnqueueJob(job); err != nil {
			log.Printf("Failed to enqueue: %v", err)
		}
	}

	// Jobs execute in background
	time.Sleep(2 * time.Second)

	fmt.Println("All jobs submitted")
	// Output: All jobs submitted
}

// Example 2: Monitoring dispatcher statistics
func ExampleDispatcher_monitoring() {
	ctx := context.Background()
	dispatcher := worker.NewDispatcher(ctx, 10, 200)

	if err := dispatcher.Start(); err != nil {
		log.Fatal(err)
	}
	defer dispatcher.Stop()

	// Submit jobs
	for i := 0; i < 50; i++ {
		job := jobs.NewSimpleJob(
			fmt.Sprintf("task-%d", i),
			"data",
			100*time.Millisecond,
		)
		dispatcher.EnqueueJob(job)
	}

	// Monitor progress
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			stats := dispatcher.Stats()
			fmt.Printf("Workers: %d, Pending: %d, Running: %v\n",
				stats.Workers, stats.PendingJobs, stats.IsRunning)

			if stats.PendingJobs == 0 {
				fmt.Println("All jobs completed")
				return
			}

		case <-timeout:
			fmt.Println("Monitoring timeout")
			return
		}
	}
}

// Example 3: Graceful vs immediate shutdown
func ExampleDispatcher_shutdown() {
	ctx := context.Background()

	// Option A: Graceful shutdown (wait for completion)
	fmt.Println("Testing graceful shutdown:")
	dispatcher1 := worker.NewDispatcher(ctx, 5, 100)
	if err := dispatcher1.Start(); err != nil {
		log.Fatal(err)
	}

	// Submit batch
	for i := 0; i < 20; i++ {
		job := jobs.NewSimpleJob(
			fmt.Sprintf("task-%d", i),
			"data",
			50*time.Millisecond,
		)
		dispatcher1.EnqueueJob(job)
	}

	if err := dispatcher1.StopGraceful(5 * time.Second); err != nil {
		fmt.Printf("Graceful shutdown timeout: %v\n", err)
	} else {
		fmt.Println("Graceful shutdown completed")
	}

	// Option B: Immediate shutdown (cancel remaining)
	fmt.Println("\nTesting immediate shutdown:")
	dispatcher2 := worker.NewDispatcher(ctx, 5, 100)
	if err := dispatcher2.Start(); err != nil {
		log.Fatal(err)
	}

	// Submit batch
	for i := 0; i < 20; i++ {
		job := jobs.NewSimpleJob(
			fmt.Sprintf("task-%d", i),
			"data",
			500*time.Millisecond,
		)
		dispatcher2.EnqueueJob(job)
	}

	dispatcher2.Stop()
	fmt.Println("Immediate shutdown completed")

	// Output:
	// Testing graceful shutdown:
	// Graceful shutdown completed
	//
	// Testing immediate shutdown:
	// Immediate shutdown completed
}

// Example 4: Error handling in concurrent environment
func ExampleDispatcher_concurrent() {
	ctx := context.Background()
	dispatcher := worker.NewDispatcher(ctx, 3, 50)

	if err := dispatcher.Start(); err != nil {
		log.Fatal(err)
	}
	defer dispatcher.StopGraceful(5 * time.Second)

	// Simulate concurrent job submission from multiple goroutines
	var done = make(chan bool, 3)

	for workerID := 0; workerID < 3; workerID++ {
		go func(id int) {
			for i := 0; i < 5; i++ {
				job := jobs.NewSimpleJob(
					fmt.Sprintf("worker-%d-task-%d", id, i),
					fmt.Sprintf("data-%d", i),
					50*time.Millisecond,
				)

				if err := dispatcher.EnqueueJob(job); err != nil {
					log.Printf("Worker %d failed to enqueue job %d: %v", id, i, err)
				}
			}
			done <- true
		}(workerID)
	}

	// Wait for all workers to submit their jobs
	for i := 0; i < 3; i++ {
		<-done
	}

	fmt.Println("All concurrent submissions completed")
	// Output: All concurrent submissions completed
}

// Example 5: Working with dispatcher lifecycle
func ExampleDispatcher_lifecycle() {
	ctx := context.Background()
	dispatcher := worker.NewDispatcher(ctx, 5, 50)

	// Check initial state
	stats := dispatcher.Stats()
	fmt.Printf("Before Start - Running: %v, Workers: %d\n", stats.IsRunning, stats.Workers)

	// Start dispatcher
	if err := dispatcher.Start(); err != nil {
		log.Fatal(err)
	}

	// Check running state
	stats = dispatcher.Stats()
	fmt.Printf("After Start - Running: %v, Workers: %d\n", stats.IsRunning, stats.Workers)

	// Submit a job
	job := jobs.NewSimpleJob("task-1", "data", 50*time.Millisecond)
	if err := dispatcher.EnqueueJob(job); err != nil {
		log.Printf("Failed to enqueue: %v", err)
	}

	// Wait for job to complete
	time.Sleep(200 * time.Millisecond)

	// Stop dispatcher
	dispatcher.Stop()

	// Check stopped state
	stats = dispatcher.Stats()
	fmt.Printf("After Stop - Running: %v, Workers: %d\n", stats.IsRunning, stats.Workers)

	// Output:
	// Before Start - Running: false, Workers: 5
	// After Start - Running: true, Workers: 5
	// After Stop - Running: false, Workers: 5
}
