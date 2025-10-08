package domain

// ExecutionResult contains the outcome of plan execution.
type ExecutionResult struct {
	Executed   []OperationID
	Failed     []OperationID
	RolledBack []OperationID
	Errors     []error
}

// Success returns true if all operations executed successfully.
func (r ExecutionResult) Success() bool {
	return len(r.Failed) == 0 && len(r.Errors) == 0
}

// PartialFailure returns true if some but not all operations succeeded.
func (r ExecutionResult) PartialFailure() bool {
	return len(r.Executed) > 0 && len(r.Failed) > 0
}
