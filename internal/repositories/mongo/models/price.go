package models

import (
	"time"
)

type PriceModel struct {
	ID       string `bson:"_id"`
	Type     string `bson:"type"`
	Amount   int64  `bson:"amount"`   // Minor units (cents)
	Currency string `bson:"currency"` // ISO code
	SKU      string `bson:"sku"`

	// Dates - flattened structure
	CreateDate time.Time  `bson:"create_date"`
	UpdateDate time.Time  `bson:"update_date"`
	DropDate   *time.Time `bson:"drop_date,omitempty"`
	ValidFrom  time.Time  `bson:"valid_from"`
	ValidTo    *time.Time `bson:"valid_to,omitempty"`
}
