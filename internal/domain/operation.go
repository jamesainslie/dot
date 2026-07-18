package domain

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

	// OpKindDirRemoveAll recursively removes a directory and all its contents.
	OpKindDirRemoveAll

	// OpKindFileMove moves a file.
	OpKindFileMove

	// OpKindFileBackup creates a backup copy of a file.
	OpKindFileBackup

	// OpKindFileDelete deletes a file.
	OpKindFileDelete

	// OpKindDirCopy recursively copies a directory.
	OpKindDirCopy
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
	case OpKindDirRemoveAll:
		return "DirRemoveAll"
	case OpKindFileMove:
		return "FileMove"
	case OpKindFileBackup:
		return "FileBackup"
	case OpKindFileDelete:
		return "FileDelete"
	case OpKindDirCopy:
		return "DirCopy"
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

	// Destination is the symlink destination observed when the operation was
	// planned. When set, Execute verifies the link still points there before
	// deleting, and Rollback recreates the link. When empty, Execute still
	// refuses to delete anything that is not a symlink, but Rollback cannot
	// restore the link.
	Destination string
}

// NewLinkDelete creates a new link deletion operation.
func NewLinkDelete(id OperationID, target TargetPath) LinkDelete {
	return LinkDelete{
		OpID:   id,
		Target: target,
	}
}

// NewLinkDeleteWithDestination creates a link deletion operation that records
// the symlink destination observed at plan time.
func NewLinkDeleteWithDestination(id OperationID, target TargetPath, destination string) LinkDelete {
	return LinkDelete{
		OpID:        id,
		Target:      target,
		Destination: destination,
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
	target := op.Target.String()

	// Missing target means the desired state is already achieved (idempotent).
	info, err := fs.Lstat(ctx, target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("lstat %s: %w", target, err)
	}

	// Never delete anything that is not a symlink.
	if info.Mode()&os.ModeSymlink == 0 {
		return ErrUnsafeDelete{
			Path:   target,
			Reason: "expected a symlink but found a regular file or directory",
		}
	}

	// When a destination was recorded at plan time, verify the link still
	// points there before deleting.
	if op.Destination != "" {
		actual, err := fs.ReadLink(ctx, target)
		if err != nil {
			return fmt.Errorf("readlink %s: %w", target, err)
		}
		if actual != op.Destination {
			return ErrUnsafeDelete{
				Path:   target,
				Reason: fmt.Sprintf("symlink points to %q, expected %q", actual, op.Destination),
			}
		}
	}

	err = fs.Remove(ctx, target)
	if err != nil && os.IsNotExist(err) {
		// Already removed - desired state achieved
		return nil
	}
	return err
}

func (op LinkDelete) Rollback(ctx context.Context, fs FS) error {
	target := op.Target.String()

	if op.Destination == "" {
		return ErrRollbackImpossible{
			Path:   target,
			Reason: "symlink destination was not recorded at plan time",
		}
	}

	// If something already exists at the target, only accept a symlink that
	// already points at the recorded destination.
	if info, err := fs.Lstat(ctx, target); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if actual, readErr := fs.ReadLink(ctx, target); readErr == nil && actual == op.Destination {
				return nil
			}
		}
		return fmt.Errorf("cannot restore link %s: path already exists", target)
	}

	if err := fs.Symlink(ctx, op.Destination, target); err != nil {
		return fmt.Errorf("restore link %s -> %s: %w", target, op.Destination, err)
	}
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
	return fs.MkdirAll(ctx, op.Path.String(), DefaultDirPerms)
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

	// Mode is the directory permission mode observed when the operation was
	// planned. Rollback restores it; when zero, DefaultDirPerms is used.
	Mode os.FileMode
}

// NewDirDelete creates a new directory deletion operation.
func NewDirDelete(id OperationID, path FilePath) DirDelete {
	return DirDelete{
		OpID: id,
		Path: path,
	}
}

