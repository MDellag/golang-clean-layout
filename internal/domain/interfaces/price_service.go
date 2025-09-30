package interfaces

import "context"

// PriceServiceClient define la interfaz para el cliente del servicio de precios
type PriceServiceClient interface {
	// Cancel cancela una oferta por su ID
	// Retorna error si la request no fue exitosa (status != 200)
	Cancel(ctx context.Context, offerID string) error
}