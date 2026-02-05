package domain

type Position struct {
	UserID     int64
	Symbol     string
	Qty        float64 // >0 多仓 <0 空仓
	EntryPrice float64
	Leverage   float64
	Margin     float64 // 当前保证金
}
