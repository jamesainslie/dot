package main

import (
	"runtime/debug"
	"testing"
)

func TestEffectiveVersion(t *testing.T) {
	t.Parallel()

	buildInfo := func(v string) *debug.BuildInfo {
		return &debug.BuildInfo{Main: debug.Module{Version: v}}
	}

	tests := []struct {
		name    string
		ldflags string
		info    *debug.BuildInfo
		want    string
	}{
		{
			name:    "ldflags stamped version wins",
			ldflags: "v0.7.0",
			info:    buildInfo("v0.6.6"),
			want:    "v0.7.0",
		},
		{
			name:    "go install build falls back to module version",
			ldflags: "dev",
			info:    buildInfo("v0.6.6"),
			want:    "v0.6.6",
		},
		{
			name:    "source build in repo reports dev",
			ldflags: "dev",
			info:    buildInfo("(devel)"),
			want:    "dev",
		},
		{
			name:    "missing build info reports dev",
			ldflags: "dev",
			info:    nil,
			want:    "dev",
		},
		{
			name:    "empty ldflags with module version",
			ldflags: "",
			info:    buildInfo("v0.6.6"),
			want:    "v0.6.6",
		},
		{
			name:    "empty everything degrades to dev",
			ldflags: "",
			info:    &debug.BuildInfo{},
			want:    "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := effectiveVersion(tt.ldflags, tt.info)
			if got != tt.want {
				t.Errorf("effectiveVersion(%q, ...) = %q, want %q", tt.ldflags, got, tt.want)
			}
		})
	}
}
