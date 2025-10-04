package planner

import "github.com/jamesainslie/dot/pkg/dot"

// DependencyGraph represents operation dependencies for topological sorting.
// It maintains a directed graph where nodes are operations and edges represent
// dependencies (an edge from A to B means A depends on B).
type DependencyGraph struct {
	// nodes maps each operation to its index for quick lookup
	nodes map[dot.Operation]int

	// ops stores operations in insertion order
	ops []dot.Operation

	// edges stores dependencies: edges[op] = operations that op depends on
	edges map[dot.Operation][]dot.Operation
}

// BuildGraph constructs a dependency graph from a list of operations.
// It analyzes the Dependencies() of each operation to build the graph edges.
//
// Time complexity: O(n + e) where n is the number of operations and e is
// the total number of dependencies across all operations.
func BuildGraph(ops []dot.Operation) *DependencyGraph {
	graph := &DependencyGraph{
		nodes: make(map[dot.Operation]int, len(ops)),
		ops:   make([]dot.Operation, 0, len(ops)),
		edges: make(map[dot.Operation][]dot.Operation, len(ops)),
	}

	// Add all operations as nodes
	for i, op := range ops {
		graph.nodes[op] = i
		graph.ops = append(graph.ops, op)

		// Build edges from dependencies
		deps := op.Dependencies()
		if len(deps) > 0 {
			graph.edges[op] = deps
		}
	}

	return graph
}

// Size returns the number of operations in the graph.
func (g *DependencyGraph) Size() int {
	return len(g.ops)
}

// HasOperation returns true if the operation exists in the graph.
func (g *DependencyGraph) HasOperation(op dot.Operation) bool {
	_, exists := g.nodes[op]
	return exists
}

// Dependencies returns the list of operations that the given operation depends on.
// Returns an empty slice if the operation has no dependencies or is not in the graph.
func (g *DependencyGraph) Dependencies(op dot.Operation) []dot.Operation {
	if deps, exists := g.edges[op]; exists {
		// Return a copy to prevent external modification
		result := make([]dot.Operation, len(deps))
		copy(result, deps)
		return result
	}
	return nil
}

// Operations returns all operations in the graph.
// The returned slice is a copy to prevent external modification.
func (g *DependencyGraph) Operations() []dot.Operation {
	result := make([]dot.Operation, len(g.ops))
	copy(result, g.ops)
	return result
}
