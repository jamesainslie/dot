package dot

import "github.com/jamesainslie/dot/internal/domain"

// Operation type re-exports from internal/domain

// OperationKind identifies the type of operation.
type OperationKind = domain.OperationKind

// Operation kind constants
const (
	OpKindLinkCreate = domain.OpKindLinkCreate
	OpKindLinkDelete = domain.OpKindLinkDelete
	OpKindDirCreate  = domain.OpKindDirCreate
	OpKindDirDelete  = domain.OpKindDirDelete
	OpKindFileMove   = domain.OpKindFileMove
	OpKindFileBackup = domain.OpKindFileBackup
)

// OperationID uniquely identifies an operation.
type OperationID = domain.OperationID

// Operation represents a filesystem operation to be executed.
type Operation = domain.Operation

// LinkCreate creates a symbolic link.
type LinkCreate = domain.LinkCreate

// LinkDelete removes a symbolic link.
type LinkDelete = domain.LinkDelete

// DirCreate creates a directory.
type DirCreate = domain.DirCreate

// DirDelete removes a directory.
type DirDelete = domain.DirDelete

// FileMove moves a file from one location to another.
type FileMove = domain.FileMove

// FileBackup backs up a file before modification.
type FileBackup = domain.FileBackup

// NewLinkCreate creates a new LinkCreate operation.
func NewLinkCreate(id OperationID, source, target FilePath) LinkCreate {
	return domain.NewLinkCreate(id, source, target)
}

// NewLinkDelete creates a new LinkDelete operation.
func NewLinkDelete(id OperationID, target FilePath) LinkDelete {
	return domain.NewLinkDelete(id, target)
}

// NewDirCreate creates a new DirCreate operation.
func NewDirCreate(id OperationID, path FilePath) DirCreate {
	return domain.NewDirCreate(id, path)
}

// NewDirDelete creates a new DirDelete operation.
func NewDirDelete(id OperationID, path FilePath) DirDelete {
	return domain.NewDirDelete(id, path)
}

// NewFileMove creates a new FileMove operation.
func NewFileMove(id OperationID, source, dest FilePath) FileMove {
	return domain.NewFileMove(id, source, dest)
}

// NewFileBackup creates a new FileBackup operation.
func NewFileBackup(id OperationID, source, backup FilePath) FileBackup {
	return domain.NewFileBackup(id, source, backup)
}