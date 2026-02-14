# 订单撮合引擎详解

## 概述

订单撮合引擎（Matching Engine）是交易所的核心组件，负责将买单和卖单按照**价格-时间优先**原则进行匹配，生成成交记录。Atlas OMS 的撮合引擎采用基于**堆（Heap）**的高效实现，确保了 O(log N) 的订单插入和匹配性能。

---

## 核心组件架构

```
┌──────────────────────────────────────────────┐
│         MatchingEngine                       │
│  ┌────────────────────────────────────┐      │
│  │  books: map[symbol]*OrderBook      │      │
│  └────────────────────────────────────┘      │
│              │                               │
│              ▼                               │
│  ┌───────────────────────────────────┐       │
│  │      OrderBook (BTCUSDT)          │       │
│  │  ┌──────────┐      ┌──────────┐   │       │
│  │  │   Bids   │      │   Asks   │   │       │
│  │  │  (买单)  │      │  (卖单)  │   │       │
│  │  └──────────┘      └──────────┘   │       │
│  │       │                  │         │       │
│  │       ▼                  ▼         │       │
│  │  PriceHeap          PriceHeap      │       │
│  │  (Max Heap)         (Min Heap)     │       │
│  └───────────────────────────────────┘       │
└──────────────────────────────────────────────┘
```

### 1. MatchingEngine（主引擎）

```go
type MatchingEngine struct {
    mu    sync.Mutex              // 全局锁
    books map[string]*OrderBook   // Symbol -> OrderBook 映射
}
```

**职责**：
- 管理所有交易对的订单簿
- 提供线程安全的订单提交接口
- 自动创建新交易对的订单簿

**关键方法**：
```go
func (m *MatchingEngine) SubmitOrder(order *domain.Order) []*domain.Trade
```

### 2. OrderBook（订单簿）

```go
type OrderBook struct {
    symbol string       // 交易对符号
    bids   *PriceHeap  // 买单堆（最高价在顶部）
    asks   *PriceHeap  // 卖单堆（最低价在顶部）
}
```

**职责**：
- 维护单个交易对的买卖挂单
- 执行订单撮合逻辑
- 生成成交记录

### 3. PriceHeap（价格堆）

```go
type PriceHeap struct {
    side   domain.Side      // BUY 或 SELL
    orders []*domain.Order  // 订单数组
}
```

**职责**：
- 实现 Go 标准库的 `heap.Interface`
- 维护价格-时间优先的订单排序
- 提供 O(log N) 的插入和弹出性能

---

## 价格-时间优先算法

### 核心原则

1. **价格优先**：
   - 买单：价格越高越优先
   - 卖单：价格越低越优先

2. **时间优先**：
   - 相同价格下，先提交的订单优先成交

### 堆排序实现

```go
func (h PriceHeap) Less(i, j int) bool {
    // 1. 价格不同时，按价格排序
    if h.orders[i].Price != h.orders[j].Price {
        if h.side == domain.Buy {
            return h.orders[i].Price > h.orders[j].Price  // 买单：最高价在顶
        }
        return h.orders[i].Price < h.orders[j].Price      // 卖单：最低价在顶
    }
    
    // 2. 价格相同时，按时间排序（早的在前）
    return h.orders[i].CreatedAt.Before(h.orders[j].CreatedAt)
}
```

### 示例：买单堆

```
订单簿买单（Bids）- Max Heap
┌─────────────────────────────────┐
│   Price: 31000  (最高价, 堆顶)   │  ← 优先成交
│   Time:  10:00:00               │
├─────────────────────────────────┤
│   Price: 31000  (相同价格)       │  ← 时间稍晚
│   Time:  10:00:01               │
├─────────────────────────────────┤
│   Price: 30000                  │
│   Time:  10:00:02               │
└─────────────────────────────────┘
```

---

## 订单撮合流程

### 完整流程图

```
新订单到达
    │
    ▼
┌─────────────────┐
│ SubmitOrder()   │
│  加锁保护       │
└────────┬────────┘
         │
         ▼
┌──────────────────┐
│ 获取/创建订单簿  │
│ getBook(symbol)  │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Match(order)    │  ← 核心撮合逻辑
└────────┬─────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
 成功撮合   无法撮合
    │         │
    ▼         ▼
 生成成交   挂单到簿
 返回Trades  (非IOC)
```

### 撮合算法详解

