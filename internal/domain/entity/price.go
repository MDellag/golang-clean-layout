package entity

import (
	"clean-arq-layout/internal/domain/constants"
	"errors"
	"strings"
	"time"

	"github.com/govalues/money"
)

// PriceDate holds all date-related metadata for a price
type PriceDate struct {
	CreateDate time.Time
	UpdateDate time.Time
	DropDate   *time.Time // Nullable - when price was deprecated
	ValidFrom  time.Time
	ValidTo    *time.Time // Nullable - indefinite if nil
}

// Price represents a product price with validity period
type Price struct {
	ID    string
	Type  constants.PriceType
	Value money.Amount
	SKU   string
	Date  PriceDate
}

// IsActive checks if price is valid at given time
func (p *Price) IsActive(at time.Time) bool {
	// Check if price is valid at given time
	if at.Before(p.Date.ValidFrom) {
		return false
	}
	if p.Date.ValidTo != nil && at.After(*p.Date.ValidTo) {
		return false
	}
	return p.Date.DropDate == nil
}

// Validate performs business validation on the price
func (p *Price) Validate() error {
	if p.ID == "" {
		return errors.New("price ID is required")
	}
	if !strings.HasPrefix(p.ID, "price-") {
		return errors.New("price ID must start with 'price-'")
	}
	if p.SKU == "" {
		return errors.New("SKU is required")
	}
	if p.Date.ValidTo != nil && p.Date.ValidFrom.After(*p.Date.ValidTo) {
		return errors.New("ValidFrom must be before ValidTo")
	}
	return nil
}

// Drop marks the price as deprecated at the given time
func (p *Price) Drop(at time.Time) {
	p.Date.DropDate = &at
	p.Date.UpdateDate = at
}
