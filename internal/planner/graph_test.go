package planner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamesainslie/dot/pkg/dot"
)

func TestBuildGraph_Empty(t *testing.T) {
	ops := []dot.Operation{}

	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 0, graph.Size(), "empty operation list should produce empty graph")
}

func TestBuildGraph_SingleOperation(t *testing.T) {
	source := mustParsePath("/stow/package/file")
	target := mustParsePath("/home/user/.config/file")
	op := dot.NewLinkCreate(source, target)

	ops := []dot.Operation{op}
	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 1, graph.Size(), "single operation should produce graph with one node")
	assert.True(t, graph.HasOperation(op), "graph should contain the operation")
}

func TestBuildGraph_IndependentOperations(t *testing.T) {
	op1 := dot.NewLinkCreate(
		mustParsePath("/stow/pkg/file1"),
		mustParsePath("/home/user/file1"),
	)
	op2 := dot.NewLinkCreate(
		mustParsePath("/stow/pkg/file2"),
		mustParsePath("/home/user/file2"),
	)

	ops := []dot.Operation{op1, op2}
	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 2, graph.Size())
	assert.True(t, graph.HasOperation(op1))
	assert.True(t, graph.HasOperation(op2))

	// No dependencies between operations
	deps1 := graph.Dependencies(op1)
	deps2 := graph.Dependencies(op2)
	assert.Empty(t, deps1, "independent operations should have no dependencies")
	assert.Empty(t, deps2, "independent operations should have no dependencies")
}

func TestBuildGraph_LinearDependencies(t *testing.T) {
	// Create a linear dependency chain: dirCreate -> linkCreate
	dirPath := mustParsePath("/home/user/.config")
	dirOp := dot.NewDirCreate(dirPath)

	linkOp := dot.NewLinkCreate(
		mustParsePath("/stow/pkg/config"),
		mustParsePath("/home/user/.config/app.conf"),
	)

	// Mock linkOp to depend on dirOp
	linkOpWithDep := &mockOperation{
		op:   linkOp,
		deps: []dot.Operation{dirOp},
	}

	ops := []dot.Operation{linkOpWithDep, dirOp}
	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 2, graph.Size())

	// linkOp should depend on dirOp
	deps := graph.Dependencies(linkOpWithDep)
	require.Len(t, deps, 1)
	assert.True(t, dirOp.Equals(deps[0]), "link should depend on directory creation")

	// dirOp should have no dependencies
	dirDeps := graph.Dependencies(dirOp)
	assert.Empty(t, dirDeps)
}

func TestBuildGraph_DiamondPattern(t *testing.T) {
	// Diamond dependency: A -> B, A -> C, B -> D, C -> D
	opA := dot.NewDirCreate(mustParsePath("/home/user/.config"))
	opB := dot.NewDirCreate(mustParsePath("/home/user/.config/app1"))
	opC := dot.NewDirCreate(mustParsePath("/home/user/.config/app2"))
	opD := dot.NewLinkCreate(
		mustParsePath("/stow/pkg/file"),
		mustParsePath("/home/user/.config/file"),
	)

	// Create mock operations with dependencies
	opBWithDep := &mockOperation{op: opB, deps: []dot.Operation{opA}}
	opCWithDep := &mockOperation{op: opC, deps: []dot.Operation{opA}}
	opDWithDep := &mockOperation{op: opD, deps: []dot.Operation{opBWithDep, opCWithDep}}

	ops := []dot.Operation{opDWithDep, opCWithDep, opBWithDep, opA}
	graph := BuildGraph(ops)

	require.NotNil(t, graph)
	assert.Equal(t, 4, graph.Size())

	// Verify dependencies
	assert.Empty(t, graph.Dependencies(opA))
	assert.Len(t, graph.Dependencies(opBWithDep), 1)
	assert.Len(t, graph.Dependencies(opCWithDep), 1)
	assert.Len(t, graph.Dependencies(opDWithDep), 2)
}

func TestGraph_Size(t *testing.T) {
	tests := []struct {
		name     string
		ops      []dot.Operation
		expected int
	}{
		{
			name:     "empty graph",
			ops:      []dot.Operation{},
			expected: 0,
		},
		{
			name: "single operation",
			ops: []dot.Operation{
				dot.NewLinkCreate(mustParsePath("/a"), mustParsePath("/b")),
			},
			expected: 1,
		},
		{
			name: "multiple operations",
			ops: []dot.Operation{
				dot.NewLinkCreate(mustParsePath("/a"), mustParsePath("/b")),
				dot.NewLinkCreate(mustParsePath("/c"), mustParsePath("/d")),
				dot.NewDirCreate(mustParsePath("/e")),
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := BuildGraph(tt.ops)
			assert.Equal(t, tt.expected, graph.Size())
		})
	}
}

func TestGraph_HasOperation(t *testing.T) {
	op1 := dot.NewLinkCreate(mustParsePath("/a"), mustParsePath("/b"))
	op2 := dot.NewLinkCreate(mustParsePath("/c"), mustParsePath("/d"))
	op3 := dot.NewDirCreate(mustParsePath("/e"))

	graph := BuildGraph([]dot.Operation{op1, op2})

	assert.True(t, graph.HasOperation(op1))
	assert.True(t, graph.HasOperation(op2))
	assert.False(t, graph.HasOperation(op3), "operation not in graph should return false")
}

// mockOperation wraps an operation with custom dependencies for testing
type mockOperation struct {
	op   dot.Operation
	deps []dot.Operation
}

func (m *mockOperation) Kind() dot.OperationKind {
	return m.op.Kind()
}

func (m *mockOperation) Validate() error {
	return m.op.Validate()
}

func (m *mockOperation) Dependencies() []dot.Operation {
	return m.deps
}

func (m *mockOperation) String() string {
	return m.op.String()
}

func (m *mockOperation) Equals(other dot.Operation) bool {
	if otherMock, ok := other.(*mockOperation); ok {
		return m.op.Equals(otherMock.op)
	}
	return m.op.Equals(other)
}

// mustParsePath creates a FilePath or panics (for test convenience)
func mustParsePath(s string) dot.FilePath {
	result := dot.NewFilePath(s)
	if !result.IsOk() {
		panic(result.UnwrapErr())
	}
	return result.Unwrap()
}
