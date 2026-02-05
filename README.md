

# 🚀 Atlas OMS

**Production-Grade Contract Order Management System**

> A high-performance, exchange-grade Order Management System (OMS) designed for derivatives trading platforms.
> Inspired by the internal architectures of **Binance / OKX / Bybit**, built in Go.

---

## Overview

**Atlas OMS** is a **production-ready order and position management core** for centralized derivatives exchanges and professional trading platforms.

It is designed to operate as a **Trading Gateway / OMS layer**, handling:

* Order lifecycle management
* Risk & margin validation
* Position accounting
* Leverage & liquidation logic
* Event-driven integration with a matching engine

> ❌ Atlas OMS does **not** perform price matching
> ✅ It integrates seamlessly with external matching engines via events

---

## Key Capabilities

### 🧾 Order Management

* Limit / Market orders
* Full order lifecycle state machine
* Idempotent order handling
* In-memory acceleration + durable persistence ready

### 📊 Position & Margin Engine

* One-way position mode
* Isolated margin
* Configurable leverage
* Real-time PnL & equity calculation

### ⚡ Liquidation Engine

* Mark-price driven liquidation checks
* Maintenance margin enforcement
* Deterministic liquidation triggering
* Designed to generate **IOC liquidation orders** (matching-engine friendly)

### 🔐 Strong Consistency Model

* **Single-order sequential execution**
* Hash-based worker dispatching
* Deterministic state transitions
* Replay-safe event handling

---

## Architecture

```
Client / API
     │
     ▼
┌───────────────┐
│   Atlas OMS   │
│───────────────│
│ Order Service │
│ Position Svc  │
│ Margin Engine │
│ Liquidation   │
└───────┬───────┘
        │ Events (Kafka / NATS / MQ)
        ▼
  Matching Engine
```

### Design Principles

* **Event-Driven**
* **Memory-First, Persistence-Backed**
* **Deterministic & Replayable**
* **Horizontally Scalable**

---

## Repository Structure

```
oms-demo/
├── cmd/oms/                # Service entrypoint
├── internal/
│   ├── domain/             # Pure domain models
│   ├── service/            # Core business logic
│   ├── engine/             # Sequential execution engine
│   ├── memory/             # In-memory state (orders / positions)
│
├── pkg/                    # Shared utilities (ID generator, etc.)
├── go.mod
└── README.md
```

---

## Core Components

### Order Service

Responsible for:

* Order validation & lifecycle
* Driving position updates
* Emitting order events

### Position Service

* Maintains per-user per-symbol positions
* Computes average entry price
* Tracks margin & leverage

### Liquidation Engine

* Evaluates margin safety using mark price
* Triggers forced liquidation when equity ≤ maintenance margin
* Designed for seamless IOC liquidation order creation

---

## Consistency Model

Atlas OMS guarantees:

* **Per-order strong consistency**
* No concurrent state mutation for the same order
* Deterministic execution across restarts

Implementation:

* Hash-based dispatcher
* Single goroutine per order key

This is the same execution model used by **tier-1 exchanges**.

---

## Supported Trading Model (v1)

| Feature                | Status |
| ---------------------- | ------ |
| Linear Contracts       | ✅      |
| Isolated Margin        | ✅      |
| One-Way Position       | ✅      |
| Mark Price Liquidation | ✅      |
| Cross Margin           | ⏳      |
| Hedge Mode             | ⏳      |
| ADL / Insurance Fund   | ⏳      |

---

## Getting Started

### Requirements

* Go **1.21+**
* Linux / macOS

### Run

```bash
go run ./cmd/oms
```

You will see:

* Order submission
* Trade execution
* Position update
* Liquidation check

---

## Production Readiness

Atlas OMS is designed with production deployment in mind:

* Clear service boundaries
* Dependency injection
* Stateless workers
* Ready for Kafka / Redis / MySQL integration
* Designed for multi-AZ horizontal scaling

> ⚠️ Persistence and messaging layers are intentionally abstracted
> to allow seamless integration into existing infrastructure

---

## Target Use Cases

* Centralized derivatives exchanges
* Prop trading platforms
* Institutional trading gateways
* Quant trading OMS
* Exchange simulation & research environments

---

## Roadmap

### Short-Term

* IOC liquidation order generation
* Risk limit tiers (dynamic MM)
* Position close & reduce-only logic

### Mid-Term

* Cross margin support
* Dual-side (hedge mode)
* Insurance fund integration

### Long-Term

* ADL engine
* Portfolio margin
* Multi-asset collateral

---

## Why Atlas OMS

* **Not a demo**
* **Not a toy matching engine**
* Built with **real exchange failure modes in mind**
* Designed by engineers who understand **trading system invariants**

If you are building a **real exchange**, this is the layer you cannot afford to get wrong.

---

## License

Apache 2.0 (or proprietary license available upon request)

---

## Contact

For commercial licensing, consulting, or integration support:

> 📧 [business@atlas-oms.io](mailto:business@atlas-oms.io) *(placeholder)*

---

### ⭐ Star this repo if you’re serious about trading infrastructure