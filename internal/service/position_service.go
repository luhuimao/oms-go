package service

import (
	"oms-contract/internal/domain"
	"oms-contract/internal/memory"
)

type PositionService struct {
	book *memory.PositionBook
}

func NewPositionService(book *memory.PositionBook) *PositionService {
	return &PositionService{book: book}
}

func (s *PositionService) Get(
	uid int64,
	symbol string,
) (*domain.Position, bool) {
	return s.book.Get(uid, symbol)
}

func (s *PositionService) OnTrade(
	userID int64,
	symbol string,
	qty float64,
	price float64,
	leverage float64,
) {

	p, ok := s.book.Get(userID, symbol)
	if !ok {
		// 开新仓
		notional := abs(qty) * price
		margin := notional / leverage

		p = &domain.Position{
			UserID:     userID,
			Symbol:     symbol,
			Qty:        qty,
			EntryPrice: price,
			Leverage:   leverage,
			Margin:     margin,
		}
	} else {
		// 加仓（简化：同方向）
		p.EntryPrice = (p.EntryPrice*p.Qty + price*qty) / (p.Qty + qty)
		p.Qty += qty
	}

	s.book.Save(p)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
