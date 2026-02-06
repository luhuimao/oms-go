package domain

type Trade struct {
	TradeID int64
	OrderID int64
	Qty     float64
	Price   float64
	UserID  int64
	Symbol  string
	Side    Side
	IsMaker bool
}
