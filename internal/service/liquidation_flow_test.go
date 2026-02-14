package service

import (
	"oms-contract/internal/domain"
	"oms-contract/internal/engine"
	"oms-contract/internal/memory"
	"testing"
)

func TestLiquidationFlow(t *testing.T) {
	// This test is deprecated as LiquidationEngine.OnMarkPrice method is commented out
	// Liquidation functionality is now handled by LiquidationService
	// See cmd/oms/main.go Scenario 4 for liquidation demonstration
	t.Skip("LiquidationEngine.OnMarkPrice is deprecated, see LiquidationService instead")

	match := engine.NewMatchingEngine()
	positionBook := memory.NewPositionBook()
	positionSvc := NewPositionService(positionBook)
	liq := NewLiquidationEngine(match, positionSvc)

	// Open a 10x long at 50000
	trade1 := &domain.Trade{
		OrderID: 1,
		UserID:  100,
		Symbol:  "BTCUSDT",
		Qty:     1,
		Price:   50000,
		IsMaker: false,
	}

	positionSvc.OnTrade(
		trade1.UserID,
		trade1.Symbol,
		trade1.Qty,
		trade1.Price,
		10,
	)

	// Note: OnMarkPrice is now commented out in liquidation_engine.go
	// This test would need to be rewritten to use LiquidationService
	_ = liq

	pos, _ := positionSvc.Get(100, "BTCUSDT")
	if pos == nil {
		t.Fatal("position should exist")
	}
}
