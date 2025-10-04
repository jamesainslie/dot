package planner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestTopologicalSort_EmptyGraph(t *testing.T) {
	graph := BuildGraph([]dot.Operation{})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	assert.Empty(t, sorted, "empty graph should produce empty result")
}

func TestTopologicalSort_SingleNode(t *testing.T) {
	op := dot.NewLinkCreate(mustParsePath("/a"), mustParsePath("/b"))
	graph := BuildGraph([]dot.Operation{op})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 1)
	assert.True(t, op.Equals(sorted[0]))
}

func TestTopologicalSort_IndependentNodes(t *testing.T) {
	op1 := dot.NewLinkCreate(mustParsePath("/a"), mustParsePath("/b"))
	op2 := dot.NewLinkCreate(mustParsePath("/c"), mustParsePath("/d"))
	graph := BuildGraph([]dot.Operation{op1, op2})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 2)
	// Both operations should be present (order doesn't matter for independent ops)
	assert.Contains(t, sorted, op1)
	assert.Contains(t, sorted, op2)
}

func TestTopologicalSort_LinearChain(t *testing.T) {
	// Create linear dependency: A -> B -> C
	opA := dot.NewDirCreate(mustParsePath("/dir1"))
	opB := &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/dir1/dir2")),
		deps: []dot.Operation{opA},
	}
	opC := &mockOperation{
		op:   dot.NewLinkCreate(mustParsePath("/src"), mustParsePath("/dir1/dir2/file")),
		deps: []dot.Operation{opB},
	}

	graph := BuildGraph([]dot.Operation{opC, opB, opA})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 3)

	// Verify order: A must come before B, B must come before C
	posA := findOperationIndex(sorted, opA)
	posB := findOperationIndex(sorted, opB)
	posC := findOperationIndex(sorted, opC)

	assert.Less(t, posA, posB, "A should come before B")
	assert.Less(t, posB, posC, "B should come before C")
}

func TestTopologicalSort_DiamondPattern(t *testing.T) {
	// Diamond: A -> B, A -> C, B -> D, C -> D
	opA := dot.NewDirCreate(mustParsePath("/root"))
	opB := &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/root/dir1")),
		deps: []dot.Operation{opA},
	}
	opC := &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/root/dir2")),
		deps: []dot.Operation{opA},
	}
	opD := &mockOperation{
		op:   dot.NewLinkCreate(mustParsePath("/src"), mustParsePath("/root/file")),
		deps: []dot.Operation{opB, opC},
	}

	graph := BuildGraph([]dot.Operation{opD, opC, opB, opA})

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 4)

	// Verify dependencies are satisfied
	posA := findOperationIndex(sorted, opA)
	posB := findOperationIndex(sorted, opB)
	posC := findOperationIndex(sorted, opC)
	posD := findOperationIndex(sorted, opD)

	assert.Less(t, posA, posB, "A should come before B")
	assert.Less(t, posA, posC, "A should come before C")
	assert.Less(t, posB, posD, "B should come before D")
	assert.Less(t, posC, posD, "C should come before D")
}

func TestTopologicalSort_ComplexGraph(t *testing.T) {
	// More complex graph with multiple dependencies
	ops := make([]dot.Operation, 6)
	ops[0] = dot.NewDirCreate(mustParsePath("/a"))
	ops[1] = &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/b")),
		deps: []dot.Operation{ops[0]},
	}
	ops[2] = &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/c")),
		deps: []dot.Operation{ops[0]},
	}
	ops[3] = &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/d")),
		deps: []dot.Operation{ops[1], ops[2]},
	}
	ops[4] = &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/e")),
		deps: []dot.Operation{ops[2]},
	}
	ops[5] = &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/f")),
		deps: []dot.Operation{ops[3], ops[4]},
	}

	graph := BuildGraph(ops)

	sorted, err := graph.TopologicalSort()

	require.NoError(t, err)
	require.Len(t, sorted, 6)

	// Verify all dependencies satisfied
	for i, op := range sorted {
		deps := op.Dependencies()
		for _, dep := range deps {
			depPos := findOperationIndex(sorted, dep)
			assert.Less(t, depPos, i, "dependency %v should come before %v", dep, op)
		}
	}
}

func TestFindCycle_NoCycle(t *testing.T) {
	opA := dot.NewDirCreate(mustParsePath("/a"))
	opB := &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/b")),
		deps: []dot.Operation{opA},
	}

	graph := BuildGraph([]dot.Operation{opB, opA})

	cycle := graph.FindCycle()

	assert.Nil(t, cycle, "acyclic graph should not have cycles")
}

