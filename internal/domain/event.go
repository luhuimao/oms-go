package domain

type OrderEvent struct {
	Type  string // NEW_ORDER
	Order *LiquidationOrder
}

type TradeEvent struct {
	OrderID int64
	UserID  int64
	Symbol  string
	Qty     float64
	Price   float64
}
