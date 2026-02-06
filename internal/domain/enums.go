package domain

type Side string
type OrderType string
type OrderStatus string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"

	Limit      OrderType   = "LIMIT"
	Market     OrderType   = "MARKET"
	IOC        OrderType   = "IOC"
	Submitted  OrderStatus = "SUBMITTED"
	PartFilled OrderStatus = "PART_FILLED"
	Filled     OrderStatus = "FILLED"
	Canceled   OrderStatus = "CANCELED"
	Rejected   OrderStatus = "REJECTED"
)
