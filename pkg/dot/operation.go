package dot

import "fmt"

// OperationKind identifies the type of operation.
type OperationKind int

const (
	// OpKindLinkCreate creates a symbolic link.
	OpKindLinkCreate OperationKind = iota
	
	// OpKindLinkDelete removes a symbolic link.
	OpKindLinkDelete
	
	// OpKindDirCreate creates a directory.
	OpKindDirCreate
	
	// OpKindDirDelete removes an empty directory.
	OpKindDirDelete
	
	// OpKindFileMove moves a file.
	OpKindFileMove
	
	// OpKindFileBackup creates a backup copy of a file.
	OpKindFileBackup
)

// String returns the string representation of an OperationKind.
func (k OperationKind) String() string {
	switch k {
	case OpKindLinkCreate:
		return "LinkCreate"
	case OpKindLinkDelete:
		return "LinkDelete"
	case OpKindDirCreate:
		return "DirCreate"
	case OpKindDirDelete:
		return "DirDelete"
	case OpKindFileMove:
		return "FileMove"
	case OpKindFileBackup:
		return "FileBackup"
	default:
		return "Unknown"
	}
}

// Operation represents a filesystem operation.
// Operations are pure data structures with no side effects.
type Operation interface {
	// Kind returns the operation type.
	Kind() OperationKind
	
	// Validate checks if the operation is valid.
	Validate() error
	
	// Dependencies returns operations that must execute before this one.
	Dependencies() []Operation
	
	// String returns a human-readable description.
	String() string
	
	// Equals checks if two operations are equivalent.
	Equals(other Operation) bool
}

// LinkCreate creates a symbolic link from source to target.
type LinkCreate struct {
	Source FilePath
	Target FilePath
}

// NewLinkCreate creates a new link creation operation.
func NewLinkCreate(source, target FilePath) LinkCreate {
	return LinkCreate{
		Source: source,
		Target: target,
	}
}

func (op LinkCreate) Kind() OperationKind {
	return OpKindLinkCreate
}

func (op LinkCreate) Validate() error {
	// Validation will be implemented when we have filesystem access
	return nil
}

func (op LinkCreate) Dependencies() []Operation {
	return nil
}

func (op LinkCreate) String() string {
	return fmt.Sprintf("create link %s -> %s", op.Target.String(), op.Source.String())
}

func (op LinkCreate) Equals(other Operation) bool {
	if other.Kind() != OpKindLinkCreate {
		return false
	}
	o, ok := other.(LinkCreate)
	if !ok {
		return false
	}
	return op.Source.Equals(o.Source) && op.Target.Equals(o.Target)
}

// LinkDelete removes a symbolic link at target.
type LinkDelete struct {
	Target FilePath
}

// NewLinkDelete creates a new link deletion operation.
func NewLinkDelete(target FilePath) LinkDelete {
	return LinkDelete{
		Target: target,
	}
}

func (op LinkDelete) Kind() OperationKind {
	return OpKindLinkDelete
}

func (op LinkDelete) Validate() error {
	return nil
}

func (op LinkDelete) Dependencies() []Operation {
	return nil
}

func (op LinkDelete) String() string {
	return fmt.Sprintf("delete link %s", op.Target.String())
}

func (op LinkDelete) Equals(other Operation) bool {
	if other.Kind() != OpKindLinkDelete {
		return false
	}
	o, ok := other.(LinkDelete)
	if !ok {
		return false
	}
	return op.Target.Equals(o.Target)
}

// DirCreate creates a directory at path.
type DirCreate struct {
	Path FilePath
}

// NewDirCreate creates a new directory creation operation.
func NewDirCreate(path FilePath) DirCreate {
	return DirCreate{
		Path: path,
	}
}

func (op DirCreate) Kind() OperationKind {
	return OpKindDirCreate
}

func (op DirCreate) Validate() error {
	return nil
}

func (op DirCreate) Dependencies() []Operation {
	return nil
}

func (op DirCreate) String() string {
	return fmt.Sprintf("create directory %s", op.Path.String())
}

func (op DirCreate) Equals(other Operation) bool {
	if other.Kind() != OpKindDirCreate {
		return false
	}
	o, ok := other.(DirCreate)
	if !ok {
		return false
	}
	return op.Path.Equals(o.Path)
}

// DirDelete removes an empty directory at path.
type DirDelete struct {
	Path FilePath
}

// NewDirDelete creates a new directory deletion operation.
func NewDirDelete(path FilePath) DirDelete {
	return DirDelete{
		Path: path,
	}
}

func (op DirDelete) Kind() OperationKind {
	return OpKindDirDelete
}

func (op DirDelete) Validate() error {
	return nil
}

func (op DirDelete) Dependencies() []Operation {
	return nil
}

func (op DirDelete) String() string {
	return fmt.Sprintf("delete directory %s", op.Path.String())
}

func (op DirDelete) Equals(other Operation) bool {
	if other.Kind() != OpKindDirDelete {
		return false
	}
	o, ok := other.(DirDelete)
	if !ok {
		return false
	}
	return op.Path.Equals(o.Path)
}

// FileMove moves a file from source to destination.
type FileMove struct {
	Source FilePath
	Dest   FilePath
}

// NewFileMove creates a new file move operation.
func NewFileMove(source, dest FilePath) FileMove {
	return FileMove{
		Source: source,
		Dest:   dest,
	}
}

func (op FileMove) Kind() OperationKind {
	return OpKindFileMove
}

func (op FileMove) Validate() error {
	return nil
}

func (op FileMove) Dependencies() []Operation {
	return nil
}

func (op FileMove) String() string {
	return fmt.Sprintf("move file %s -> %s", op.Source.String(), op.Dest.String())
}

func (op FileMove) Equals(other Operation) bool {
	if other.Kind() != OpKindFileMove {
		return false
	}
	o, ok := other.(FileMove)
	if !ok {
		return false
	}
	return op.Source.Equals(o.Source) && op.Dest.Equals(o.Dest)
}

// FileBackup creates a backup copy of a file.
type FileBackup struct {
	Source FilePath
	Backup FilePath
}

// NewFileBackup creates a new file backup operation.
func NewFileBackup(source, backup FilePath) FileBackup {
	return FileBackup{
		Source: source,
		Backup: backup,
	}
}

func (op FileBackup) Kind() OperationKind {
	return OpKindFileBackup
}

func (op FileBackup) Validate() error {
	return nil
}

func (op FileBackup) Dependencies() []Operation {
	return nil
}

func (op FileBackup) String() string {
	return fmt.Sprintf("backup file %s -> %s", op.Source.String(), op.Backup.String())
}

func (op FileBackup) Equals(other Operation) bool {
	if other.Kind() != OpKindFileBackup {
		return false
	}
	o, ok := other.(FileBackup)
	if !ok {
		return false
	}
	return op.Source.Equals(o.Source) && op.Backup.Equals(o.Backup)
}

