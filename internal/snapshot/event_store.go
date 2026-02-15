package snapshot

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// EventStore manages the append-only event log (Write-Ahead Log)
type EventStore struct {
	filename string
	file     *os.File
	writer   *bufio.Writer
	mu       sync.Mutex
	sequence int64
	closed   bool
}

// NewEventStore creates a new event store
func NewEventStore(dir string) (*EventStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create event directory: %w", err)
	}

	filename := filepath.Join(dir, "events.log")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open event log: %w", err)
	}

	// Read existing events to determine last sequence number
	sequence, err := readLastSequence(filename)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to read last sequence: %w", err)
	}

	es := &EventStore{
		filename: filename,
		file:     file,
		writer:   bufio.NewWriter(file),
		sequence: sequence,
		closed:   false,
	}

	return es, nil
}

// Append adds a new event to the log
func (es *EventStore) Append(event *Event) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.closed {
		return fmt.Errorf("event store is closed")
	}

	// Assign sequence ID
	event.ID = atomic.AddInt64(&es.sequence, 1)

	// Recalculate checksum with new ID
	event.Checksum = event.calculateChecksum()

	// Serialize event
	data, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Write to log (one event per line)
	if _, err := es.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}
	if _, err := es.writer.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	// Flush to disk
	if err := es.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush: %w", err)
	}

	// Force sync to disk (durability)
	if err := es.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	return nil
}

// ReadAll reads all events from the log
func (es *EventStore) ReadAll() ([]*Event, error) {
	return es.ReadFrom(0)
}

// ReadFrom reads events starting from a specific sequence ID
func (es *EventStore) ReadFrom(sequenceID int64) ([]*Event, error) {
	file, err := os.Open(es.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Event{}, nil
		}
		return nil, fmt.Errorf("failed to open event log: %w", err)
	}
	defer file.Close()

	var events []*Event
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		event, err := UnmarshalEvent(line)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %w", err)
		}

		// Verify checksum
		if !event.Verify() {
			return nil, fmt.Errorf("event %d failed checksum verification", event.ID)
		}

		// Filter by sequence ID
		if event.ID > sequenceID {
			events = append(events, event)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read event log: %w", err)
	}

	return events, nil
}

// LastSequenceID returns the last sequence ID in the store
func (es *EventStore) LastSequenceID() int64 {
	return atomic.LoadInt64(&es.sequence)
}

// Close closes the event store
func (es *EventStore) Close() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.closed {
		return nil
	}

	es.closed = true

	if err := es.writer.Flush(); err != nil {
		return err
	}

	return es.file.Close()
}

// readLastSequence reads the last sequence number from the log
func readLastSequence(filename string) (int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer file.Close()

	var lastSeq int64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(line, &event); err != nil {
			continue
		}

		if event.ID > lastSeq {
			lastSeq = event.ID
		}
	}

	return lastSeq, scanner.Err()
}
