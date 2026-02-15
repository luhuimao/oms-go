package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"oms-contract/internal/domain"
)

// EventType represents the type of event in the system
type EventType string

const (
	EventOrderCreated    EventType = "ORDER_CREATED"
	EventOrderFilled     EventType = "ORDER_FILLED"
	EventOrderCanceled   EventType = "ORDER_CANCELED"
	EventTradeExecuted   EventType = "TRADE_EXECUTED"
	EventPositionOpened  EventType = "POSITION_OPENED"
	EventPositionUpdated EventType = "POSITION_UPDATED"
	EventPositionClosed  EventType = "POSITION_CLOSED"
	EventLiquidation     EventType = "LIQUIDATION"
)

// Event represents a single event in the event sourcing system
type Event struct {
	ID        int64           `json:"id"`
	Type      EventType       `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
	Checksum  string          `json:"checksum"`
}

// OrderCreatedData contains data for ORDER_CREATED event
type OrderCreatedData struct {
	Order *domain.Order `json:"order"`
}

// TradeExecutedData contains data for TRADE_EXECUTED event
type TradeExecutedData struct {
	Trade *domain.Trade `json:"trade"`
}

// PositionUpdatedData contains data for POSITION_UPDATED event
type PositionUpdatedData struct {
	Position *domain.Position `json:"position"`
	Reason   string           `json:"reason"`
}

// LiquidationData contains data for LIQUIDATION event
type LiquidationData struct {
	UserID   int64   `json:"user_id"`
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Reason   string  `json:"reason"`
}

// NewEvent creates a new event with auto-generated checksum
func NewEvent(id int64, eventType EventType, data interface{}) *Event {
	dataBytes, _ := json.Marshal(data)
	event := &Event{
		ID:        id,
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}
	event.Checksum = event.calculateChecksum()
	return event
}

// calculateChecksum generates SHA256 checksum for the event
func (e *Event) calculateChecksum() string {
	// Create a copy without checksum for hashing
	temp := struct {
		ID        int64           `json:"id"`
		Type      EventType       `json:"type"`
		Timestamp time.Time       `json:"timestamp"`
		Data      json.RawMessage `json:"data"`
	}{
		ID:        e.ID,
		Type:      e.Type,
		Timestamp: e.Timestamp,
		Data:      e.Data,
	}

	data, _ := json.Marshal(temp)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Verify checks if the event's checksum is valid
func (e *Event) Verify() bool {
	expected := e.calculateChecksum()
	return e.Checksum == expected
}

// Marshal serializes the event to JSON
func (e *Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalEvent deserializes an event from JSON
func UnmarshalEvent(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
