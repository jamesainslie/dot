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

// PlanMetadata contains statistics about a plan.
type PlanMetadata struct {
	PackageCount   int
	OperationCount int
	LinkCount      int
	DirCount       int
}
