package domain

import "os"

// File permission constants for secure and consistent permission management.
// These constants follow the principle of least privilege, defaulting to
// user-only permissions for sensitive data.

const (
	// PermUserRW is user read/write only (0600).
	// Use for: configuration files, manifest files, sensitive data files.
	// Security: Excludes group and world access.
	PermUserRW = os.FileMode(0600)

	// PermUserRWX is user read/write/execute only (0700).
	// Use for: configuration directories, cache directories, data directories.
	// Security: Excludes group and world access.
	PermUserRWX = os.FileMode(0700)

	// PermUserW is user write bit (0200).
	// Use for: permission checks, write capability validation.
	// Security: Isolated write bit for bitwise operations.
	PermUserW = os.FileMode(0200)

	// PermGroupWorld is group/world readable (0044).
	// Use for: detecting insecure permissions, security validation.
	// Security: Used to check if files have insecure group/world access.
	PermGroupWorld = os.FileMode(0044)
)
