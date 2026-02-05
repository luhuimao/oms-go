package service

import (
	"fmt"

	"oms-contract/internal/domain"
	"oms-contract/pkg/idgen"
)

const (
	MaintenanceMarginRate = 0.005 // 0.5%
)

func (l *LiquidationService) Check(
	p *domain.Position,
	markPrice float64,
) bool {

	notional := abs(p.Qty) * markPrice
	mm := notional * MaintenanceMarginRate
	upnl := (markPrice - p.EntryPrice) * p.Qty

	equity := p.Margin + upnl

	return equity <= mm
}

func (l *LiquidationService) Liquidate(p *domain.Position) {
	fmt.Printf(
		"[LIQUIDATION] user=%d symbol=%s qty=%.2f\n",
		p.UserID, p.Symbol, p.Qty,
	)
}

type LiquidationService struct {
	matching MatchingGateway
	idGen    *idgen.Generator
}

func NewLiquidationService(
	matching MatchingGateway,
	idGen *idgen.Generator,
) *LiquidationService {
	return &LiquidationService{
		matching: matching,
		idGen:    idGen,
	}
}

func (l *LiquidationService) Execute(
	p *domain.Position,
) {

	side := domain.Sell
	if p.Qty < 0 {
		side = domain.Buy
	}

	order := &domain.LiquidationOrder{
		OrderID:     l.idGen.Next(),
		UserID:      p.UserID,
		Symbol:      p.Symbol,
		Side:        side,
		Quantity:    abs(p.Qty),
		OrderType:   domain.Market,
		TimeInForce: "IOC",
		Reason:      "LIQUIDATION",
	}

	fmt.Printf(
		"[OMS] send liquidation IOC order: %+v\n",
		order,
	)

	_ = l.matching.SendLiquidationOrder(order)
}
