package models

import (
	"time"
)

type OrderModel struct {
	ID          string    `db:"id"`
	CustomerID  string    `db:"customer_id"`
	Status      string    `db:"status"`
	TotalAmount int64     `db:"total_amount"`
	Currency    string    `db:"currency"`
	CreatedAt   time.Time `db:"created_at"`
}
