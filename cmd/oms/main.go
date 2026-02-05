package main

import (
	"time"

	"oms-contract/infra/matching"
	"oms-contract/internal/domain"
	"oms-contract/internal/engine"
	"oms-contract/internal/memory"
	"oms-contract/internal/service"
	"oms-contract/pkg/idgen"
)

func main() {
	orderBook := memory.NewOrderBook()
	positionBook := memory.NewPositionBook()

	dispatcher := engine.NewDispatcher(4)

	positionSvc := service.NewPositionService(positionBook)
	liquidationSvc := &service.LiquidationService{}
	var matchingGw service.MatchingGateway

	idGen := idgen.New()

	liqSvc := service.NewLiquidationService(matchingGw, idGen)

	orderSvc := service.NewOrderService(
		orderBook,
		positionSvc,
		liqSvc,
	)

	// mock matching 反向注入 OMS
	matchingGw = matching.NewMockMatching(orderSvc)

	// orderSvc := service.NewOrderService(
	// 	orderBook,
	// 	positionSvc,
	// 	liquidationSvc,
	// )

	ids := idgen.New()

	order := &domain.Order{
		ID:       ids.Next(),
		UserID:   1,
		Symbol:   "BTC-USDT",
		Side:     domain.Buy,
		Type:     domain.Limit,
		Price:    42000,
		Quantity: 1,
	}

	dispatcher.Dispatch(order.ID, func() {
		orderSvc.CreateOrder(order)
	})

	time.Sleep(time.Second)

	dispatcher.Dispatch(order.ID, func() {
		orderSvc.OnTrade(&domain.Trade{
			OrderID: order.ID,
			Qty:     1,
			Price:   42000,
		})
	})
}