func TestFindCycle_SelfLoop(t *testing.T) {
	// Operation depends on itself
	baseOp := dot.NewDirCreate(mustParsePath("/a"))
	opA := &mockOperation{
		op:   baseOp,
		deps: nil, // Will set after creation
	}
	// Create self-reference
	opA.deps = []dot.Operation{opA}

	graph := BuildGraph([]dot.Operation{opA})

	cycle := graph.FindCycle()

	require.NotNil(t, cycle, "self-loop should be detected as cycle")
	require.Len(t, cycle, 1)
	assert.True(t, opA.Equals(cycle[0]))
}

func TestFindCycle_SimpleCycle(t *testing.T) {
	// Create cycle: A -> B -> A
	var opA, opB dot.Operation

	baseA := dot.NewDirCreate(mustParsePath("/a"))
	baseB := dot.NewDirCreate(mustParsePath("/b"))

	opA = &mockOperation{
		op:   baseA,
		deps: []dot.Operation{nil}, // Will set after creating opB
	}
	opB = &mockOperation{
		op:   baseB,
		deps: []dot.Operation{opA},
	}
	// Complete the cycle
	opA.(*mockOperation).deps = []dot.Operation{opB}

	graph := BuildGraph([]dot.Operation{opA, opB})

	cycle := graph.FindCycle()

	require.NotNil(t, cycle, "cycle A->B->A should be detected")
	assert.GreaterOrEqual(t, len(cycle), 2, "cycle should contain at least 2 operations")
}

func TestFindCycle_LongerCycle(t *testing.T) {
	// Create cycle: A -> B -> C -> A
	var opA, opB, opC dot.Operation

	baseA := dot.NewDirCreate(mustParsePath("/a"))
	baseB := dot.NewDirCreate(mustParsePath("/b"))
	baseC := dot.NewDirCreate(mustParsePath("/c"))

	opA = &mockOperation{
		op:   baseA,
		deps: []dot.Operation{nil}, // Will set after creating opC
	}
	opB = &mockOperation{
		op:   baseB,
		deps: []dot.Operation{opA},
	}
	opC = &mockOperation{
		op:   baseC,
		deps: []dot.Operation{opB},
	}
	// Complete the cycle
	opA.(*mockOperation).deps = []dot.Operation{opC}

	graph := BuildGraph([]dot.Operation{opA, opB, opC})

	cycle := graph.FindCycle()

	require.NotNil(t, cycle, "cycle A->B->C->A should be detected")
	assert.GreaterOrEqual(t, len(cycle), 3, "cycle should contain at least 3 operations")
}

func TestTopologicalSort_WithCycle(t *testing.T) {
	// Create cycle: A -> B -> A
	var opA, opB dot.Operation

	baseA := dot.NewDirCreate(mustParsePath("/a"))
	baseB := dot.NewDirCreate(mustParsePath("/b"))

	opA = &mockOperation{
		op:   baseA,
		deps: []dot.Operation{nil},
	}
	opB = &mockOperation{
		op:   baseB,
		deps: []dot.Operation{opA},
	}
	opA.(*mockOperation).deps = []dot.Operation{opB}

	graph := BuildGraph([]dot.Operation{opA, opB})

	sorted, err := graph.TopologicalSort()

	assert.Error(t, err, "cyclic graph should return error")
	assert.Nil(t, sorted, "cyclic graph should return nil operations")

	// Verify error type
	var cyclicErr dot.ErrCyclicDependency
	assert.ErrorAs(t, err, &cyclicErr, "error should be ErrCyclicDependency")
}

func TestTopologicalSort_CycleInLargerGraph(t *testing.T) {
	// Graph with acyclic part and cycle: A, B -> C -> D -> C
	opA := dot.NewDirCreate(mustParsePath("/a"))

	var opC, opD dot.Operation
	baseC := dot.NewDirCreate(mustParsePath("/c"))
	baseD := dot.NewDirCreate(mustParsePath("/d"))

	opC = &mockOperation{
		op:   baseC,
		deps: []dot.Operation{nil}, // Will set after creating opD
	}
	opD = &mockOperation{
		op:   baseD,
		deps: []dot.Operation{opC},
	}
	opC.(*mockOperation).deps = []dot.Operation{opD}

	opB := &mockOperation{
		op:   dot.NewDirCreate(mustParsePath("/b")),
		deps: []dot.Operation{opC},
	}

	graph := BuildGraph([]dot.Operation{opA, opB, opC, opD})

	sorted, err := graph.TopologicalSort()

	assert.Error(t, err, "graph with cycle should return error")
	assert.Nil(t, sorted)
}

// findOperationIndex returns the index of an operation in a slice.
// Returns -1 if not found.
func findOperationIndex(ops []dot.Operation, target dot.Operation) int {
	for i, op := range ops {
		if target.Equals(op) {
			return i
		}
	}
	return -1
}

