package worker

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"

	"clean-arq-layout/internal/workers/jobs"
	"clean-arq-layout/internal/workers/types"
)

func TestGenericDispatcher_WithUserAPIJobs(t *testing.T) {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	// Crear un dispatcher con 3 workers
	dispatcher := NewGenericDispatcher[jobs.UserResponse](3, 10)
	defer dispatcher.Stop()

	// Crear slice de jobs para obtener información de varios usuarios
	userJobs := []types.GenericJob[jobs.UserResponse]{
		jobs.NewUserAPIJob(1, 1),
		jobs.NewUserAPIJob(2, 1),
		jobs.NewUserAPIJob(3, 1),
		jobs.NewUserAPIJob(4, 1),
		jobs.NewUserAPIJob(5, 1),
	}

	// Hacer dispatch de los jobs
	responseChan, err := dispatcher.Dispatch(userJobs)
	if err != nil {
		t.Fatalf("Failed to dispatch jobs: %v", err)
	}

	// Escuchar las respuestas
	results := make([]types.GenericJobResult[jobs.UserResponse], 0)

	// Timeout para evitar bloqueos - si llega aquí, hay errores...
	timeout := time.After(30 * time.Second)

	for {
		select {
		case result, ok := <-responseChan:
			if !ok {
				// Canal cerrado, todos los jobs completados
				t.Logf("Received %d results", len(results))

				// Verificar que recibimos todos los resultados
				if len(results) != len(userJobs) {
					t.Errorf("Expected %d results, got %d", len(userJobs), len(results))
				}

				// Verificar que todos los jobs fueron exitosos
				for _, res := range results {
					if !res.Success {
						t.Errorf("Job %s failed: %v", res.JobName, res.Error)
					} else {
						t.Logf("Job %s succeeded: User ID=%d, Name=%s, Email=%s",
							res.JobName, res.Result.ID, res.Result.Name, res.Result.Email)
					}
				}
				return
			} else {
				results = append(results, result)
				t.Logf("Received result from job: %s (Success: %v)", result.JobName, result.Success)
			}

		case <-timeout:
			t.Fatal("Test timeout - jobs took too long to complete")
		}
	}
}

func TestGenericDispatcher_WithPostAPIJobs(t *testing.T) {
	// Crear un dispatcher con 5 workers
	dispatcher := NewGenericDispatcher[jobs.PostResponse](5, 20)
	defer dispatcher.Stop()

	// Crear slice de jobs para obtener varios posts
	postJobs := make([]types.GenericJob[jobs.PostResponse], 0, 10)
	for i := 1; i <= 10; i++ {
		postJobs = append(postJobs, jobs.NewPostAPIJob(i, 1))
	}

	// Hacer dispatch de los jobs
	responseChan, err := dispatcher.Dispatch(postJobs)
	if err != nil {
		t.Fatalf("Failed to dispatch jobs: %v", err)
	}

	// Escuchar las respuestas
	successCount := 0
	errorCount := 0

	for result := range responseChan {
		if result.Success {
			successCount++
			t.Logf("✓ Post %d: %s", result.Result.ID, result.Result.Title)
		} else {
			errorCount++
			t.Logf("✗ Job %s failed: %v", result.JobName, result.Error)
		}
	}

	t.Logf("Completed: %d successful, %d errors", successCount, errorCount)

	if successCount+errorCount != len(postJobs) {
		t.Errorf("Expected %d total results, got %d", len(postJobs), successCount+errorCount)
	}
}

func TestGenericDispatcher_DispatchAndWait(t *testing.T) {
	// Crear un dispatcher
	dispatcher := NewGenericDispatcher[jobs.UserResponse](2, 5)
	defer dispatcher.Stop()

	// Crear jobs
	userJobs := []types.GenericJob[jobs.UserResponse]{
		jobs.NewUserAPIJob(1, 1),
		jobs.NewUserAPIJob(2, 1),
		jobs.NewUserAPIJob(3, 1),
	}

	// Usar DispatchAndWait para esperar todos los resultados
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := dispatcher.DispatchAndWait(ctx, userJobs)
	if err != nil {
		t.Fatalf("DispatchAndWait failed: %v", err)
	}

	t.Logf("Received %d results", len(results))

	// Verificar resultados
	for _, result := range results {
		if result.Success {
			t.Logf("✓ User: %s (%s)", result.Result.Name, result.Result.Email)
		} else {
			t.Logf("✗ Job %s failed: %v", result.JobName, result.Error)
		}
	}

	if len(results) != len(userJobs) {
		t.Errorf("Expected %d results, got %d", len(userJobs), len(results))
	}
}

func TestGenericDispatcher_MixedResults(t *testing.T) {
	// Test con jobs que pueden fallar (usando IDs inválidos)
	dispatcher := NewGenericDispatcher[jobs.UserResponse](3, 10)
	defer dispatcher.Stop()

	// Mezclar jobs válidos e inválidos
	userJobs := []types.GenericJob[jobs.UserResponse]{
		jobs.NewUserAPIJob(1, 1),    // válido
		jobs.NewUserAPIJob(9999, 1), // podría fallar
		jobs.NewUserAPIJob(2, 1),    // válido
		jobs.NewUserAPIJob(3, 1),    // válido
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := dispatcher.DispatchAndWait(ctx, userJobs)
	if err != nil {
		t.Fatalf("DispatchAndWait failed: %v", err)
	}

	successCount := 0
	errorCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			errorCount++
		}
	}

	t.Logf("Results: %d successful, %d errors", successCount, errorCount)

	if successCount+errorCount != len(userJobs) {
		t.Errorf("Expected %d total results, got %d", len(userJobs), successCount+errorCount)
	}
}
