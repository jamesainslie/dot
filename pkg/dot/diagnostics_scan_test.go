package dot_test

import (
	"testing"

	"github.com/jamesainslie/dot/pkg/dot"
	"github.com/stretchr/testify/assert"
)

func TestScanMode_String(t *testing.T) {
	tests := []struct {
		mode dot.ScanMode
		want string
	}{
		{dot.ScanOff, "off"},
		{dot.ScanScoped, "scoped"},
		{dot.ScanDeep, "deep"},
		{dot.ScanMode(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mode.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultScanConfig(t *testing.T) {
	cfg := dot.DefaultScanConfig()

	assert.Equal(t, dot.ScanOff, cfg.Mode)
	assert.Equal(t, 10, cfg.MaxDepth)
	assert.Nil(t, cfg.ScopeToDirs)
	assert.NotEmpty(t, cfg.SkipPatterns)
	assert.Contains(t, cfg.SkipPatterns, ".git")
	assert.Contains(t, cfg.SkipPatterns, "node_modules")
}

func TestScopedScanConfig(t *testing.T) {
	cfg := dot.ScopedScanConfig()

	assert.Equal(t, dot.ScanScoped, cfg.Mode)
	assert.Equal(t, 10, cfg.MaxDepth)
	assert.Nil(t, cfg.ScopeToDirs)
	assert.NotEmpty(t, cfg.SkipPatterns)
}

func TestDeepScanConfig(t *testing.T) {
	t.Run("with positive depth", func(t *testing.T) {
		cfg := dot.DeepScanConfig(15)

		assert.Equal(t, dot.ScanDeep, cfg.Mode)
		assert.Equal(t, 15, cfg.MaxDepth)
		assert.Nil(t, cfg.ScopeToDirs)
		assert.NotEmpty(t, cfg.SkipPatterns)
	})

	t.Run("with zero depth defaults to 10", func(t *testing.T) {
		cfg := dot.DeepScanConfig(0)

		assert.Equal(t, dot.ScanDeep, cfg.Mode)
		assert.Equal(t, 10, cfg.MaxDepth)
	})

	t.Run("with negative depth defaults to 10", func(t *testing.T) {
		cfg := dot.DeepScanConfig(-5)

		assert.Equal(t, dot.ScanDeep, cfg.Mode)
		assert.Equal(t, 10, cfg.MaxDepth)
	})
}

func TestScanConfig_CustomConfiguration(t *testing.T) {
	cfg := dot.ScanConfig{
		Mode:         dot.ScanScoped,
		MaxDepth:     5,
		ScopeToDirs:  []string{"/home/user/.config", "/home/user/.local"},
		SkipPatterns: []string{".git", "target"},
	}

	assert.Equal(t, dot.ScanScoped, cfg.Mode)
	assert.Equal(t, 5, cfg.MaxDepth)
	assert.Len(t, cfg.ScopeToDirs, 2)
	assert.Len(t, cfg.SkipPatterns, 2)
}

