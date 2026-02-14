package service

import (
	"oms-contract/internal/domain"
	"oms-contract/internal/engine"
)

type LiquidationEngine struct {
	matching *engine.MatchingEngine
	position *PositionService
}

func NewLiquidationEngine(
	matching *engine.MatchingEngine,
	position *PositionService,
) *LiquidationEngine {
	return &LiquidationEngine{
		matching: matching,
		position: position,
	}
}

// OnMarkPrice is commented out due to missing AllBySymbol method
// This functionality is handled by LiquidationService instead
/*
func (l *LiquidationEngine) OnMarkPrice(
	symbol string,
	markPrice float64,
) {
	positions := l.position.AllBySymbol(symbol)

	for _, p := range positions {
		if !p.ShouldLiquidate(markPrice) {
			continue
		}
		tradeIDGen := idgen.NewTradeIDGen(1)
		tradeID := tradeIDGen.Next()
		order := &domain.Order{
			ID:        tradeID,
			UserID:    p.UserID,
			Symbol:    p.Symbol,
			Side:      oppositeSide(p.Side),
			Price:     aggressivePrice(p.Side, markPrice),
			Quantity:  abs(p.Size),
			Type:      domain.IOC,
			IsSystem:  true,
			CreatedAt: time.Now(),
		}

		trades := l.matching.SubmitOrder(order)

		for _, t := range trades {
			l.position.OnTrade(t.UserID, t.Symbol, t.Qty, t.Price, 0)
		}
	}
}
*/

func oppositeSide(s domain.Side) domain.Side {
	if s == domain.Buy {
		return domain.Sell
	}
	return domain.Buy
}

func aggressivePrice(side domain.Side, mark float64) float64 {
	if side == domain.Buy {
		return mark * 2 // 强平卖单：极低价吃单
	}
	return mark / 2 // 强平买单：极高价吃单
}

// func abs(v float64) float64 {
// 	if v < 0 {
// 		return -v
// 	}
// 	return v
// }
