package domain

import (
	"os"
	"testing"
)

func TestPermissionConstants(t *testing.T) {
	tests := []struct {
		name     string
		perm     os.FileMode
		expected os.FileMode
	}{
		{"DefaultDirPerms", DefaultDirPerms, 0755},
		{"DefaultFilePerms", DefaultFilePerms, 0644},
		{"SecureFilePerms", SecureFilePerms, 0600},
		{"SecureDirPerms", SecureDirPerms, 0700},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.perm != tt.expected {
				t.Errorf("%s = %o, want %o", tt.name, tt.perm, tt.expected)
			}
		})
	}
}
