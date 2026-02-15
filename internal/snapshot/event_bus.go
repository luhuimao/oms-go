package snapshot

import (
	"sync"
)

// EventBus handles event publishing
type EventBus struct {
	store *EventStore
	state *SystemState
	mu    sync.Mutex
}

// NewEventBus creates a new event bus
func NewEventBus(store *EventStore, state *SystemState) *EventBus {
	return &EventBus{
		store: store,
		state: state,
	}
}

// Publish publishes an event to the system
func (b *EventBus) Publish(event *Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. Apply to state first (in-memory update)
	// This ensures our in-memory state is always up-to-date with events
	// Note: In a real system, we might want to validate against state BEFORE creating the event
	// but here we assume the service layer has already validated the command.
	if err := b.state.ApplyEvent(event); err != nil {
		return err
	}

	// 2. Persist to store
	if err := b.store.Append(event); err != nil {
		// If persistence fails, we are in an inconsistent state!
		// In production, we would need a way to rollback state or halt the system.
		// For now, we return the error.
		return err
	}

	return nil
}
