package memory

import (
	"sync"

	"oms-contract/internal/domain"
)

type OrderBook struct {
	mu     sync.RWMutex
	orders map[int64]*domain.Order
}

func NewOrderBook() *OrderBook {
	return &OrderBook{orders: make(map[int64]*domain.Order)}
}

func (b *OrderBook) Add(o *domain.Order) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.orders[o.ID] = o
}

func (b *OrderBook) Get(id int64) (*domain.Order, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	o, ok := b.orders[id]
	return o, ok
}
