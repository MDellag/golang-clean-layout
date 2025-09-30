package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PriceServiceHTTPClient implementa PriceServiceClient usando HTTP
type PriceServiceHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	reason     string
}

// OfferCancelRequest representa la estructura de la request
type OfferCancelRequest struct {
	OfferID string `json:"offer_id"`
	Reason  string `json:"reason,omitempty"`
}

// NewPriceServiceHTTPClient crea un nuevo cliente HTTP para el servicio de precios
func NewPriceServiceHTTPClient(baseURL, reason string) *PriceServiceHTTPClient {
	return &PriceServiceHTTPClient{
		baseURL: baseURL,
		reason:  reason,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Cancel implementa la interfaz PriceServiceClient
func (c *PriceServiceHTTPClient) Cancel(ctx context.Context, offerID string) error {
	// Preparar request body
	requestBody := OfferCancelRequest{
		OfferID: offerID,
		Reason:  c.reason,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Crear request HTTP
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// Agregar headers de autenticación si es necesario
	// req.Header.Set("Authorization", "Bearer "+token)

	// Realizar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Verificar status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// SetHTTPClient permite configurar un cliente HTTP personalizado
func (c *PriceServiceHTTPClient) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// SetReason permite cambiar la razón de cancelación
func (c *PriceServiceHTTPClient) SetReason(reason string) {
	c.reason = reason
}