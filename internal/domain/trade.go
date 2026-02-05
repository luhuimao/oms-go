package domain

type Trade struct {
	OrderID int64
	Qty     float64
	Price   float64
	UserID  int64
	Symbol  string
}
