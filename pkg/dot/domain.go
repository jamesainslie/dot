package dot

// Package represents a collection of configuration files to be managed.
type Package struct {
	Name string
	Path PackagePath
	Tree *Node // Optional: file tree for the package
}

// NodeType identifies the type of filesystem node.
type NodeType int

const (
	// NodeFile represents a regular file.
	NodeFile NodeType = iota

	// NodeDir represents a directory.
	NodeDir

	// NodeSymlink represents a symbolic link.
	NodeSymlink
)

// String returns the string representation of a NodeType.
func (t NodeType) String() string {
	switch t {
	case NodeFile:
		return "File"
	case NodeDir:
		return "Dir"
	case NodeSymlink:
		return "Symlink"
	default:
		return "Unknown"
	}
}

// Node represents a node in a filesystem tree.
type Node struct {
	Path     FilePath
	Type     NodeType
	Children []Node
}

// IsFile returns true if the node is a file.
func (n Node) IsFile() bool {
	return n.Type == NodeFile
}

// IsDir returns true if the node is a directory.
func (n Node) IsDir() bool {
	return n.Type == NodeDir
}

// IsSymlink returns true if the node is a symbolic link.
func (n Node) IsSymlink() bool {
	return n.Type == NodeSymlink
}

// Plan represents a set of operations to execute.
type Plan struct {
	Operations []Operation
	Metadata   PlanMetadata
	Batches    [][]Operation // Parallel execution batches (if computed)
}

// Validate checks if the plan is valid.
func (p Plan) Validate() error {
	// Validate each operation
	for _, op := range p.Operations {
		if err := op.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// CanParallelize returns true if the plan has computed parallel batches.
func (p Plan) CanParallelize() bool {
	return len(p.Batches) > 0
}

// ParallelBatches returns the parallel execution batches.
// Returns nil if parallelization has not been computed.
func (p Plan) ParallelBatches() [][]Operation {
	return p.Batches
}

// PlanMetadata contains statistics and diagnostic information about a plan.
type PlanMetadata struct {
	PackageCount   int            `json:"package_count"`
	OperationCount int            `json:"operation_count"`
	LinkCount      int            `json:"link_count"`
	DirCount       int            `json:"dir_count"`
	Conflicts      []ConflictInfo `json:"conflicts,omitempty"`
	Warnings       []WarningInfo  `json:"warnings,omitempty"`
}
