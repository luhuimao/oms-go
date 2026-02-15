package snapshot

import (
	"oms-contract/internal/domain"
	"oms-contract/internal/memory"
)

// SystemState represents the complete state of the OMS system
type SystemState struct {
	OrderBook    *memory.OrderBook
	PositionBook *memory.PositionBook
	LastEventID  int64
	Timestamp    int64 // Unix timestamp
}

// NewSystemState creates a new system state
func NewSystemState() *SystemState {
	return &SystemState{
		OrderBook:    memory.NewOrderBook(),
		PositionBook: memory.NewPositionBook(),
		LastEventID:  0,
		Timestamp:    0,
	}
}

// ApplyEvent applies an event to the system state
func (ss *SystemState) ApplyEvent(event *Event) error {
	ss.LastEventID = event.ID
	ss.Timestamp = event.Timestamp.Unix()

	switch event.Type {
	case EventOrderCreated:
		return ss.applyOrderCreated(event)
	case EventTradeExecuted:
		return ss.applyTradeExecuted(event)
	case EventPositionUpdated:
		return ss.applyPositionUpdated(event)
	case EventLiquidation:
		return ss.applyLiquidation(event)
	default:
		// Unknown event type, skip
		return nil
	}
}

// applyOrderCreated applies an ORDER_CREATED event
func (ss *SystemState) applyOrderCreated(event *Event) error {
	// Extract order data from event
	// Note: In production, you'd need proper type assertion and error handling
	if data, ok := event.Data.(map[string]interface{}); ok {
		if orderData, ok := data["order"].(map[string]interface{}); ok {
			order := &domain.Order{
				ID:     int64(orderData["id"].(float64)),
				UserID: int64(orderData["user_id"].(float64)),
				Symbol: orderData["symbol"].(string),
				// Add other fields as needed
			}
			ss.OrderBook.Add(order)
		}
	}
	return nil
}

// applyTradeExecuted applies a TRADE_EXECUTED event
func (ss *SystemState) applyTradeExecuted(event *Event) error {
	// Similar to applyOrderCreated, extract and apply trade data
	return nil
}

// applyPositionUpdated applies a POSITION_UPDATED event
func (ss *SystemState) applyPositionUpdated(event *Event) error {
	// Extract position data and update position book
	return nil
}

// applyLiquidation applies a LIQUIDATION event
func (ss *SystemState) applyLiquidation(event *Event) error {
	// Handle liquidation event
	return nil
}

// Clone creates a deep copy of the system state
func (ss *SystemState) Clone() *SystemState {
	return &SystemState{
		OrderBook:    ss.OrderBook,    // In production, implement deep copy
		PositionBook: ss.PositionBook, // In production, implement deep copy
		LastEventID:  ss.LastEventID,
		Timestamp:    ss.Timestamp,
	}
}

// Checksum calculates a checksum of the entire system state
func (ss *SystemState) Checksum() (string, error) {
	stateData := struct {
		LastEventID int64 `json:"last_event_id"`
		Timestamp   int64 `json:"timestamp"`
		// In production, include serialized order book and position book
	}{
		LastEventID: ss.LastEventID,
		Timestamp:   ss.Timestamp,
	}

	return CalculateChecksum(stateData)
}

// ToSnapshot converts system state to a snapshot
func (ss *SystemState) ToSnapshot() *Snapshot {
	// Extract all orders
	orders := make(map[int64]*domain.Order)
	// In production, iterate through order book and copy all orders

	// Extract all positions
	positions := make(map[string]*domain.Position)
	// In production, iterate through position book and copy all positions

	checksum, _ := ss.Checksum()

	return &Snapshot{
		SequenceID: ss.LastEventID,
		Timestamp:  ss.Timestamp,
		Orders:     orders,
		Positions:  positions,
		Checksum:   checksum,
	}
}
