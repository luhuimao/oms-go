package snapshot

import (
	"encoding/json"
	"oms-contract/internal/domain"
	"oms-contract/internal/memory"
)

// SystemState represents the complete state of the OMS system
type SystemState struct {
	OrderBook    *memory.OrderBook    `json:"-"`
	PositionBook *memory.PositionBook `json:"-"`
	LastEventID  int64                `json:"last_event_id"`
	Timestamp    int64                `json:"timestamp"` // Unix timestamp
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
	// case EventOrderFilled: // To be implemented if specialized logic needed
	// case EventOrderCanceled: // To be implemented
	case EventTradeExecuted:
		return ss.applyTradeExecuted(event)
	case EventPositionOpened, EventPositionUpdated, EventPositionClosed:
		return ss.applyPositionUpdated(event)
	case EventLiquidation:
		return ss.applyLiquidation(event)
	default:
		// Unknown or unhandled event type for state reconstruction, skip
		return nil
	}
}

// applyOrderCreated applies an ORDER_CREATED event
func (ss *SystemState) applyOrderCreated(event *Event) error {
	var data OrderCreatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	if data.Order != nil {
		ss.OrderBook.Add(data.Order)
	}
	return nil
}

// applyTradeExecuted applies a TRADE_EXECUTED event
func (ss *SystemState) applyTradeExecuted(event *Event) error {
	var data TradeExecutedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	// For trade execution, we might need to update the order status
	// But simply adding the order again (if it contains updated state) works for snapshot
	// However, usually trade event contains Trade info, not full Order info.
	// If EventTradeExecuted only has Trade, we need to know how it affects Order/Position
	// verification depends on what's in the event.
	// For this implementation, we assume the event might carry updated order state or we just log it.
	// The crucial part for state is the OrderBook and PositionBook.
	// If the event sourcing model relies on re-calculating state from inputs, that's one way.
	// If events carry the *result* state (e.g. "Order updated to Filled"), we just apply that.

	// Let's assume for this simple contract, we rely on the fact that standard flow
	// produces Order/Position updates which should be captured.
	// If TradeExecuted is just an output, maybe it doesn't change state directly unless it implies order update.
	// We'll leave it empty for now unless we change Event definitions to carry resulting state.
	return nil
}

// applyPositionUpdated applies a POSITION_UPDATED/OPENED/CLOSED event
func (ss *SystemState) applyPositionUpdated(event *Event) error {
	var data PositionUpdatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return err
	}

	if data.Position != nil {
		ss.PositionBook.Save(data.Position)
	}
	return nil
}

// applyLiquidation applies a LIQUIDATION event
func (ss *SystemState) applyLiquidation(event *Event) error {
	// Liquidation might trigger position updates, which should be covered by PositionUpdated events
	// if they are emitted sequentially.
	// Use this handler if Liquidation event carries state change info itself.
	return nil
}

// Clone creates a deep copy of the system state
func (ss *SystemState) Clone() *SystemState {
	newState := NewSystemState()
	newState.LastEventID = ss.LastEventID
	newState.Timestamp = ss.Timestamp

	// Deep copy orders
	for _, o := range ss.OrderBook.GetAll() {
		// Manual deep copy of order if needed, but Order struct is simple enough for now
		// Assuming Order is immutable once created or pointer is not shared dangerously
		orderCopy := *o
		newState.OrderBook.Add(&orderCopy)
	}

	// Deep copy positions
	for _, p := range ss.PositionBook.GetAll() {
		posCopy := *p
		newState.PositionBook.Save(&posCopy)
	}

	return newState
}

// Checksum calculates a checksum of the entire system state
func (ss *SystemState) Checksum() (string, error) {
	// We need deterministic ordering for checksum
	// So we pull all data and maybe sort it or use a method that handles simple structs
	// Since maps are unordered, simply marshalling the whole state might be flaky
	// unless we sort keys.

	orders := ss.OrderBook.GetAll()
	positions := ss.PositionBook.GetAll()

	stateData := struct {
		LastEventID int64                       `json:"last_event_id"`
		Timestamp   int64                       `json:"timestamp"`
		Orders      map[int64]*domain.Order     `json:"orders"`
		Positions   map[string]*domain.Position `json:"positions"`
	}{
		LastEventID: ss.LastEventID,
		Timestamp:   ss.Timestamp,
		Orders:      orders,
		Positions:   positions,
	}

	return CalculateChecksum(stateData)
}

// ToSnapshot converts system state to a snapshot
func (ss *SystemState) ToSnapshot() *Snapshot {
	checksum, _ := ss.Checksum()

	return &Snapshot{
		SequenceID: ss.LastEventID,
		Timestamp:  ss.Timestamp,
		Orders:     ss.OrderBook.GetAll(),
		Positions:  ss.PositionBook.GetAll(),
		Checksum:   checksum,
	}
}
