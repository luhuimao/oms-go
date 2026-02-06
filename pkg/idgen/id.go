package idgen

import (
	"sync"
	"sync/atomic"
	"time"
)

type Generator struct {
	id int64
}

func New() *Generator {
	return &Generator{id: 1000}
}

func (g *Generator) Next() int64 {
	return atomic.AddInt64(&g.id, 1)
}

type TradeIDGen struct {
	mu       sync.Mutex
	lastTs   int64
	sequence int64
	nodeID   int64
}

func NewTradeIDGen(nodeID int64) *TradeIDGen {
	return &TradeIDGen{nodeID: nodeID}
}

func (g *TradeIDGen) Next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	ts := time.Now().UnixMilli()

	if ts == g.lastTs {
		g.sequence++
	} else {
		g.sequence = 0
		g.lastTs = ts
	}

	return (ts << 22) | (g.nodeID << 12) | g.sequence
}
