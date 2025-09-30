package clients

import (
	"context"
	"fmt"
	"time"
)

// MockPriceServiceClient es un mock del cliente para testing
type MockPriceServiceClient struct {
	shouldFail    bool
	failOfferIDs  map[string]bool
	delay         time.Duration
	callCount     int
	calledOffers  []string
}

// NewMockPriceServiceClient crea un nuevo cliente mock
func NewMockPriceServiceClient() *MockPriceServiceClient {
	return &MockPriceServiceClient{
		shouldFail:   false,
		failOfferIDs: make(map[string]bool),
		delay:        100 * time.Millisecond,
		calledOffers: make([]string, 0),
	}
}

// Cancel implementa la interfaz PriceServiceClient
func (m *MockPriceServiceClient) Cancel(ctx context.Context, offerID string) error {
	m.callCount++
	m.calledOffers = append(m.calledOffers, offerID)

	// Simular delay de red
	select {
	case <-time.After(m.delay):
	case <-ctx.Done():
		return ctx.Err()
	}

	// Verificar si esta oferta específica debe fallar
	if m.failOfferIDs[offerID] {
		return fmt.Errorf("mock error for offer %s", offerID)
	}

	// Verificar si todas las llamadas deben fallar
	if m.shouldFail {
		return fmt.Errorf("mock service error")
	}

	return nil
}

// SetShouldFail configura si todas las llamadas deben fallar
func (m *MockPriceServiceClient) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// SetFailForOffer configura si una oferta específica debe fallar
func (m *MockPriceServiceClient) SetFailForOffer(offerID string, shouldFail bool) {
	if shouldFail {
		m.failOfferIDs[offerID] = true
	} else {
		delete(m.failOfferIDs, offerID)
	}
}

// SetDelay configura el delay simulado de red
func (m *MockPriceServiceClient) SetDelay(delay time.Duration) {
	m.delay = delay
}

// GetCallCount devuelve el número de llamadas realizadas
func (m *MockPriceServiceClient) GetCallCount() int {
	return m.callCount
}

// GetCalledOffers devuelve la lista de offers que fueron llamadas
func (m *MockPriceServiceClient) GetCalledOffers() []string {
	return m.calledOffers
}

// Reset reinicia las métricas del mock
func (m *MockPriceServiceClient) Reset() {
	m.callCount = 0
	m.calledOffers = make([]string, 0)
	m.failOfferIDs = make(map[string]bool)
	m.shouldFail = false
}