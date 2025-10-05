package dot

import (
	"context"
	"sync"
	"time"
)

// CheckpointID uniquely identifies a checkpoint.
type CheckpointID string

// Checkpoint stores the state needed for rollback.
type Checkpoint struct {
	ID         CheckpointID
	CreatedAt  time.Time
	Operations map[OperationID]Operation
	mu         sync.RWMutex
}

// Record adds an executed operation to the checkpoint.
func (c *Checkpoint) Record(id OperationID, op Operation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Operations == nil {
		c.Operations = make(map[OperationID]Operation)
	}
	c.Operations[id] = op
}

// Lookup retrieves an operation from the checkpoint.
func (c *Checkpoint) Lookup(id OperationID) Operation {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Operations[id]
}

// CheckpointStore manages checkpoints for rollback.
type CheckpointStore interface {
	Create(ctx context.Context) *Checkpoint
	Delete(ctx context.Context, id CheckpointID) error
	Restore(ctx context.Context, id CheckpointID) (*Checkpoint, error)
}
