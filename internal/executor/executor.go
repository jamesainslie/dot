// Package executor implements the imperative shell for executing plans.
// It provides transaction safety with two-phase commit and rollback capabilities.
package executor

import (
	"context"
	"fmt"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Executor executes validated plans with transaction safety.
type Executor struct {
	fs         dot.FS
	log        dot.Logger
	tracer     dot.Tracer
	checkpoint CheckpointStore
}

// Opts configures executor creation.
type Opts struct {
	FS         dot.FS
	Logger     dot.Logger
	Tracer     dot.Tracer
	Metrics    dot.Metrics
	Checkpoint CheckpointStore
}

// New creates a new Executor with the given options.
// If no checkpoint store is provided, a memory-based store is used.
func New(opts Opts) *Executor {
	if opts.Checkpoint == nil {
		opts.Checkpoint = NewMemoryCheckpointStore()
	}

	return &Executor{
		fs:         opts.FS,
		log:        opts.Logger,
		tracer:     opts.Tracer,
		checkpoint: opts.Checkpoint,
	}
}

// Execute executes a plan with two-phase commit and automatic rollback on failure.
func (e *Executor) Execute(ctx context.Context, plan dot.Plan) dot.Result[ExecutionResult] {
	ctx, span := e.tracer.Start(ctx, "executor.Execute")
	defer span.End()

	// Validate plan is not empty
	if len(plan.Operations) == 0 {
		err := dot.ErrEmptyPlan{}
		e.log.Error(ctx, "empty_plan")
		span.RecordError(err)
		return dot.Err[ExecutionResult](err)
	}

	e.log.Info(ctx, "executing_plan",
		"operation_count", len(plan.Operations))

	// Phase 1: Prepare - validate all operations
	if err := e.prepare(ctx, plan); err != nil {
		e.log.Error(ctx, "prepare_failed", "error", err)
		span.RecordError(err)
		return dot.Err[ExecutionResult](err)
	}

	// Create checkpoint before execution
	checkpoint := e.checkpoint.Create(ctx)
	e.log.Info(ctx, "checkpoint_created", "checkpoint_id", checkpoint.ID)

	// Phase 2: Commit - execute operations
	var result ExecutionResult
	if plan.CanParallelize() {
		result = e.executeParallel(ctx, plan, checkpoint)
	} else {
		result = e.executeSequential(ctx, plan, checkpoint)
	}

	if len(result.Failed) > 0 {
		// Automatic rollback
		e.log.Warn(ctx, "execution_failed_rolling_back", "failed_count", len(result.Failed))
		rolledBack := e.rollback(ctx, result.Executed, checkpoint)
		result.RolledBack = rolledBack

		err := dot.ErrExecutionFailed{
			Executed:   len(result.Executed),
			Failed:     len(result.Failed),
			RolledBack: len(result.RolledBack),
			Errors:     result.Errors,
		}
		return dot.Err[ExecutionResult](err)
	}

	// Success - delete checkpoint
	e.checkpoint.Delete(ctx, checkpoint.ID)
	e.log.Info(ctx, "execution_complete", "operations", len(result.Executed))

	return dot.Ok(result)
}

// prepare validates all operations and checks preconditions.
func (e *Executor) prepare(ctx context.Context, plan dot.Plan) error {
	ctx, span := e.tracer.Start(ctx, "executor.Prepare")
	defer span.End()

	e.log.Debug(ctx, "preparing_plan", "operations", len(plan.Operations))

	for _, op := range plan.Operations {
		if err := op.Validate(); err != nil {
			return fmt.Errorf("validation failed for %v: %w", op.ID(), err)
		}

		if err := e.checkPreconditions(ctx, op); err != nil {
			return fmt.Errorf("precondition check failed for %v: %w", op.ID(), err)
		}
	}

	e.log.Debug(ctx, "prepare_complete")
	return nil
}

// checkPreconditions verifies operation preconditions before execution.
func (e *Executor) checkPreconditions(ctx context.Context, op dot.Operation) error {
	switch operation := op.(type) {
	case dot.LinkCreate:
		return e.checkLinkCreatePreconditions(ctx, operation)
	case dot.DirCreate:
		return e.checkDirCreatePreconditions(ctx, operation)
	case dot.FileMove:
		return e.checkFileMovePreconditions(ctx, operation)
	default:
		return nil
	}
}

func (e *Executor) checkLinkCreatePreconditions(ctx context.Context, op dot.LinkCreate) error {
	// Verify source exists
	if !e.fs.Exists(ctx, op.Source.String()) {
		return dot.ErrSourceNotFound{Path: op.Source.String()}
	}

	// Verify target parent directory exists
	parent := op.Target.Parent()
	if !parent.IsOk() {
		return parent.UnwrapErr()
	}
	parentPath := parent.Unwrap()

	if !e.fs.Exists(ctx, parentPath.String()) {
		return dot.ErrParentNotFound{Path: parentPath.String()}
	}

	// Check write permission on parent (simplified check)
	info, err := e.fs.Stat(ctx, parentPath.String())
	if err != nil {
		return err
	}

	if info.Mode().Perm()&0200 == 0 {
		return dot.ErrPermissionDenied{
			Path:      parentPath.String(),
			Operation: "write",
		}
	}

	return nil
}

func (e *Executor) checkDirCreatePreconditions(ctx context.Context, op dot.DirCreate) error {
	// Check parent directory exists
	parent := op.Path.Parent()
	if !parent.IsOk() {
		// Root directory or no parent
		return nil
	}
	parentPath := parent.Unwrap()

	if !e.fs.Exists(ctx, parentPath.String()) {
		return dot.ErrParentNotFound{Path: parentPath.String()}
	}

	return nil
}

func (e *Executor) checkFileMovePreconditions(ctx context.Context, op dot.FileMove) error {
	// Verify source exists
	if !e.fs.Exists(ctx, op.Source.String()) {
		return dot.ErrSourceNotFound{Path: op.Source.String()}
	}

	// Verify destination parent exists
	parent := op.Dest.Parent()
	if !parent.IsOk() {
		return parent.UnwrapErr()
	}
	parentPath := parent.Unwrap()

	if !e.fs.Exists(ctx, parentPath.String()) {
		return dot.ErrParentNotFound{Path: parentPath.String()}
	}

	return nil
}

// executeSequential executes operations sequentially, stopping on first failure.
func (e *Executor) executeSequential(ctx context.Context, plan dot.Plan, checkpoint *Checkpoint) ExecutionResult {
	result := ExecutionResult{
		Executed:   []dot.OperationID{},
		Failed:     []dot.OperationID{},
		RolledBack: []dot.OperationID{},
		Errors:     []error{},
	}

	for _, op := range plan.Operations {
		opID := op.ID()

		ctx, span := e.tracer.Start(ctx, "operation.Execute")
		e.log.Debug(ctx, "executing_operation",
			"op_id", opID,
			"op_kind", op.Kind())

		if err := op.Execute(ctx, e.fs); err != nil {
			e.log.Error(ctx, "operation_failed", "op_id", opID, "error", err)
			result.Failed = append(result.Failed, opID)
			result.Errors = append(result.Errors, err)
			span.RecordError(err)
			span.End()
			break
		}

		result.Executed = append(result.Executed, opID)
		checkpoint.Record(opID, op)
		span.End()
	}

	return result
}

// rollback reverses executed operations in reverse order.
func (e *Executor) rollback(ctx context.Context, executed []dot.OperationID, checkpoint *Checkpoint) []dot.OperationID {
	ctx, span := e.tracer.Start(ctx, "executor.Rollback")
	defer span.End()

	e.log.Warn(ctx, "starting_rollback", "operations", len(executed))

	var rolledBack []dot.OperationID

	// Rollback in reverse order
	for i := len(executed) - 1; i >= 0; i-- {
		opID := executed[i]
		op := checkpoint.Lookup(opID)

		if op == nil {
			e.log.Error(ctx, "operation_not_in_checkpoint", "op_id", opID)
			continue
		}

		e.log.Debug(ctx, "rolling_back_operation", "op_id", opID, "op_kind", op.Kind())

		if err := op.Rollback(ctx, e.fs); err != nil {
			e.log.Error(ctx, "rollback_failed", "op_id", opID, "error", err)
			// Continue rolling back other operations
		} else {
			rolledBack = append(rolledBack, opID)
		}
	}

	e.log.Info(ctx, "rollback_complete",
		"attempted", len(executed),
		"succeeded", len(rolledBack))

	return rolledBack
}

// executeParallel executes operations in parallel batches based on dependencies.
func (e *Executor) executeParallel(ctx context.Context, plan dot.Plan, checkpoint *Checkpoint) ExecutionResult {
	batches := plan.ParallelBatches()

	e.log.Info(ctx, "executing_parallel",
		"batch_count", len(batches),
		"total_operations", len(plan.Operations))

	result := ExecutionResult{
		Executed:   []dot.OperationID{},
		Failed:     []dot.OperationID{},
		RolledBack: []dot.OperationID{},
		Errors:     []error{},
	}

	for i, batch := range batches {
		e.log.Debug(ctx, "executing_batch", "batch", i, "size", len(batch))

		batchResult := e.executeBatch(ctx, batch, checkpoint)

		result.Executed = append(result.Executed, batchResult.Executed...)
		result.Failed = append(result.Failed, batchResult.Failed...)
		result.Errors = append(result.Errors, batchResult.Errors...)

		if len(batchResult.Failed) > 0 {
			// Stop on first batch failure
			e.log.Error(ctx, "batch_failed", "batch", i, "failures", len(batchResult.Failed))
			break
		}
	}

	return result
}

// executeBatch executes a batch of operations concurrently.
func (e *Executor) executeBatch(ctx context.Context, batch []dot.Operation, checkpoint *Checkpoint) ExecutionResult {
	result := ExecutionResult{
		Executed:   []dot.OperationID{},
		Failed:     []dot.OperationID{},
		RolledBack: []dot.OperationID{},
		Errors:     []error{},
	}

	if len(batch) == 0 {
		return result
	}

	// If only one operation, execute sequentially
	if len(batch) == 1 {
		op := batch[0]
		opID := op.ID()

		e.log.Debug(ctx, "executing_operation", "op_id", opID, "op_kind", op.Kind())

		if err := op.Execute(ctx, e.fs); err != nil {
			e.log.Error(ctx, "operation_failed", "op_id", opID, "error", err)
			result.Failed = append(result.Failed, opID)
			result.Errors = append(result.Errors, err)
		} else {
			result.Executed = append(result.Executed, opID)
			checkpoint.Record(opID, op)
		}

		return result
	}

	// Execute multiple operations concurrently
	type opResult struct {
		id  dot.OperationID
		err error
	}

	resultCh := make(chan opResult, len(batch))

	for _, op := range batch {
		go func(operation dot.Operation) {
			opID := operation.ID()

			e.log.Debug(ctx, "executing_operation_parallel",
				"op_id", opID,
				"op_kind", operation.Kind())

			err := operation.Execute(ctx, e.fs)
			resultCh <- opResult{id: opID, err: err}
		}(op)
	}

	// Collect results
	opMap := make(map[dot.OperationID]dot.Operation)
	for _, op := range batch {
		opMap[op.ID()] = op
	}

	for i := 0; i < len(batch); i++ {
		res := <-resultCh

		if res.err != nil {
			e.log.Error(ctx, "operation_failed", "op_id", res.id, "error", res.err)
			result.Failed = append(result.Failed, res.id)
			result.Errors = append(result.Errors, res.err)
		} else {
			result.Executed = append(result.Executed, res.id)
			checkpoint.Record(res.id, opMap[res.id])
		}
	}

	return result
}
