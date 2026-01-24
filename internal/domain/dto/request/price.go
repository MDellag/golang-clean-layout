package request

import (
	"clean-arq-layout/internal/domain/constants"
	"time"
)

type CreatePriceRequest struct {
	Type      constants.PriceType `json:"type" validate:"required"`
	Amount    int64               `json:"amount" validate:"required"`    // In minor units (cents)
	Currency  string              `json:"currency" validate:"required"`  // ISO code
	SKU       string              `json:"sku" validate:"required"`
	ValidFrom time.Time           `json:"valid_from" validate:"required"`
	ValidTo   *time.Time          `json:"valid_to"` // Nullable
}

type UpdatePriceRequest struct {
	Amount    *int64     `json:"amount"`
	Currency  *string    `json:"currency"`
	ValidFrom *time.Time `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to"`
}
