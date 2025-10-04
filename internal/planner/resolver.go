package planner

// ConflictType categorizes conflicts by their nature
type ConflictType int

const (
	// ConflictFileExists indicates a file exists at the link target location
	ConflictFileExists ConflictType = iota
	// ConflictWrongLink indicates a symlink points to the wrong source
	ConflictWrongLink
	// ConflictPermission indicates permission denied for operation
	ConflictPermission
	// ConflictCircular indicates a circular symlink dependency
	ConflictCircular
	// ConflictDirExpected indicates a directory was expected but file found
	ConflictDirExpected
	// ConflictFileExpected indicates a file was expected but directory found
	ConflictFileExpected
)

// String returns the string representation of ConflictType
func (ct ConflictType) String() string {
	switch ct {
	case ConflictFileExists:
		return "file_exists"
	case ConflictWrongLink:
		return "wrong_link"
	case ConflictPermission:
		return "permission"
	case ConflictCircular:
		return "circular"
	case ConflictDirExpected:
		return "dir_expected"
	case ConflictFileExpected:
		return "file_expected"
	default:
		return "unknown"
	}
}

