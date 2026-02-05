# Atlas OMS – Production-Grade Order Management System

## Overview

**Atlas OMS** is a high-performance, production-ready Order Management System (OMS) designed for centralized derivatives exchanges and professional trading platforms. Inspired by the internal architectures of **Binance**, **OKX**, and **Bybit**, Atlas OMS provides the core infrastructure to handle order lifecycle, risk management, position tracking, and liquidation in a modular, scalable manner.

This is **not a demo**; it is designed with production reliability and extensibility in mind.

---

## Core Features

### Order Management

* Full order lifecycle (NEW, PARTIALLY_FILLED, FILLED, CANCELED)
* Limit and Market orders
* Idempotent order processing
* Integration with matching engine via events

### Position and Margin Engine

* One-way position mode
* Isolated margin support
* Configurable leverage per user
* Real-time PnL calculation and equity tracking

### Liquidation Engine

* Mark-price-driven liquidation checks
* Maintenance margin enforcement
* IOC liquidation orders for seamless market execution
* Hooks for Insurance Fund and ADL (Automatic Deleveraging)

### Risk Management

* Margin and leverage validation
* Maintenance margin monitoring
* Dynamic risk limits (planned for future versions)

### System Design

* Event-driven architecture
* Memory-first with persistence backend optional
* Deterministic and replayable state transitions
* Hash-based worker dispatch for per-order serialization

---

## IOC Liquidation → Matching Engine Closed Loop

Atlas OMS implements a **production-grade liquidation flow**:

1. **Liquidation Trigger**: When a position’s maintenance margin is breached.
2. **Position Freeze**: Prevent further user actions.
3. **IOC Liquidation Order Creation**: Generate system market order with Immediate-Or-Cancel TIF.
4. **Send to Matching Engine**: Matching engine executes against order book.
5. **Trade Event Backflow**: Trades are returned to OMS to update position, margin, and realized PnL.
6. **Post-Liquidation Handling**: Remaining positions are processed via Insurance Fund or ADL if necessary.

This design ensures:

* Strong consistency
* Market-driven price discovery
* Auditability of liquidation events

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
│ Risk Engine   │
└───────┬───────┘
        │ Events (Kafka / NATS / MQ)
        ▼
  Matching Engine
```

### Design Principles

* Event-driven and decoupled
* Deterministic execution
* Horizontal scalability
* Replayable state for fault tolerance

---

## Repo Structure

```
oms-project/
├── cmd/oms/                # Service entrypoint
├── internal/
│   ├── domain/             # Models and event definitions
│   ├── service/            # Core business logic (OMS, Position, Liquidation, Risk)
│   ├── engine/             # Sequential execution engine / dispatcher
│   ├── memory/             # In-memory state stores
│   └── infra/              # Integration / mock matching engine
├── pkg/                    # Utilities (ID generator, helpers)
├── go.mod
└── README.md
```

---

## Getting Started

### Prerequisites

* Go 1.21+
* Linux/macOS

### Run

```bash
go run ./cmd/oms
```

This will start the OMS service, handle example order submission, trade execution, position updates, and demonstrate IOC liquidation.

---

## Target Use Cases

* Centralized derivatives exchanges
* Prop trading platforms
* Institutional trading gateways
* Quantitative research environments
* Exchange simulation platforms

---

## Roadmap

**Short-Term:** IOC liquidation order generation, Risk Limit tiers, Reduce-Only positions.

**Mid-Term:** Cross-margin support, Hedge mode, Insurance fund integration.

**Long-Term:** ADL engine, Portfolio margin, Multi-asset collateral.

---

## Licensing

Apache 2.0 (or commercial licensing available).

## Contact

For commercial licensing or integration support:

> 📧 [luhuimao@gmail.com](mailto:luhuimao@gmail.com) 

---

**Atlas OMS is designed to be the backbone of a real-world derivatives trading platform. It provides a scalable, deterministic, and auditable infrastructure suitable for production environments.**
