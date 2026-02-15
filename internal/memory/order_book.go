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

// GetAll returns a copy of the current order map
// Note: In a real high-performance system, we might want to avoid full copies
// or use a copy-on-write structure, but for this implementation, a copy is safe.
func (b *OrderBook) GetAll() map[int64]*domain.Order {
	b.mu.RLock()
	defer b.mu.RUnlock()

	copy := make(map[int64]*domain.Order, len(b.orders))
	for k, v := range b.orders {
		copy[k] = v
	}
	return copy
}
