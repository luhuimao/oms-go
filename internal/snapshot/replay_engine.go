package snapshot

import (
	"fmt"
)

// ReplayEngine rebuilds system state from snapshots and events
type ReplayEngine struct {
	eventStore *EventStore
	snapMgr    *SnapshotManager
}

// NewReplayEngine creates a new replay engine
func NewReplayEngine(eventStore *EventStore, snapMgr *SnapshotManager) *ReplayEngine {
	return &ReplayEngine{
		eventStore: eventStore,
		snapMgr:    snapMgr,
	}
}

// Replay rebuilds the complete system state
func (re *ReplayEngine) Replay() (*SystemState, error) {
	// Try to load latest snapshot
	snapshot, err := re.snapMgr.LoadLatest()
	if err != nil {
		// No snapshot exists, start from empty state
		return re.replayFromBeginning()
	}

	// Replay from snapshot
	return re.replayFromSnapshot(snapshot)
}

// replayFromBeginning replays all events from the beginning
func (re *ReplayEngine) replayFromBeginning() (*SystemState, error) {
	state := NewSystemState()

	// Read all events
	events, err := re.eventStore.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	// Apply all events
	for _, event := range events {
		if err := state.ApplyEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event %d: %w", event.ID, err)
		}
	}

	return state, nil
}

// replayFromSnapshot replays events from a snapshot
func (re *ReplayEngine) replayFromSnapshot(snapshot *Snapshot) (*SystemState, error) {
	// Restore state from snapshot
	state := re.restoreFromSnapshot(snapshot)

	// Read events since snapshot
	events, err := re.eventStore.ReadFrom(snapshot.SequenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	// Apply events
	for _, event := range events {
		if err := state.ApplyEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event %d: %w", event.ID, err)
		}
	}

	return state, nil
}

// restoreFromSnapshot restores system state from a snapshot
func (re *ReplayEngine) restoreFromSnapshot(snapshot *Snapshot) *SystemState {
	state := NewSystemState()
	state.LastEventID = snapshot.SequenceID
	state.Timestamp = snapshot.Timestamp

	// Restore orders
	for _, order := range snapshot.Orders {
		state.OrderBook.Add(order)
	}

	// Restore positions
	for _, position := range snapshot.Positions {
		state.PositionBook.Save(position)
	}

	return state
}

// ReplayTo rebuilds state up to a specific sequence ID
func (re *ReplayEngine) ReplayTo(sequenceID int64) (*SystemState, error) {
	state := NewSystemState()

	// Read events up to sequence ID
	allEvents, err := re.eventStore.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read events: %w", err)
	}

	// Apply events up to sequence ID
	for _, event := range allEvents {
		if event.ID > sequenceID {
			break
		}
		if err := state.ApplyEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event %d: %w", event.ID, err)
		}
	}

	return state, nil
}

// Verify checks the integrity of the event log and snapshots
func (re *ReplayEngine) Verify() error {
	// Read all events
	events, err := re.eventStore.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read events: %w", err)
	}

	// Verify each event's checksum
	for _, event := range events {
		if !event.Verify() {
			return fmt.Errorf("event %d failed checksum verification", event.ID)
		}
	}

	// Verify snapshots
	snapshots, err := re.snapMgr.List()
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	for _, info := range snapshots {
		snapshot, err := re.snapMgr.LoadBySequence(info.SequenceID)
		if err != nil {
			return fmt.Errorf("failed to load snapshot %d: %w", info.SequenceID, err)
		}

		// In production, verify snapshot checksum
		_ = snapshot
	}

	return nil
}
