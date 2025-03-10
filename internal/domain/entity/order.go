package entity

import (
	"clean-arq-layout/internal/domain/constants"
	"clean-arq-layout/internal/domain/valueobjects"
	"time"
)

type Order struct {
	ID          string
	Customer    Customer
	Items       []OrderItem
	Status      constants.OrderStatus
	TotalAmount valueobjects.Money
	CreatedAt   time.Time
}

type OrderItem struct{}

// AddItem domain methods
func (o *Order) AddItem(item OrderItem) {
	o.Items = append(o.Items, item)
	o.recalculateTotal()
}

func (o *Order) recalculateTotal() {
	// not implemented yet
}
