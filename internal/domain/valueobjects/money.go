package valueobjects

import (
	"fmt"
)

type Money struct {
	Amount   int64
	Currency string // ISO
}

func NewMoney(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", float64(m.Amount)/100, m.Currency)
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add money with different currencies: %s and %s",
			m.Currency, other.Currency)
	}

	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}
