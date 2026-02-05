package memory

import (
	"oms-contract/internal/domain"
	"strconv"
	"sync"
)

type PositionBook struct {
	mu        sync.RWMutex
	positions map[string]*domain.Position
}

func NewPositionBook() *PositionBook {
	return &PositionBook{
		positions: make(map[string]*domain.Position),
	}
}

func key(uid int64, symbol string) string {
	return symbol + ":" + strconv.FormatInt(uid, 10)

}

func (b *PositionBook) Get(uid int64, symbol string) (*domain.Position, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	p, ok := b.positions[key(uid, symbol)]
	return p, ok
}

func (b *PositionBook) Save(p *domain.Position) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.positions[key(p.UserID, p.Symbol)] = p
}