// NewDirDeleteWithMode creates a directory deletion operation that records the
// directory permissions observed at plan time so rollback can restore them.
func NewDirDeleteWithMode(id OperationID, path FilePath, mode os.FileMode) DirDelete {
	return DirDelete{
		OpID: id,
		Path: path,
		Mode: mode,
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
	mode := op.Mode
	if mode == 0 {
		mode = DefaultDirPerms
	}
	return fs.Mkdir(ctx, op.Path.String(), mode.Perm())
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

// DirRemoveAll recursively removes a directory and all its contents.
type DirRemoveAll struct {
	OpID OperationID
	Path FilePath
}

// NewDirRemoveAll creates a new recursive directory deletion operation.
func NewDirRemoveAll(id OperationID, path FilePath) DirRemoveAll {
	return DirRemoveAll{
		OpID: id,
		Path: path,
	}
}

func (op DirRemoveAll) ID() OperationID {
	return op.OpID
}

func (op DirRemoveAll) Kind() OperationKind {
	return OpKindDirRemoveAll
}

func (op DirRemoveAll) Validate() error {
	if op.OpID == "" {
		return ErrInvalidPath{Path: "", Reason: "operation ID cannot be empty"}
	}
	return nil
}

func (op DirRemoveAll) Dependencies() []Operation {
	return nil
}

func (op DirRemoveAll) Execute(ctx context.Context, fs FS) error {
	return fs.RemoveAll(ctx, op.Path.String())
}

func (op DirRemoveAll) Rollback(ctx context.Context, fs FS) error {
	// Restoring a recursively deleted directory would require the entire
	// directory tree to have been preserved, which it was not.
	return ErrRollbackImpossible{
		Path:   op.Path.String(),
		Reason: "directory contents were not preserved before recursive deletion",
	}
}

func (op DirRemoveAll) String() string {
	return fmt.Sprintf("recursively delete directory %s", op.Path.String())
}

func (op DirRemoveAll) Equals(other Operation) bool {
	if other.Kind() != OpKindDirRemoveAll {
		return false
	}
	o, ok := other.(DirRemoveAll)
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
	// Try rename first (fast path for same filesystem)
	err := fs.Rename(ctx, op.Source.String(), op.Dest.String())
	if err == nil {
		return nil
	}

	// Check if error is due to cross-device link
	// For cross-device moves, copy then delete
	if isCrossDeviceError(err) {
		return op.copyAndDelete(ctx, fs)
	}

	return err
}

// copyAndDelete performs a cross-device move by copying then deleting.
func (op FileMove) copyAndDelete(ctx context.Context, fs FS) error {
	// Check if source is a directory or file
	info, err := fs.Stat(ctx, op.Source.String())
	if err != nil {
		return fmt.Errorf("stat source for cross-device move: %w", err)
	}

	if info.IsDir() {
		// Handle directory move
		if err := copyDirRecursiveHelper(ctx, fs, op.Source.String(), op.Dest.String()); err != nil {
			return fmt.Errorf("copy directory for cross-device move: %w", err)
		}
	} else {
		// Handle file move
		data, err := fs.ReadFile(ctx, op.Source.String())
		if err != nil {
			return fmt.Errorf("read source for cross-device move: %w", err)
		}

		// Write to destination with same permissions
		if err := fs.WriteFile(ctx, op.Dest.String(), data, info.Mode().Perm()); err != nil {
			return fmt.Errorf("write dest for cross-device move: %w", err)
		}
	}

	// Remove source (works for both files and directories)
	if err := fs.RemoveAll(ctx, op.Source.String()); err != nil {
		// Try to clean up the destination
		_ = fs.RemoveAll(ctx, op.Dest.String())
		return fmt.Errorf("remove source for cross-device move: %w", err)
	}

	return nil
}

// isCrossDeviceError checks if an error is a cross-device link error.
func isCrossDeviceError(err error) bool {
	// Check for Unix EXDEV (cross-device link)
	if errors.Is(err, syscall.EXDEV) {
		return true
	}

	// Check for Windows ERROR_NOT_SAME_DEVICE (errno 17)
	var errno syscall.Errno
	if errors.As(err, &errno) {
		// Windows ERROR_NOT_SAME_DEVICE
		if errno == 17 {
			return true
		}
	}

	// Fallback: check error message for cross-device indicators
	msg := err.Error()
	return strings.Contains(msg, "cross-device") || strings.Contains(msg, "invalid cross-device link")
}

func (op FileMove) Rollback(ctx context.Context, fs FS) error {
	// Try rename first (fast path for same filesystem)
	err := fs.Rename(ctx, op.Dest.String(), op.Source.String())
	if err == nil {
		return nil
	}

	// Check if error is due to cross-device link
	if isCrossDeviceError(err) {
		// Create a reversed FileMove operation for rollback
		reversedOp := FileMove{
			OpID:   op.OpID + "-rollback",
			Source: TargetPath{path: op.Dest.path},
			Dest:   FilePath{path: op.Source.path},
		}
		return reversedOp.copyAndDelete(ctx, fs)
	}

	return err
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
	// Get source file info to preserve permissions
	info, err := fs.Stat(ctx, op.Source.String())
	if err != nil {
		return err
	}

	// Read source file data
	data, err := fs.ReadFile(ctx, op.Source.String())
	if err != nil {
		return err
	}

	// Ensure the backup directory exists; nothing earlier in the plan is
	// guaranteed to have created it.
	backupDir := filepath.Dir(op.Backup.String())
	if err := fs.MkdirAll(ctx, backupDir, DefaultDirPerms); err != nil {
		return fmt.Errorf("creating backup directory %s: %w", backupDir, err)
	}

	// Write backup with same permissions as source
	return fs.WriteFile(ctx, op.Backup.String(), data, info.Mode())
}

func (op FileBackup) Rollback(ctx context.Context, fs FS) error {
	source := op.Source.String()
	backup := op.Backup.String()

	// Never delete the backup while the original is missing: the backup may
	// be the only remaining copy of the file.
	info, err := fs.Lstat(ctx, source)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("preserving backup %s: original %s no longer exists", backup, source)
		}
		return fmt.Errorf("lstat %s during backup rollback: %w", source, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("preserving backup %s: original %s is not a regular file", backup, source)
	}

	return fs.Remove(ctx, backup)
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

// FileDelete deletes a file.
type FileDelete struct {
	OpID OperationID
	Path FilePath

	// Backup is the path of a backup copy taken before deletion. When set,
	// Rollback restores the file from it. When empty, rollback is impossible.
	Backup FilePath
}

// NewFileDelete creates a new file delete operation.
func NewFileDelete(id OperationID, path FilePath) FileDelete {
	return FileDelete{
		OpID: id,
		Path: path,
	}
}

// NewFileDeleteWithBackup creates a file delete operation that records the
// path of a backup copy so rollback can restore the deleted file.
func NewFileDeleteWithBackup(id OperationID, path, backup FilePath) FileDelete {
	return FileDelete{
		OpID:   id,
		Path:   path,
		Backup: backup,
	}
}

func (op FileDelete) ID() OperationID {
	return op.OpID
}

func (op FileDelete) Kind() OperationKind {
	return OpKindFileDelete
}

func (op FileDelete) Validate() error {
	if op.OpID == "" {
		return ErrInvalidPath{Path: "", Reason: "operation ID cannot be empty"}
	}
	return nil
}

func (op FileDelete) Dependencies() []Operation {
	return nil
}

func (op FileDelete) Execute(ctx context.Context, fs FS) error {
	return fs.Remove(ctx, op.Path.String())
}

func (op FileDelete) Rollback(ctx context.Context, fs FS) error {
	path := op.Path.String()
	backup := op.Backup.String()

	if backup == "" {
		return ErrRollbackImpossible{
			Path:   path,
			Reason: "no backup was recorded for the deleted file",
		}
	}

	// Refuse to overwrite anything that now occupies the path (for example a
	// symlink whose own rollback failed): writing through it could corrupt
	// package contents.
	if _, err := fs.Lstat(ctx, path); err == nil {
		return fmt.Errorf("cannot restore %s from backup %s: path already exists", path, backup)
	}

	info, err := fs.Stat(ctx, backup)
	if err != nil {
		return fmt.Errorf("stat backup %s: %w", backup, err)
	}
	data, err := fs.ReadFile(ctx, backup)
	if err != nil {
		return fmt.Errorf("read backup %s: %w", backup, err)
	}
	if err := fs.WriteFile(ctx, path, data, info.Mode().Perm()); err != nil {
		return fmt.Errorf("restore %s from backup %s: %w", path, backup, err)
	}
	return nil
}

func (op FileDelete) String() string {
	return fmt.Sprintf("delete file %s", op.Path.String())
}

func (op FileDelete) Equals(other Operation) bool {
	if other.Kind() != OpKindFileDelete {
		return false
	}
	o, ok := other.(FileDelete)
	if !ok {
		return false
	}
	return op.Path.Equals(o.Path)
}

// DirCopy recursively copies a directory without removing the source.
type DirCopy struct {
	OpID   OperationID
	Source FilePath
	Dest   FilePath
}

// NewDirCopy creates a new directory copy operation.
func NewDirCopy(id OperationID, source, dest FilePath) DirCopy {
	return DirCopy{
		OpID:   id,
		Source: source,
		Dest:   dest,
	}
}

func (op DirCopy) ID() OperationID {
	return op.OpID
}

func (op DirCopy) Kind() OperationKind {
	return OpKindDirCopy
}

func (op DirCopy) Validate() error {
	if op.OpID == "" {
		return ErrInvalidPath{Path: "", Reason: "operation ID cannot be empty"}
	}
	return nil
}

func (op DirCopy) Dependencies() []Operation {
	return nil
}

func (op DirCopy) Execute(ctx context.Context, fs FS) error {
	return copyDirRecursiveHelper(ctx, fs, op.Source.String(), op.Dest.String())
}

func (op DirCopy) Rollback(ctx context.Context, fs FS) error {
	// Remove the destination directory
	return fs.RemoveAll(ctx, op.Dest.String())
}

func (op DirCopy) String() string {
	return fmt.Sprintf("copy directory %s -> %s", op.Source.String(), op.Dest.String())
}

func (op DirCopy) Equals(other Operation) bool {
	if other.Kind() != OpKindDirCopy {
		return false
	}
	o, ok := other.(DirCopy)
	if !ok {
		return false
	}
	return op.Source.Equals(o.Source) && op.Dest.Equals(o.Dest)
}

// copyDirRecursiveHelper recursively copies a directory and all its contents.
// This is a package-level helper used by both FileMove and DirCopy operations.
func copyDirRecursiveHelper(ctx context.Context, fs FS, src, dst string) error {
	// Create destination directory
	srcInfo, err := fs.Stat(ctx, src)
	if err != nil {
		return err
	}
	if err := fs.Mkdir(ctx, dst, srcInfo.Mode().Perm()); err != nil {
		return err
	}

	// Read source directory entries
	entries, err := fs.ReadDir(ctx, src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := src + "/" + entry.Name()
		dstPath := dst + "/" + entry.Name()

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDirRecursiveHelper(ctx, fs, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			data, err := fs.ReadFile(ctx, srcPath)
			if err != nil {
				return err
			}

			info, err := fs.Stat(ctx, srcPath)
			if err != nil {
				return err
			}

			if err := fs.WriteFile(ctx, dstPath, data, info.Mode().Perm()); err != nil {
				return err
			}
		}
	}

	return nil
}
