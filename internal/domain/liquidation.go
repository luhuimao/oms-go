package domain

type LiquidationOrder struct {
	OrderID     int64
	UserID      int64
	Symbol      string
	Side        Side
	Quantity    float64
	OrderType   OrderType // 永远是 MARKET
	TimeInForce string    // IOC
	Reason      string    // LIQUIDATION
}
