package jobs

import (
	"context"
	"fmt"
	"time"

	"clean-arq-layout/internal/domain/interfaces"
	"clean-arq-layout/internal/workers/types"
)

// OfferCancelJob job para cancelar ofertas usando un cliente de servicio inyectado
type OfferCancelJob struct {
	id              string
	offerID         string
	priceService    interfaces.PriceServiceClient
	responseChannel chan<- types.JobResult
	maxRetries      int
	currentRetry    int
}

// NewOfferCancelJob crea un nuevo job de cancelación de oferta
func NewOfferCancelJob(id, offerID string, priceService interfaces.PriceServiceClient, responseChannel chan<- types.JobResult) *OfferCancelJob {
	return &OfferCancelJob{
		id:              id,
		offerID:         offerID,
		priceService:    priceService,
		responseChannel: responseChannel,
		maxRetries:      3,
	}
}

// Execute implementa la interfaz Job
func (j *OfferCancelJob) Execute(ctx context.Context) error {
	for j.currentRetry <= j.maxRetries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Llamar al méthodo Cancel del cliente de servicio
		err := j.priceService.Cancel(ctx, j.offerID)
		if err != nil {
			j.currentRetry++
			if j.currentRetry <= j.maxRetries {
				// Backoff exponencial
				backoff := time.Duration(j.currentRetry*j.currentRetry) * time.Second
				timer := time.NewTimer(backoff)

				select {
				case <-timer.C:
					continue
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				}
			} else {
				return fmt.Errorf("offer cancellation failed after %d attempts: %w", j.maxRetries, err)
			}
		} else {
			// Éxito
			return nil
		}
	}

	return fmt.Errorf("offer cancellation exhausted all retry attempts for offer %s", j.offerID)
}

// Name implementa la interfaz Job
func (j *OfferCancelJob) Name() string {
	return fmt.Sprintf("offer-cancel-%s", j.offerID)
}

// Priority implementa la interfaz Job
func (j *OfferCancelJob) Priority() int {
	return 3 // Alta prioridad para cancelaciones
}

// ID implementa la interfaz JobWithResponse
func (j *OfferCancelJob) ID() string {
	return j.id
}

// ResponseChannel implementa la interfaz JobWithResponse
func (j *OfferCancelJob) ResponseChannel() chan<- types.JobResult {
	return j.responseChannel
}

// GetOfferID devuelve el ID de la oferta
func (j *OfferCancelJob) GetOfferID() string {
	return j.offerID
}

// SetMaxRetries permite configurar el número máximo de reintentos
func (j *OfferCancelJob) SetMaxRetries(maxRetries int) {
	j.maxRetries = maxRetries
}
