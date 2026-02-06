package engine

import (
	"container/heap"
	"oms-contract/internal/domain"
	"oms-contract/pkg/idgen"
	"sync"
)

// MatchingEngine implements a price-time priority order matching engine
// It is production-oriented but simplified for clarity.
type MatchingEngine struct {
	mu    sync.Mutex
	books map[string]*OrderBook
}

func NewMatchingEngine() *MatchingEngine {
	return &MatchingEngine{
		books: make(map[string]*OrderBook),
	}
}

func (m *MatchingEngine) getBook(symbol string) *OrderBook {
	book, ok := m.books[symbol]
	if !ok {
		book = NewOrderBook(symbol)
		m.books[symbol] = book
	}
	return book
}

// SubmitOrder sends an order into the matching engine
func (m *MatchingEngine) SubmitOrder(order *domain.Order) []*domain.Trade {
	m.mu.Lock()
	defer m.mu.Unlock()

	book := m.getBook(order.Symbol)
	return book.Match(order)
}

// ================= OrderBook =================

type OrderBook struct {
	symbol string
	bids   *PriceHeap
	asks   *PriceHeap
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		symbol: symbol,
		bids:   NewPriceHeap(domain.Buy),
		asks:   NewPriceHeap(domain.Sell),
	}
}

func (ob *OrderBook) Match(order *domain.Order) []*domain.Trade {
	trades := make([]*domain.Trade, 0)

	var bookSide *PriceHeap
	if order.Side == domain.Buy {
		bookSide = ob.asks
	} else {
		bookSide = ob.bids
	}

	for order.Quantity > 0 && bookSide.Len() > 0 {
		best := heap.Pop(bookSide).(*domain.Order)

		// price check
		if order.Side == domain.Buy && order.Price < best.Price {
			heap.Push(bookSide, best)
			break
		}
		if order.Side == domain.Sell && order.Price > best.Price {
			heap.Push(bookSide, best)
			break
		}

		qty := min(order.Quantity, best.Quantity)

		// taker trade
		takerTrade := &domain.Trade{
			TradeID: genTradeID(),
			OrderID: order.ID,
			UserID:  order.UserID,
			Symbol:  order.Symbol,
			Side:    order.Side,
			Price:   best.Price,
			Qty:     qty,
			IsMaker: false,
		}

		// maker trade
		makerTrade := &domain.Trade{
			TradeID: genTradeID(),
			OrderID: best.ID,
			UserID:  best.UserID,
			Symbol:  best.Symbol,
			Side:    best.Side,
			Price:   best.Price,
			Qty:     qty,
			IsMaker: true,
		}

		trades = append(trades, takerTrade, makerTrade)

		order.Quantity -= qty
		best.Quantity -= qty

		if best.Quantity > 0 {
			heap.Push(bookSide, best)
		}
	}

	// Non-IOC orders can rest on book
	if order.Quantity > 0 && order.Type != domain.IOC {
		if order.Side == domain.Buy {
			heap.Push(ob.bids, order)
		} else {
			heap.Push(ob.asks, order)
		}
	}

	return trades
}

// ================= PriceHeap =================

type PriceHeap struct {
	side   domain.Side
	orders []*domain.Order
}

func NewPriceHeap(side domain.Side) *PriceHeap {
	h := &PriceHeap{side: side}
	heap.Init(h)
	return h
}

func (h PriceHeap) Len() int { return len(h.orders) }

func (h PriceHeap) Less(i, j int) bool {
	if h.orders[i].Price == h.orders[j].Price {
		return h.orders[i].CreatedAt.Before(h.orders[j].CreatedAt)
	}
	if h.side == domain.Buy {
		return h.orders[i].Price > h.orders[j].Price
	}
	return h.orders[i].Price < h.orders[j].Price
}

func (h PriceHeap) Swap(i, j int) {
	h.orders[i], h.orders[j] = h.orders[j], h.orders[i]
}

func (h *PriceHeap) Push(x any) {
	h.orders = append(h.orders, x.(*domain.Order))
}

func (h *PriceHeap) Pop() any {
	n := len(h.orders)
	item := h.orders[n-1]
	h.orders = h.orders[:n-1]
	return item
}

// ================= helpers =================

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func buyID(a, b *domain.Order) int64 {
	if a.Side == domain.Buy {
		return a.ID
	}
	return b.ID
}

func sellID(a, b *domain.Order) int64 {
	if a.Side == domain.Sell {
		return a.ID
	}
	return b.ID
}

func genTradeID() int64 {
	tradeIDGen := idgen.NewTradeIDGen(1)
	tradeID := tradeIDGen.Next()
	return tradeID
}
