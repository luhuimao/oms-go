package service

import (
	"fmt"
	"time"

	"oms-contract/internal/domain"
	"oms-contract/internal/memory"
	"oms-contract/internal/snapshot"
	"oms-contract/pkg/idgen"
)

type OrderService struct {
	book     *memory.OrderBook
	risk     *RiskService
	margin   *MarginService
	eventBus *snapshot.EventBus
	idGen    *idgen.Generator

	position   *PositionService // ✅ 必须有
	liquidator *LiquidationService
}

func NewOrderService(book *memory.OrderBook,
	pos *PositionService,
	liq *LiquidationService,
	eb *snapshot.EventBus,
	idGen *idgen.Generator) *OrderService {
	return &OrderService{
		book:       book,
		risk:       &RiskService{},
		margin:     &MarginService{},
		position:   pos,
		liquidator: liq,
		eventBus:   eb,
		idGen:      idGen,
	}
}

func (s *OrderService) CreateOrder(o *domain.Order) int64 {
	if err := s.risk.Check(o); err != nil {
		o.Status = domain.Rejected
		return 0
	}

	_ = s.margin.Freeze(o)

	o.ID = s.idGen.Next()
	o.Status = domain.Submitted
	o.CreatedAt = time.Now()

	// Publish event instead of direct book modification
	// The EventBus applies it to the state (which shares the book, or updates it)
	// If book is shared, this is fine.
	event := snapshot.NewEvent(
		0, // ID allocated by store
		snapshot.EventOrderCreated,
		snapshot.OrderCreatedData{Order: o},
	)

	if s.eventBus != nil {
		if err := s.eventBus.Publish(event); err != nil {
			fmt.Printf("[OMS] failed to publish order created event: %v\n", err)
			// Should we fail? For now just log.
		}
	} else {
		// Fallback for tests or if event bus is not configured (add to book directly?
		// No, NewOrderService assumes EventBus is the way.
		// If nil, maybe we just add to book directly here?
		// No, let's keep it simple: if eventBus is nil, we assume it's a test using the book directly elsewhere
		// OR we should fallback to direct book manipulation if we want the service to work without eventbus.
		// Given strict event sourcing, direct manipulation breaks pattern.
		// But for legacy tests...
		s.book.Add(o)
	}

	fmt.Printf("[OMS] order submitted: %+v\n", o)
	return o.ID
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
