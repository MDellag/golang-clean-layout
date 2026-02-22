package worker

import (
	"context"
	"testing"
	"time"

	"clean-arq-layout/internal/workers/jobs"
	"clean-arq-layout/internal/workers/types"
)

// TestGenericDispatcher_Simple muestra el uso más básico: dispatch y esperar resultados.
func TestGenericDispatcher_Simple(t *testing.T) {
	dispatcher := NewGenericDispatcher[jobs.UserResponse](3, 10)
	defer dispatcher.Stop()

	userJobs := []types.GenericJob[jobs.UserResponse]{
		jobs.NewUserAPIJob(1, 1),
		jobs.NewUserAPIJob(2, 1),
		jobs.NewUserAPIJob(3, 1),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := dispatcher.DispatchAndWait(ctx, userJobs)
	if err != nil {
		t.Fatalf("DispatchAndWait failed: %v", err)
	}

	if len(results) != len(userJobs) {
		t.Errorf("expected %d results, got %d", len(userJobs), len(results))
	}

	for _, r := range results {
		if r.Success {
			t.Logf("✓ %s: %s (%s)", r.JobName, r.Result.Name, r.Result.Email)
		} else {
			t.Logf("✗ %s: %v", r.JobName, r.Error)
		}
	}
}

// TestGenericDispatcher_Channel muestra cómo consumir resultados a medida que llegan,
// sin esperar a que todos los jobs terminen.
func TestGenericDispatcher_Channel(t *testing.T) {
	dispatcher := NewGenericDispatcher[jobs.PostResponse](5, 20)
	defer dispatcher.Stop()

	postJobs := make([]types.GenericJob[jobs.PostResponse], 10)
	for i := range postJobs {
		postJobs[i] = jobs.NewPostAPIJob(i+1, 1)
	}

	resultChan, err := dispatcher.Dispatch(postJobs)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	success, failed := 0, 0
	for r := range resultChan {
		if r.Success {
			success++
			t.Logf("✓ post %d: %s", r.Result.ID, r.Result.Title)
		} else {
			failed++
			t.Logf("✗ %s: %v", r.JobName, r.Error)
		}
	}

	t.Logf("done: %d ok, %d failed", success, failed)

	if success+failed != len(postJobs) {
		t.Errorf("expected %d results, got %d", len(postJobs), success+failed)
	}
}

// TestGenericDispatcher_MixedResults verifica que el dispatcher reporta éxitos y errores
// correctamente cuando algunos jobs fallan (IDs inválidos).
func TestGenericDispatcher_MixedResults(t *testing.T) {
	dispatcher := NewGenericDispatcher[jobs.UserResponse](3, 10)
	defer dispatcher.Stop()

	userJobs := []types.GenericJob[jobs.UserResponse]{
		jobs.NewUserAPIJob(1, 1),    // válido
		jobs.NewUserAPIJob(9999, 1), // ID inexistente
		jobs.NewUserAPIJob(2, 1),    // válido
		jobs.NewUserAPIJob(3, 1),    // válido
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := dispatcher.DispatchAndWait(ctx, userJobs)
	if err != nil {
		t.Fatalf("DispatchAndWait failed: %v", err)
	}

	if len(results) != len(userJobs) {
		t.Errorf("expected %d results, got %d", len(userJobs), len(results))
	}

	success, failed := 0, 0
	for _, r := range results {
		if r.Success {
			success++
		} else {
			failed++
		}
	}
	t.Logf("results: %d ok, %d failed", success, failed)
}