```go
func (ob *OrderBook) Match(order *domain.Order) []*domain.Trade {
    trades := make([]*domain.Trade, 0)
    
    // 1. 选择对手盘
    var bookSide *PriceHeap
    if order.Side == domain.Buy {
        bookSide = ob.asks  // 买单对卖单
    } else {
        bookSide = ob.bids  // 卖单对买单
    }
    
    // 2. 循环撮合
    for order.Quantity > 0 && bookSide.Len() > 0 {
        best := heap.Pop(bookSide).(*domain.Order)  // 取对手盘最优价
        
        // 3. 价格检查
        if !canMatch(order, best) {
            heap.Push(bookSide, best)  // 价格不匹配，放回
            break
        }
        
        // 4. 计算成交数量
        qty := min(order.Quantity, best.Quantity)
        
        // 5. 生成成交记录（Taker + Maker）
        trades = append(trades, 
            newTakerTrade(order, best.Price, qty),
            newMakerTrade(best, qty))
        
        // 6. 更新订单数量
        order.Quantity -= qty
        best.Quantity -= qty
        
        // 7. Maker 订单未完全成交，放回订单簿
        if best.Quantity > 0 {
            heap.Push(bookSide, best)
        }
    }
    
    // 8. Taker 订单未完全成交且非 IOC，挂单
    if order.Quantity > 0 && order.Type != domain.IOC {
        if order.Side == domain.Buy {
            heap.Push(ob.bids, order)
        } else {
            heap.Push(ob.asks, order)
        }
    }
    
    return trades
}
```

### 价格匹配规则

```go
func canMatch(taker, maker *domain.Order) bool {
    if taker.Side == domain.Buy {
        // 买单：出价 >= 卖单价格才能成交
        return taker.Price >= maker.Price
    } else {
        // 卖单：出价 <= 买单价格才能成交
        return taker.Price <= maker.Price
    }
}
```

---

## 成交记录生成

### Trade 结构

```go
type Trade struct {
    TradeID int64       // 成交 ID
    OrderID int64       // 订单 ID
    UserID  int64       // 用户 ID
    Symbol  string      // 交易对
    Side    Side        // BUY/SELL
    Price   float64     // 成交价格（总是 Maker 的价格）
    Qty     float64     // 成交数量
    IsMaker bool        // 是否为 Maker
}
```

### Taker vs Maker

| 角色 | 定义 | 成交价格 | 手续费（通常） |
|------|------|----------|--------------|
| **Taker** | 主动吃单的订单 | Maker 的价格 | 较高（如 0.10%） |
| **Maker** | 被动挂单的订单 | 自己的价格 | 较低（如 0.05%） |

### 成交示例

```go
// 场景：买单吃卖单
Maker (挂单): SELL 1 BTC @ 30000
Taker (新单): BUY  1 BTC @ 31000

// 成交结果：
成交价格: 30000 (Maker 价格)
成交数量: 1 BTC

// 生成 2 条 Trade 记录：
Trade 1 (Taker):
  - OrderID: Taker订单ID
  - Price: 30000
  - Qty: 1
  - IsMaker: false

Trade 2 (Maker):
  - OrderID: Maker订单ID
  - Price: 30000
  - Qty: 1
  - IsMaker: true
```

---

## 订单类型处理

### 1. Limit Order（限价单）

```go
order := &Order{
    Type:     domain.Limit,
    Side:     domain.Buy,
    Price:    30000,
    Quantity: 1,
}
```

**行为**：
- 未成交部分会挂单到订单簿
- 保证成交价格不差于指定价格

### 2. Market Order（市价单）

```go
order := &Order{
    Type:     domain.Market,
    Side:     domain.Buy,
    Price:    999999,  // 极高价格确保成交
    Quantity: 1,
}
```

**行为**：
- 以当前市场最优价格立即成交
- 可能出现滑点（Slippage）

### 3. IOC Order（立即成交或取消）

```go
order := &Order{
    Type:     domain.IOC,
    Side:     domain.Sell,
    Price:    30000,
    Quantity: 10,
}
```

**行为**：
- 立即尝试成交
- 未成交部分直接取消，**不挂单**
- 用于强制平仓等场景

**代码实现**：
```go
// IOC 订单不挂单
if order.Quantity > 0 && order.Type != domain.IOC {
    heap.Push(orderBook, order)  // 普通订单挂单
}
// IOC 订单的未成交部分自动丢弃
```

---

## 实战案例分析

### 案例 1：完全成交

```
订单簿状态：
Asks: [30000: 1 BTC, 30100: 2 BTC]
Bids: [29900: 1 BTC]

新订单：BUY 1 BTC @ 30000

撮合过程：
1. 取出最低卖单：30000, 1 BTC
2. 价格匹配：30000 >= 30000 ✓
3. 成交数量：min(1, 1) = 1 BTC
4. 生成 2 条 Trade
5. 订单完全成交，结束

最终状态：
Asks: [30100: 2 BTC]
Bids: [29900: 1 BTC]
Trades: 2 条
```

