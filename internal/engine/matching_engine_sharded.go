package engine

import (
	"hash/fnv"

	"oms-contract/internal/domain"
)

// =============================
// , ShardedSingle-Threaded Matching Engine
// =============================

// ShardedMatchingEngine dispatches orders by symbol hash
// Each shard is strictly single-threaded, guaranteeing determinism
// and eliminating locks inside the order book.
type ShardedMatchingEngine struct {
	shards []*engineShard
}

func NewShardedMatchingEngine(shardNum int) *ShardedMatchingEngine {
	shards := make([]*engineShard, shardNum)
	for i := 0; i < shardNum; i++ {
		shards[i] = newEngineShard(i)
	}
	return &ShardedMatchingEngine{shards: shards}
}

func (e *ShardedMatchingEngine) Submit(order *domain.Order) []*domain.Trade {
	shard := e.pickShard(order.Symbol)
	return shard.submit(order)
}

func (e *ShardedMatchingEngine) pickShard(symbol string) *engineShard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(symbol))
	idx := int(h.Sum32()) % len(e.shards)
	return e.shards[idx]
}

// =============================
// engineShard (single goroutine)
// =============================

type engineShard struct {
	id     int
	inCh   chan *submitReq
	books  map[string]*OrderBook
	closed chan struct{}
}

type submitReq struct {
	order *domain.Order
	resp  chan []*domain.Trade
}

func newEngineShard(id int) *engineShard {
	s := &engineShard{
		id:     id,
		inCh:   make(chan *submitReq, 1024),
		books:  make(map[string]*OrderBook),
		closed: make(chan struct{}),
	}

	go s.loop()
	return s
}

func (s *engineShard) loop() {
	for {
		select {
		case req := <-s.inCh:
			book := s.getBook(req.order.Symbol)
			trades := book.Match(req.order)
			req.resp <- trades
		case <-s.closed:
			return
		}
	}
}

func (s *engineShard) submit(order *domain.Order) []*domain.Trade {
	resp := make(chan []*domain.Trade, 1)
	s.inCh <- &submitReq{order: order, resp: resp}
	return <-resp
}

func (s *engineShard) getBook(symbol string) *OrderBook {
	book, ok := s.books[symbol]
	if !ok {
		book = NewOrderBook(symbol)
		s.books[symbol] = book
	}
	return book
}

// =============================
// Lifecycle
// =============================

func (e *ShardedMatchingEngine) Close() {
	for _, s := range e.shards {
		close(s.closed)
	}
}

// =============================
// Compile-time check
// =============================

var _ interface {
	Submit(*domain.Order) []*domain.Trade
} = (*ShardedMatchingEngine)(nil)
