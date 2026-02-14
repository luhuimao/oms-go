package engine_test

import (
	"sync"
	"testing"
	"time"

	"oms-contract/internal/domain"
	"oms-contract/internal/engine"
	"oms-contract/pkg/idgen"

	"github.com/stretchr/testify/require"
)

var orderIDGen = idgen.New()

// Helper function to create a limit order
func newLimitOrder(symbol string, side domain.Side, price, qty float64) *domain.Order {
	return &domain.Order{
		ID:        orderIDGen.Next(),
		UserID:    100,
		Symbol:    symbol,
		Side:      side,
		Type:      domain.Limit,
		Price:     price,
		Quantity:  qty,
		Status:    domain.Submitted,
		CreatedAt: time.Now(),
	}
}

func Test_ShardedMatchingEngine_BasicMatch(t *testing.T) {
	e := engine.NewShardedMatchingEngine(4)

	// Create sell order first (maker)
	sell := newLimitOrder("BTCUSDT", domain.Sell, 30000, 1)
	trades1 := e.Submit(sell)
	require.Len(t, trades1, 0) // No match yet

	// Create buy order (taker)
	buy := newLimitOrder("BTCUSDT", domain.Buy, 31000, 1)
	trades2 := e.Submit(buy)

	// Should have 2 trades (one for maker, one for taker)
	require.Len(t, trades2, 2)
	require.Equal(t, "BTCUSDT", trades2[0].Symbol)
	require.Equal(t, 30000.0, trades2[0].Price)
	require.Equal(t, 1.0, trades2[0].Qty)
}

func Test_ShardedMatchingEngine_PriceTimePriority(t *testing.T) {
	e := engine.NewShardedMatchingEngine(2)

	// Same price, first come first served
	e.Submit(newLimitOrder("ETHUSDT", domain.Sell, 2000, 5)) // maker1
	time.Sleep(time.Millisecond)
	e.Submit(newLimitOrder("ETHUSDT", domain.Sell, 2000, 5)) // maker2

	trades := e.Submit(newLimitOrder("ETHUSDT", domain.Buy, 2000, 10)) // taker

	// Should have 4 trades total (2 makers × 2 trades each)
	require.Len(t, trades, 4)

	// Verify quantities
	require.Equal(t, 5.0, trades[0].Qty)
	require.Equal(t, 5.0, trades[2].Qty)
}

func Test_ShardedMatchingEngine_ConcurrentSymbols(t *testing.T) {
	e := engine.NewShardedMatchingEngine(8)

	symbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "XRPUSDT"}

	var wg sync.WaitGroup
	var mu sync.Mutex
	totalTrades := 0

	for _, sym := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()

			for i := 0; i < 100; i++ {
				// Submit sell first
				e.Submit(newLimitOrder(symbol, domain.Sell, 1000, 1))
				// Then buy to match
				trades := e.Submit(newLimitOrder(symbol, domain.Buy, 1000, 1))

				mu.Lock()
				totalTrades += len(trades)
				mu.Unlock()
			}
		}(sym)
	}

	wg.Wait()

	// Each symbol: 100 iterations × 2 trades per match = 200 trades per symbol
	// 4 symbols × 200 = 800 trades total
	require.Equal(t, 800, totalTrades)
}

func Test_ShardedMatchingEngine_DeterministicReplay(t *testing.T) {
	run := func() []*domain.Trade {
		e := engine.NewShardedMatchingEngine(4)

		e.Submit(newLimitOrder("BTCUSDT", domain.Sell, 100, 5))
		e.Submit(newLimitOrder("BTCUSDT", domain.Sell, 101, 5))
		trades := e.Submit(newLimitOrder("BTCUSDT", domain.Buy, 101, 10))

		return trades
	}

	trades1 := run()
	trades2 := run()

	// Verify both runs produce the same number of trades
	require.Equal(t, len(trades1), len(trades2))

	// Verify trade details match
	for i := range trades1 {
		require.Equal(t, trades1[i].Symbol, trades2[i].Symbol)
		require.Equal(t, trades1[i].Price, trades2[i].Price)
		require.Equal(t, trades1[i].Qty, trades2[i].Qty)
		require.Equal(t, trades1[i].Side, trades2[i].Side)
	}
}

func Test_ShardedMatchingEngine_Stress(t *testing.T) {
	e := engine.NewShardedMatchingEngine(16)

	const total = 10_000
	totalTrades := 0

	for i := 0; i < total; i++ {
		e.Submit(newLimitOrder("BTCUSDT", domain.Sell, 1000, 1))
		trades := e.Submit(newLimitOrder("BTCUSDT", domain.Buy, 1000, 1))
		totalTrades += len(trades)
	}

	// Each match produces 2 trades (maker + taker)
	require.Equal(t, total*2, totalTrades)
}

func Test_ShardedMatchingEngine_NoMatch(t *testing.T) {
	e := engine.NewShardedMatchingEngine(4)

	// Create buy order at low price
	trades1 := e.Submit(newLimitOrder("BTCUSDT", domain.Buy, 29000, 1))
	require.Len(t, trades1, 0)

	// Create sell order at high price (no match)
	trades2 := e.Submit(newLimitOrder("BTCUSDT", domain.Sell, 30000, 1))
	require.Len(t, trades2, 0)
}

func Test_ShardedMatchingEngine_PartialFill(t *testing.T) {
	e := engine.NewShardedMatchingEngine(4)

	// Maker: sell 10 @ 1000
	e.Submit(newLimitOrder("BTCUSDT", domain.Sell, 1000, 10))

	// Taker: buy 5 @ 1000 (partial fill)
	trades := e.Submit(newLimitOrder("BTCUSDT", domain.Buy, 1000, 5))

	require.Len(t, trades, 2)
	require.Equal(t, 5.0, trades[0].Qty)
	require.Equal(t, 5.0, trades[1].Qty)
}
