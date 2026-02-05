package domain

import "time"

type Order struct {
	ID        int64
	UserID    int64
	Symbol    string
	Side      Side
	Type      OrderType
	Price     float64
	Quantity  float64
	FilledQty float64
	Status    OrderStatus
	CreatedAt time.Time
}
