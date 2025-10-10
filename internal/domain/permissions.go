package domain

import "os"

// File and directory permission constants.
// These define the default permissions for filesystem operations.
const (
	// DefaultDirPerms is the default permission mode for directories (rwxr-xr-x).
	// Owner can read, write, and execute. Group and others can read and execute.
	DefaultDirPerms os.FileMode = 0755

	// DefaultFilePerms is the default permission mode for regular files (rw-r--r--).
	// Owner can read and write. Group and others can read only.
	DefaultFilePerms os.FileMode = 0644

	// SecureFilePerms is the permission mode for sensitive files (rw-------).
	// Only owner can read and write. No permissions for group or others.
	SecureFilePerms os.FileMode = 0600

	// SecureDirPerms is the permission mode for sensitive directories (rwx------).
	// Only owner can read, write, and execute. No permissions for group or others.
	SecureDirPerms os.FileMode = 0700
)

// Permission bit constants for testing file mode permissions.
// These can be used with bitwise AND to check specific permission bits.
const (
	// PermUserR is the user read permission bit (0400).
	PermUserR os.FileMode = 0400

	// PermUserW is the user write permission bit (0200).
	PermUserW os.FileMode = 0200

	// PermUserX is the user execute permission bit (0100).
	PermUserX os.FileMode = 0100

	// PermUserRW is the user read+write permission bits (0600).
	PermUserRW os.FileMode = 0600

	// PermUserRWX is the user read+write+execute permission bits (0700).
	PermUserRWX os.FileMode = 0700

	// PermGroupR is the group read permission bit (0040).
	PermGroupR os.FileMode = 0040

	// PermGroupW is the group write permission bit (0020).
	PermGroupW os.FileMode = 0020

	// PermGroupX is the group execute permission bit (0010).
	PermGroupX os.FileMode = 0010

	// PermOtherR is the other read permission bit (0004).
	PermOtherR os.FileMode = 0004

	// PermOtherW is the other write permission bit (0002).
	PermOtherW os.FileMode = 0002

	// PermOtherX is the other execute permission bit (0001).
	PermOtherX os.FileMode = 0001
)