### 案例 2：部分成交后挂单

```
订单簿状态：
Asks: [30000: 0.5 BTC, 30100: 1 BTC]
Bids: []

新订单：BUY 2 BTC @ 30050

撮合过程：
1. 吃掉 30000 的 0.5 BTC → 剩余 1.5 BTC
2. 吃掉 30100 的 1 BTC   → 剩余 0.5 BTC
3. 无更多可匹配订单
4. 剩余 0.5 BTC 挂单到 Bids

最终状态：
Asks: []
Bids: [30050: 0.5 BTC]
Trades: 4 条 (2次成交 × 2条记录)
```

### 案例 3：价格-时间优先

```
订单簿状态：
Asks: [
    30000: 1 BTC (10:00:00) ← 相同价格，时间早
    30000: 1 BTC (10:00:05) ← 相同价格，时间晚
    30100: 1 BTC
]

新订单：BUY 1 BTC @ 30000

撮合结果：
成交对手：30000, 10:00:00 的订单
原因：价格相同时，时间优先
```

---

## 性能分析

### 时间复杂度

| 操作 | 复杂度 | 说明 |
|------|--------|------|
| 订单插入 | O(log N) | 堆插入操作 |
| 取最优价 | O(log N) | 堆弹出操作 |
| 完全成交 | O(M log N) | M 为成交次数 |
| 查找订单簿 | O(1) | HashMap 查找 |

### 空间复杂度

- 订单簿：O(N)，N 为挂单总数
- 每个交易对独立维护

### 并发性能

**当前实现（全局锁）**：
```go
func (m *MatchingEngine) SubmitOrder(order *domain.Order) []*domain.Trade {
    m.mu.Lock()         // 全局锁
    defer m.mu.Unlock()
    // ...
}
```

**限制**：
- 所有交易对共享一个锁
- 并发性能受限

**改进方案**：使用 Symbol Sharding（见 [symbol-sharding-architecture.md](symbol-sharding-architecture.md)）

---

## 线程安全

### 锁策略

```go
type MatchingEngine struct {
    mu    sync.Mutex              // 全局互斥锁
    books map[string]*OrderBook
}

func (m *MatchingEngine) SubmitOrder(order *domain.Order) []*domain.Trade {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    book := m.getBook(order.Symbol)
    return book.Match(order)
}
```

**保护范围**：
- 订单簿的创建（`getBook`）
- 订单的提交和撮合

**优点**：
- 实现简单
- 确保线程安全

**缺点**：
- 全局锁限制并发
- 不同交易对也会互相阻塞

---

## 与生产系统的差异

### 当前实现（简化版）

```go
// 单一堆结构
type PriceHeap struct {
    orders []*domain.Order
}

// 全局锁
mu sync.Mutex
```

### 生产级实现（如 Binance）

```cpp
// 多级价格索引
std::map<Price, OrderList> price_levels;

// 每个交易对独立锁（或无锁）
per_symbol_lock

// 内存池优化
custom_allocator
```

**关键差异**：
1. **价格级别**：生产系统通常按价格级别（Price Level）组织订单
2. **内存管理**：使用内存池避免频繁分配
3. **并发模型**：无锁队列或 per-symbol 锁
4. **事件通知**：发布订单簿快照和增量更新

---

## 扩展阅读

### 相关文件

- [`matching_engine.go`](file:///Users/benjamin/Library/Mobile%20Documents/com~apple~CloudDocs/Documents/github/go-project/oms-contract/internal/engine/matching_engine.go) - 完整实现代码
- [`matching_engine_sharded.go`](file:///Users/benjamin/Library/Mobile%20Documents/com~apple~CloudDocs/Documents/github/go-project/oms-contract/internal/engine/matching_engine_sharded.go) - 分片版本
- [Symbol Sharding 架构](symbol-sharding-architecture.md) - 高并发方案

### 测试用例

- [`engine_test/`](file:///Users/benjamin/Library/Mobile%20Documents/com~apple~CloudDocs/Documents/github/go-project/oms-contract/internal/engine_test) - 单元测试和集成测试

---

## 总结

Atlas OMS 订单撮合引擎的核心特性：

✅ **价格-时间优先**：严格遵循市场公平原则  
✅ **堆优化**：O(log N) 高效性能  
✅ **多订单类型**：支持 Limit、Market、IOC  
✅ **线程安全**：互斥锁保护  
✅ **双向成交**：为 Taker 和 Maker 都生成 Trade 记录  

这种设计在简洁性和性能之间取得了良好平衡，适合作为学习和快速原型开发的基础。对于超高频交易场景，建议使用 Sharded 版本以提升并发能力。
