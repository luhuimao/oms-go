package snapshot

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"oms-contract/internal/domain"
)

// Snapshot represents a point-in-time snapshot of the system state
type Snapshot struct {
	SequenceID int64                       `json:"sequence_id"`
	Timestamp  int64                       `json:"timestamp"`
	Orders     map[int64]*domain.Order     `json:"orders"`
	Positions  map[string]*domain.Position `json:"positions"`
	Checksum   string                      `json:"checksum"`
}

// SnapshotInfo contains metadata about a snapshot
type SnapshotInfo struct {
	Filename   string
	SequenceID int64
	Timestamp  int64
	Size       int64
}

// SnapshotManager handles snapshot creation and loading
type SnapshotManager struct {
	dir              string
	retentionCount   int
	compressionLevel int
	mu               sync.Mutex
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(dir string, retentionCount int) (*SnapshotManager, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	return &SnapshotManager{
		dir:              dir,
		retentionCount:   retentionCount,
		compressionLevel: gzip.BestCompression,
	}, nil
}

// TakeSnapshot creates a new snapshot
func (sm *SnapshotManager) TakeSnapshot(state *SystemState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshot := state.ToSnapshot()

	// Generate filename with timestamp and sequence
	filename := fmt.Sprintf("snapshot_%d_%d.snap.gz", snapshot.Timestamp, snapshot.SequenceID)
	tmpPath := filepath.Join(sm.dir, filename+".tmp")
	finalPath := filepath.Join(sm.dir, filename)

	// Write to temporary file first (atomic write pattern)
	if err := sm.writeSnapshot(tmpPath, snapshot); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename snapshot: %w", err)
	}

	// Clean up old snapshots
	if err := sm.cleanupOldSnapshots(); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: failed to cleanup old snapshots: %v\n", err)
	}

	return nil
}

// writeSnapshot writes a snapshot to a file with compression
func (sm *SnapshotManager) writeSnapshot(path string, snapshot *Snapshot) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create gzip writer
	gzWriter, err := gzip.NewWriterLevel(file, sm.compressionLevel)
	if err != nil {
		return err
	}
	defer gzWriter.Close()

	// Encode snapshot as JSON
	encoder := json.NewEncoder(gzWriter)
	if err := encoder.Encode(snapshot); err != nil {
		return err
	}

	// Flush and sync
	if err := gzWriter.Close(); err != nil {
		return err
	}

	return file.Sync()
}

// LoadLatest loads the most recent snapshot
func (sm *SnapshotManager) LoadLatest() (*Snapshot, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshots, err := sm.listSnapshots()
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots found")
	}

	// Sort by sequence ID descending
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].SequenceID > snapshots[j].SequenceID
	})

	// Load the latest
	latestFile := snapshots[0].Filename
	return sm.loadSnapshot(latestFile)
}

// LoadBySequence loads a specific snapshot by sequence ID
func (sm *SnapshotManager) LoadBySequence(sequenceID int64) (*Snapshot, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshots, err := sm.listSnapshots()
	if err != nil {
		return nil, err
	}

	// Find snapshot with matching sequence ID
	for _, info := range snapshots {
		if info.SequenceID == sequenceID {
			return sm.loadSnapshot(info.Filename)
		}
	}

	return nil, fmt.Errorf("snapshot with sequence %d not found", sequenceID)
}

// loadSnapshot loads a snapshot from a file
func (sm *SnapshotManager) loadSnapshot(filename string) (*Snapshot, error) {
	path := filepath.Join(sm.dir, filename)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	// Decode snapshot
	var snapshot Snapshot
	decoder := json.NewDecoder(gzReader)
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, err
	}

	// Verify checksum
	// In production, recalculate checksum and compare

	return &snapshot, nil
}

// List returns information about all snapshots
func (sm *SnapshotManager) List() ([]SnapshotInfo, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.listSnapshots()
}

// listSnapshots lists all snapshot files
func (sm *SnapshotManager) listSnapshots() ([]SnapshotInfo, error) {
	files, err := ioutil.ReadDir(sm.dir)
	if err != nil {
		return nil, err
	}

	var snapshots []SnapshotInfo
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".snap.gz") {
			continue
		}

		// Parse filename: snapshot_<timestamp>_<sequence>.snap.gz
		var timestamp, sequenceID int64
		_, err := fmt.Sscanf(file.Name(), "snapshot_%d_%d.snap.gz", &timestamp, &sequenceID)
		if err != nil {
			continue
		}

		snapshots = append(snapshots, SnapshotInfo{
			Filename:   file.Name(),
			SequenceID: sequenceID,
			Timestamp:  timestamp,
			Size:       file.Size(),
		})
	}

	return snapshots, nil
}

// cleanupOldSnapshots removes old snapshots beyond retention count
func (sm *SnapshotManager) cleanupOldSnapshots() error {
	snapshots, err := sm.listSnapshots()
	if err != nil {
		return err
	}

	if len(snapshots) <= sm.retentionCount {
		return nil
	}

	// Sort by sequence ID descending
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].SequenceID > snapshots[j].SequenceID
	})

	// Delete old snapshots
	for i := sm.retentionCount; i < len(snapshots); i++ {
		path := filepath.Join(sm.dir, snapshots[i].Filename)
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	return nil
}

// TakeSnapshotPeriodic creates snapshots periodically
func (sm *SnapshotManager) TakeSnapshotPeriodic(state *SystemState, interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sm.TakeSnapshot(state); err != nil {
				fmt.Printf("Error taking periodic snapshot: %v\n", err)
			} else {
				fmt.Printf("Snapshot created at sequence %d\n", state.LastEventID)
			}
		case <-done:
			return
		}
	}
}
