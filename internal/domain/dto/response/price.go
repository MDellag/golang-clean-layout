package response

import (
	"clean-arq-layout/internal/domain/constants"
	"time"
)

type PriceResponse struct {
	ID         string              `json:"id"`
	Type       constants.PriceType `json:"type"`
	Amount     int64               `json:"amount"`
	Currency   string              `json:"currency"`
	SKU        string              `json:"sku"`
	CreateDate time.Time           `json:"create_date"`
	UpdateDate time.Time           `json:"update_date"`
	DropDate   *time.Time          `json:"drop_date,omitempty"`
	ValidFrom  time.Time           `json:"valid_from"`
	ValidTo    *time.Time          `json:"valid_to,omitempty"`
	IsActive   bool                `json:"is_active"`
}
