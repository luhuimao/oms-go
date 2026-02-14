package engine

import (
	"oms-contract/internal/domain"
	"testing"
	"time"
)

func TestMatchingEngine_BasicMatch(t *testing.T) {
	engine := NewMatchingEngine()

	sell := &domain.Order{
		ID:        1,
		UserID:    100,
		Symbol:    "BTCUSDT",
		Side:      domain.Sell,
		Price:     30000,
		Quantity:  1,
		CreatedAt: time.Now(),
	}

	buy := &domain.Order{
		ID:        2,
		UserID:    200,
		Symbol:    "BTCUSDT",
		Side:      domain.Buy,
		Price:     31000,
		Quantity:  1,
		CreatedAt: time.Now(),
	}

	engine.SubmitOrder(sell)
	trades := engine.SubmitOrder(buy)
	// fmt.Printf(
	// 	"trades: %+v\n",
	// 	trades,
	// )
	t.Logf(
		"trades: %+v\n",
		trades)
	if len(trades) != 2 {
		t.Fatalf("expected 2 trades (maker+taker), got %d", len(trades))
	}

	if trades[0].Price != 30000 {
		t.Fatalf("unexpected trade price: %f", trades[0].Price)
	}
}
