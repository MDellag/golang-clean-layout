package types

import (
	"context"
	"time"
)

// ResponseAggregator recolecta y procesa resultados de jobs
type ResponseAggregator struct {
	results   chan JobResult
	processed []JobResult
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewResponseAggregator crea un nuevo agregador de respuestas
func NewResponseAggregator(bufferSize int) *ResponseAggregator {
	ctx, cancel := context.WithCancel(context.Background())
	return &ResponseAggregator{
		results:   make(chan JobResult, bufferSize),
		processed: make([]JobResult, 0),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start inicia el agregador
func (ra *ResponseAggregator) Start() {
	go func() {
		for {
			select {
			case result := <-ra.results:
				ra.processed = append(ra.processed, result)
			case <-ra.ctx.Done():
				return
			}
		}
	}()
}

// Stop detiene el agregador
func (ra *ResponseAggregator) Stop() {
	ra.cancel()
	close(ra.results)
}

// GetResults devuelve todos los resultados procesados
func (ra *ResponseAggregator) GetResults() []JobResult {
	return ra.processed
}

// GetResultsChannel devuelve el canal para enviar resultados
func (ra *ResponseAggregator) GetResultsChannel() chan<- JobResult {
	return ra.results
}

// WaitForResults espera hasta recibir el número esperado de resultados o timeout
func (ra *ResponseAggregator) WaitForResults(expectedCount int, timeout time.Duration) []JobResult {
	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return ra.processed
		case <-ticker.C:
			if len(ra.processed) >= expectedCount {
				return ra.processed
			}
		case <-ra.ctx.Done():
			return ra.processed
		}
	}
}