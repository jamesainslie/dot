package executor

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jamesainslie/dot/pkg/dot"
)

// CheckpointID uniquely identifies a checkpoint.
type CheckpointID string

// Checkpoint records executed operations for rollback.
type Checkpoint struct {
	ID         CheckpointID
	CreatedAt  time.Time
	Operations map[dot.OperationID]dot.Operation
	mu         sync.RWMutex
}

// Record stores an executed operation in the checkpoint.
func (c *Checkpoint) Record(id dot.OperationID, op dot.Operation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Operations[id] = op
}

// Lookup retrieves an operation from the checkpoint.
func (c *Checkpoint) Lookup(id dot.OperationID) dot.Operation {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Operations[id]
}

// CheckpointStore manages checkpoint persistence.
type CheckpointStore interface {
	Create(ctx context.Context) *Checkpoint
	Delete(ctx context.Context, id CheckpointID) error
	Restore(ctx context.Context, id CheckpointID) (*Checkpoint, error)
}

// MemoryCheckpointStore keeps checkpoints in memory.
// Suitable for testing and simple cases where persistence is not required.
type MemoryCheckpointStore struct {
	checkpoints map[CheckpointID]*Checkpoint
}

// NewMemoryCheckpointStore creates a new in-memory checkpoint store.
func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{
		checkpoints: make(map[CheckpointID]*Checkpoint),
	}
}

func (s *MemoryCheckpointStore) Create(ctx context.Context) *Checkpoint {
	id := CheckpointID(uuid.New().String())
	checkpoint := &Checkpoint{
		ID:         id,
		CreatedAt:  time.Now(),
		Operations: make(map[dot.OperationID]dot.Operation),
	}
	s.checkpoints[id] = checkpoint
	return checkpoint
}

func (s *MemoryCheckpointStore) Delete(ctx context.Context, id CheckpointID) error {
	delete(s.checkpoints, id)
	return nil
}

func (s *MemoryCheckpointStore) Restore(ctx context.Context, id CheckpointID) (*Checkpoint, error) {
	checkpoint, exists := s.checkpoints[id]
	if !exists {
		return nil, dot.ErrCheckpointNotFound{ID: string(id)}
	}
	return checkpoint, nil
}
