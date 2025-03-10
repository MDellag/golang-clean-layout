package models

type ProductModel struct {
	ID              string  `db:"id"`
	Name            string  `db:"name"`
	BasePrice       int64   `db:"base_price"`
	Currency        string  `db:"currency"`
	DiscountPercent float64 `db:"discount_percent"`
}
