package snapshot_test

import (
	"io/ioutil"
	"os"
	"testing"

	"oms-contract/internal/domain"
	"oms-contract/internal/snapshot"
)

func TestSnapshotAndReplay(t *testing.T) {
	// Setup temporary directory
	tmpDir, err := ioutil.TempDir("", "snapshot_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize components
	eventStore, err := snapshot.NewEventStore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	snapshotManager, err := snapshot.NewSnapshotManager(tmpDir, 5)
	if err != nil {
		t.Fatal(err)
	}
	systemState := snapshot.NewSystemState()

	// 1. Generate some events
	order1 := &domain.Order{ID: 1, UserID: 101, Symbol: "BTCUSDT", Price: 50000, Quantity: 1.0, Type: domain.Limit, Side: domain.Buy}
	event1 := snapshot.NewEvent(1, snapshot.EventOrderCreated, snapshot.OrderCreatedData{Order: order1})

	order2 := &domain.Order{ID: 2, UserID: 102, Symbol: "ETHUSDT", Price: 3000, Quantity: 10.0, Type: domain.Limit, Side: domain.Sell}
	event2 := snapshot.NewEvent(2, snapshot.EventOrderCreated, snapshot.OrderCreatedData{Order: order2})

	// 2. Apply events and append to store
	if err := systemState.ApplyEvent(event1); err != nil {
		t.Fatalf("Failed to apply event 1: %v", err)
	}
	if err := eventStore.Append(event1); err != nil {
		t.Fatalf("Failed to append event 1: %v", err)
	}

	if err := systemState.ApplyEvent(event2); err != nil {
		t.Fatalf("Failed to apply event 2: %v", err)
	}
	if err := eventStore.Append(event2); err != nil {
		t.Fatalf("Failed to append event 2: %v", err)
	}

	// 3. Take snapshot
	if err := snapshotManager.TakeSnapshot(systemState); err != nil {
		t.Fatalf("Failed to take snapshot: %v", err)
	}

	// 4. Create more events after snapshot
	order3 := &domain.Order{ID: 3, UserID: 103, Symbol: "BTCUSDT", Price: 51000, Quantity: 0.5, Type: domain.Limit, Side: domain.Sell}
	event3 := snapshot.NewEvent(3, snapshot.EventOrderCreated, snapshot.OrderCreatedData{Order: order3})

	if err := systemState.ApplyEvent(event3); err != nil {
		t.Fatalf("Failed to apply event 3: %v", err)
	}
	if err := eventStore.Append(event3); err != nil {
		t.Fatalf("Failed to append event 3: %v", err)
	}

	// 5. Initialize ReplayEngine
	replayEngine := snapshot.NewReplayEngine(eventStore, snapshotManager)

	// 6. Replay and verify state
	replayedState, err := replayEngine.Replay()
	if err != nil {
		t.Fatalf("Replay failed: %v", err)
	}

	// 7. Compare states
	originalChecksum, _ := systemState.Checksum()
	replayedChecksum, _ := replayedState.Checksum()

	if originalChecksum != replayedChecksum {
		t.Errorf("Checksum mismatch! Original: %s, Replayed: %s", originalChecksum, replayedChecksum)
	}

	// Verify specific data
	if len(replayedState.OrderBook.GetAll()) != 3 {
		t.Errorf("Expected 3 orders, got %d", len(replayedState.OrderBook.GetAll()))
	}

	// 8. Test Verify integrity check
	if err := replayEngine.Verify(); err != nil {
		t.Errorf("Verify failed: %v", err)
	}
}
