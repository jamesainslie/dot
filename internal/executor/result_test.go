package executor

import (
	"errors"
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/require"
)

func TestExecutionResult_Success(t *testing.T) {
	t.Run("success with no failures", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []dot.OperationID{"op1", "op2"},
			Failed:     []dot.OperationID{},
			RolledBack: []dot.OperationID{},
			Errors:     []error{},
		}

		require.True(t, result.Success())
	})

	t.Run("failure with failed operations", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []dot.OperationID{"op1"},
			Failed:     []dot.OperationID{"op2"},
			RolledBack: []dot.OperationID{},
			Errors:     []error{errors.New("op2 failed")},
		}

		require.False(t, result.Success())
	})

	t.Run("failure with errors but no failed ops", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []dot.OperationID{"op1"},
			Failed:     []dot.OperationID{},
			RolledBack: []dot.OperationID{},
			Errors:     []error{errors.New("some error")},
		}

		require.False(t, result.Success())
	})
}

func TestExecutionResult_PartialFailure(t *testing.T) {
	t.Run("partial failure with some executed and some failed", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []dot.OperationID{"op1", "op2"},
			Failed:     []dot.OperationID{"op3"},
			RolledBack: []dot.OperationID{},
			Errors:     []error{errors.New("op3 failed")},
		}

		require.True(t, result.PartialFailure())
	})

	t.Run("not partial if all succeeded", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []dot.OperationID{"op1", "op2"},
			Failed:     []dot.OperationID{},
			RolledBack: []dot.OperationID{},
			Errors:     []error{},
		}

		require.False(t, result.PartialFailure())
	})

	t.Run("not partial if all failed", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []dot.OperationID{},
			Failed:     []dot.OperationID{"op1", "op2"},
			RolledBack: []dot.OperationID{},
			Errors:     []error{errors.New("all failed")},
		}

		require.False(t, result.PartialFailure())
	})

	t.Run("not partial if nothing executed", func(t *testing.T) {
		result := ExecutionResult{
			Executed:   []dot.OperationID{},
			Failed:     []dot.OperationID{},
			RolledBack: []dot.OperationID{},
			Errors:     []error{},
		}

		require.False(t, result.PartialFailure())
	})
}
