package service

import (
	"fmt"
	"time"

	"oms-contract/internal/domain"
	"oms-contract/internal/memory"
)

type OrderService struct {
	book   *memory.OrderBook
	risk   *RiskService
	margin *MarginService

	position   *PositionService // ✅ 必须有
	liquidator *LiquidationService
}

func NewOrderService(book *memory.OrderBook,
	pos *PositionService,
	liq *LiquidationService) *OrderService {
	return &OrderService{
		book:       book,
		risk:       &RiskService{},
		margin:     &MarginService{},
		position:   pos,
		liquidator: liq,
	}
}

func (s *OrderService) CreateOrder(o *domain.Order) {
	if err := s.risk.Check(o); err != nil {
		o.Status = domain.Rejected
		return
	}

	_ = s.margin.Freeze(o)

	o.Status = domain.Submitted
	o.CreatedAt = time.Now()
	s.book.Add(o)

	fmt.Printf("[OMS] order submitted: %+v\n", o)
}

func (s *OrderService) OnTrade(t *domain.Trade) {
	o, ok := s.book.Get(t.OrderID)
	if ok {
		// 普通订单成交
		o.FilledQty += t.Qty
		if o.FilledQty >= o.Quantity {
			o.Status = domain.Filled
		}
	}

	// 更新仓位（正负 qty）
	s.position.OnTrade(
		t.UserID,
		t.Symbol,
		signedQty(o.Side, t.Qty),
		t.Price,
		10,
	)

	// 成交后立即做强平检查
	p, ok := s.position.Get(t.UserID, t.Symbol)
	if ok && s.liquidator.Check(p, t.Price) {
		s.liquidator.Execute(p)
	}
}

func signedQty(side domain.Side, qty float64) float64 {
	if side == domain.Sell {
		return -qty
	}
	return qty
}
