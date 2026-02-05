package matching

import (
	"fmt"
	"time"

	"oms-contract/internal/domain"
	"oms-contract/internal/service"
)

type MockMatching struct {
	orderSvc *service.OrderService
}

func NewMockMatching(orderSvc *service.OrderService) *MockMatching {
	return &MockMatching{orderSvc: orderSvc}
}

func (m *MockMatching) SendLiquidationOrder(
	o *domain.LiquidationOrder,
) error {

	fmt.Printf(
		"[MATCHING] IOC liquidation order received: %+v\n",
		o,
	)

	// 模拟立即成交
	time.Sleep(10 * time.Millisecond)

	trade := &domain.Trade{
		OrderID: o.OrderID,
		Qty:     o.Quantity,
		Price:   mockMarketPrice(o.Symbol),
	}

	m.orderSvc.OnTrade(trade)
	return nil
}

func mockMarketPrice(symbol string) float64 {
	return 38000 // demo
}
