package entity

import (
	"clean-arq-layout/internal/domain/valueobjects"
)

type Product struct {
	ID              string
	Name            string
	BasePrice       valueobjects.Money
	DiscountPercent float64
	DiscountedPrice valueobjects.Money
}

func (p *Product) ApplyDiscount(percent float64) {
	p.DiscountPercent = percent
	discountAmount := p.BasePrice.Amount * int64(p.DiscountPercent) / 100

	p.DiscountedPrice = valueobjects.Money{
		Amount:   p.BasePrice.Amount - discountAmount,
		Currency: p.BasePrice.Currency,
	}
}
