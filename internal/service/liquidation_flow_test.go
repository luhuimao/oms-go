package service

import (
	"testing"
	"time"

	"oms-contract/internal/domain"
	"oms-contract/internal/engine"
)

func TestLiquidation_IOC_Flow(t *testing.T) {
	match := engine.NewMatchingEngine()

	positionSvc := NewPositionService()
	liq := NewLiquidationEngine(match, positionSvc)

	// 1️⃣ Maker 挂卖单（提供流动性）
	match.SubmitOrder(&domain.Order{
		ID:        1,
		UserID:    999,
		Symbol:    "BTCUSDT",
		Side:      domain.Sell,
		Price:     20000,
		Quantity:  10,
		CreatedAt: time.Now(),
	})

	// 2️⃣ 用户高杠杆做多
	openTrades := match.SubmitOrder(&domain.Order{
		ID:        2,
		UserID:    100,
		Symbol:    "BTCUSDT",
		Side:      domain.Buy,
		Price:     20000,
		Quantity:  5,
		CreatedAt: time.Now(),
	})

	for _, tr := range openTrades {
		positionSvc.OnTrade(tr)
	}

	pos, ok := positionSvc.Get(100, "BTCUSDT")
	if !ok {
		t.Fatal("position not created")
	}

	// 模拟极低保证金
	pos.Margin = 100

	// 3️⃣ 价格暴跌 → 触发强平
	markPrice := int64(15000)
	liq.OnMarkPrice("BTCUSDT", markPrice)

	// 4️⃣ 检查仓位是否被平掉
	pos, ok = positionSvc.Get(100, "BTCUSDT")
	if ok && pos.Size != 0 {
		t.Fatalf("position not liquidated, size=%d", pos.Size)
	}
}
