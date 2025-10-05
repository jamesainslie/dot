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
	operations map[dot.OperationID]dot.Operation
	mu         sync.RWMutex
}

// Record stores an executed operation in the checkpoint.
func (c *Checkpoint) Record(id dot.OperationID, op dot.Operation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.operations == nil {
		c.operations = make(map[dot.OperationID]dot.Operation)
	}
	c.operations[id] = op
}

// Lookup retrieves an operation from the checkpoint.
// Returns nil if the operation is not found.
func (c *Checkpoint) Lookup(id dot.OperationID) dot.Operation {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.operations[id]
}

// GetOperation retrieves an operation by ID with thread safety.
// Returns the operation and true if found, or nil and false if not found.
func (c *Checkpoint) GetOperation(id dot.OperationID) (dot.Operation, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	op, exists := c.operations[id]
	return op, exists
}

// ListOperations returns a snapshot of all operations in the checkpoint.
// The returned slice is a copy and safe to use concurrently.
func (c *Checkpoint) ListOperations() []dot.Operation {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ops := make([]dot.Operation, 0, len(c.operations))
	for _, op := range c.operations {
		ops = append(ops, op)
	}
	return ops
}

// Len returns the number of operations in the checkpoint.
func (c *Checkpoint) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.operations)
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
	mu          sync.RWMutex
}

// NewMemoryCheckpointStore creates a new in-memory checkpoint store.
func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{
		checkpoints: make(map[CheckpointID]*Checkpoint),
	}
}

func (s *MemoryCheckpointStore) Create(ctx context.Context) *Checkpoint {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := CheckpointID(uuid.New().String())
	checkpoint := &Checkpoint{
		ID:        id,
		CreatedAt: time.Now(),
		// operations map lazily initialized in Record()
	}
	s.checkpoints[id] = checkpoint
	return checkpoint
}

func (s *MemoryCheckpointStore) Delete(ctx context.Context, id CheckpointID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.checkpoints, id)
	return nil
}

func (s *MemoryCheckpointStore) Restore(ctx context.Context, id CheckpointID) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	checkpoint, exists := s.checkpoints[id]
	if !exists {
		return nil, dot.ErrCheckpointNotFound{ID: string(id)}
	}
	return checkpoint, nil
}
