package domain

import (
	"context"
	"fmt"
)

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

// OperationID uniquely identifies an operation.
type OperationID string

// Operation represents a filesystem operation.
// Operations are pure data structures with no side effects.
type Operation interface {
	// ID returns the unique identifier for this operation.
	ID() OperationID

	// Kind returns the operation type.
	Kind() OperationKind

	// Validate checks if the operation is valid.
	Validate() error

	// Dependencies returns operations that must execute before this one.
	Dependencies() []Operation

	// Execute performs the operation with side effects.
	Execute(ctx context.Context, fs FS) error

	// Rollback undoes the operation.
	Rollback(ctx context.Context, fs FS) error

	// String returns a human-readable description.
	String() string

	// Equals checks if two operations are equivalent.
	Equals(other Operation) bool
}

// LinkCreate creates a symbolic link from source to target.
type LinkCreate struct {
	OpID   OperationID
	Source FilePath
	Target TargetPath
}

// NewLinkCreate creates a new link creation operation.
func NewLinkCreate(id OperationID, source FilePath, target TargetPath) LinkCreate {
	return LinkCreate{
		OpID:   id,
		Source: source,
		Target: target,
	}
}

func (op LinkCreate) ID() OperationID {
	return op.OpID
}

func (op LinkCreate) Kind() OperationKind {
	return OpKindLinkCreate
}

func (op LinkCreate) Validate() error {
	if op.OpID == "" {
		return ErrInvalidPath{Path: "", Reason: "operation ID cannot be empty"}
	}
	return nil
}

func (op LinkCreate) Dependencies() []Operation {
	return nil
}

func (op LinkCreate) Execute(ctx context.Context, fs FS) error {
	return fs.Symlink(ctx, op.Source.String(), op.Target.String())
}

func (op LinkCreate) Rollback(ctx context.Context, fs FS) error {
	return fs.Remove(ctx, op.Target.String())
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
	OpID   OperationID
	Target TargetPath
}

// NewLinkDelete creates a new link deletion operation.
func NewLinkDelete(id OperationID, target TargetPath) LinkDelete {
	return LinkDelete{
		OpID:   id,
		Target: target,
	}
}

func (op LinkDelete) ID() OperationID {
	return op.OpID
}

func (op LinkDelete) Kind() OperationKind {
	return OpKindLinkDelete
}

func (op LinkDelete) Validate() error {
	if op.OpID == "" {
		return ErrInvalidPath{Path: "", Reason: "operation ID cannot be empty"}
	}
	return nil
}

func (op LinkDelete) Dependencies() []Operation {
	return nil
}

func (op LinkDelete) Execute(ctx context.Context, fs FS) error {
	return fs.Remove(ctx, op.Target.String())
}

func (op LinkDelete) Rollback(ctx context.Context, fs FS) error {
	// Cannot restore deleted link without knowing original target
	// This would require storing the original target in the operation
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
	OpID OperationID
	Path FilePath
}

// NewDirCreate creates a new directory creation operation.
func NewDirCreate(id OperationID, path FilePath) DirCreate {
	return DirCreate{
		OpID: id,
		Path: path,
	}
}

func (op DirCreate) ID() OperationID {
	return op.OpID
}

func (op DirCreate) Kind() OperationKind {
	return OpKindDirCreate
}

func (op DirCreate) Validate() error {
	if op.OpID == "" {
		return ErrInvalidPath{Path: "", Reason: "operation ID cannot be empty"}
	}
	return nil
}

func (op DirCreate) Dependencies() []Operation {
	return nil
}

func (op DirCreate) Execute(ctx context.Context, fs FS) error {
	return fs.MkdirAll(ctx, op.Path.String(), 0755)
}

func (op DirCreate) Rollback(ctx context.Context, fs FS) error {
	return fs.Remove(ctx, op.Path.String())
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
	OpID OperationID
	Path FilePath
}

// NewDirDelete creates a new directory deletion operation.
func NewDirDelete(id OperationID, path FilePath) DirDelete {
	return DirDelete{
		OpID: id,
		Path: path,
	}
}

func (op DirDelete) ID() OperationID {
	return op.OpID
}

func (op DirDelete) Kind() OperationKind {
	return OpKindDirDelete
}

func (op DirDelete) Validate() error {
	if op.OpID == "" {
		return ErrInvalidPath{Path: "", Reason: "operation ID cannot be empty"}
	}
	return nil
}

func (op DirDelete) Dependencies() []Operation {
	return nil
}

func (op DirDelete) Execute(ctx context.Context, fs FS) error {
	return fs.Remove(ctx, op.Path.String())
}

func (op DirDelete) Rollback(ctx context.Context, fs FS) error {
	return fs.Mkdir(ctx, op.Path.String(), 0755)
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
	OpID   OperationID
	Source TargetPath
	Dest   FilePath
}

// NewFileMove creates a new file move operation.
func NewFileMove(id OperationID, source TargetPath, dest FilePath) FileMove {
	return FileMove{
		OpID:   id,
		Source: source,
		Dest:   dest,
	}
}

func (op FileMove) ID() OperationID {
	return op.OpID
}

func (op FileMove) Kind() OperationKind {
	return OpKindFileMove
}

func (op FileMove) Validate() error {
	if op.OpID == "" {
		return ErrInvalidPath{Path: "", Reason: "operation ID cannot be empty"}
	}
	return nil
}

func (op FileMove) Dependencies() []Operation {
	return nil
}

func (op FileMove) Execute(ctx context.Context, fs FS) error {
	return fs.Rename(ctx, op.Source.String(), op.Dest.String())
}

func (op FileMove) Rollback(ctx context.Context, fs FS) error {
	return fs.Rename(ctx, op.Dest.String(), op.Source.String())
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
	OpID   OperationID
	Source FilePath
	Backup FilePath
}

// NewFileBackup creates a new file backup operation.
func NewFileBackup(id OperationID, source, backup FilePath) FileBackup {
	return FileBackup{
		OpID:   id,
		Source: source,
		Backup: backup,
	}
}

func (op FileBackup) ID() OperationID {
	return op.OpID
}

func (op FileBackup) Kind() OperationKind {
	return OpKindFileBackup
}

func (op FileBackup) Validate() error {
	if op.OpID == "" {
		return ErrInvalidPath{Path: "", Reason: "operation ID cannot be empty"}
	}
	return nil
}

func (op FileBackup) Dependencies() []Operation {
	return nil
}

func (op FileBackup) Execute(ctx context.Context, fs FS) error {
	data, err := fs.ReadFile(ctx, op.Source.String())
	if err != nil {
		return err
	}
	return fs.WriteFile(ctx, op.Backup.String(), data, 0644)
}

func (op FileBackup) Rollback(ctx context.Context, fs FS) error {
	return fs.Remove(ctx, op.Backup.String())
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
