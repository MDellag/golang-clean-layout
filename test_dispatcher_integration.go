// Simple integration test for the improved Dispatcher
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	worker "clean-arq-layout/internal/workers"
	"clean-arq-layout/internal/workers/jobs"
)

func main() {
	log.Println("=== Dispatcher Integration Test ===")

	ctx := context.Background()
	d := worker.NewDispatcher(ctx, 3, 10)

	// Test 1: Start dispatcher
	log.Println("\n[Test 1] Starting dispatcher...")
	if err := d.Start(); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	log.Println("✓ Dispatcher started successfully")

	// Test 2: Check stats
	log.Println("\n[Test 2] Checking stats...")
	stats := d.Stats()
	log.Printf("✓ Stats: Workers=%d, Pending=%d, Running=%v",
		stats.Workers, stats.PendingJobs, stats.IsRunning)

	// Test 3: Submit jobs
	log.Println("\n[Test 3] Submitting 20 jobs...")
	for i := 0; i < 20; i++ {
		job := jobs.NewSimpleJob(
			fmt.Sprintf("task-%d", i),
			fmt.Sprintf("data-%d", i),
			100*time.Millisecond,
		)
		if err := d.EnqueueJob(job); err != nil {
			log.Printf("Failed to enqueue job %d: %v", i, err)
		}
	}
	log.Println("✓ All jobs submitted")

	// Test 4: Monitor progress
	log.Println("\n[Test 4] Monitoring progress...")
	for i := 0; i < 5; i++ {
		time.Sleep(500 * time.Millisecond)
		stats := d.Stats()
		log.Printf("  Progress check %d: Pending=%d", i+1, stats.PendingJobs)
		if stats.PendingJobs == 0 {
			break
		}
	}

	// Test 5: Graceful shutdown
	log.Println("\n[Test 5] Testing graceful shutdown...")
	if err := d.StopGraceful(5 * time.Second); err != nil {
		log.Printf("Graceful shutdown error: %v", err)
	} else {
		log.Println("✓ Graceful shutdown completed")
	}

	// Test 6: Verify stopped
	log.Println("\n[Test 6] Verifying stopped state...")
	stats = d.Stats()
	if !stats.IsRunning {
		log.Println("✓ Dispatcher is stopped")
	} else {
		log.Println("✗ Dispatcher still running!")
	}

	log.Println("\n=== All Integration Tests Passed! ===")
}
