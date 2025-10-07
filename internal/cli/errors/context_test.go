package errors

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtract_WithCommand(t *testing.T) {
	cmd := &cobra.Command{
		Use: "manage",
	}
	cmd.SetArgs([]string{"vim", "tmux"})

	cfg := &ConfigSummary{
		PackageDir: "/home/user/dotfiles",
		TargetDir:  "/home/user",
		DryRun:     false,
		Verbose:    1,
	}

	ctx := Extract(cmd, cfg)

	assert.Contains(t, ctx.Command, "manage")
	assert.Equal(t, "/home/user/dotfiles", ctx.Config.PackageDir)
	assert.Equal(t, "/home/user", ctx.Config.TargetDir)
	assert.False(t, ctx.Config.DryRun)
	assert.Equal(t, 1, ctx.Config.Verbose)
	assert.False(t, ctx.Timestamp.IsZero())
}

func TestExtract_WithNilCommand(t *testing.T) {
	cfg := &ConfigSummary{
		PackageDir: "/home/user/dotfiles",
		TargetDir:  "/home/user",
	}

	ctx := Extract(nil, cfg)

	assert.Equal(t, "", ctx.Command)
	assert.Nil(t, ctx.Arguments)
	assert.Equal(t, "/home/user/dotfiles", ctx.Config.PackageDir)
	assert.False(t, ctx.Timestamp.IsZero())
}

func TestExtract_WithNilConfig(t *testing.T) {
	cmd := &cobra.Command{
		Use: "status",
	}

	ctx := Extract(cmd, nil)

	assert.Contains(t, ctx.Command, "status")
	assert.Equal(t, "", ctx.Config.PackageDir)
	assert.Equal(t, "", ctx.Config.TargetDir)
	assert.False(t, ctx.Timestamp.IsZero())
}

func TestExtract_WithNilBoth(t *testing.T) {
	ctx := Extract(nil, nil)

	assert.Equal(t, "", ctx.Command)
	assert.Nil(t, ctx.Arguments)
	assert.Equal(t, "", ctx.Config.PackageDir)
	assert.False(t, ctx.Timestamp.IsZero())
}

func TestExtractCommand(t *testing.T) {
	cmd := &cobra.Command{
		Use: "list",
	}

	ctx := ExtractCommand(cmd)

	assert.Contains(t, ctx.Command, "list")
	assert.Equal(t, "", ctx.Config.PackageDir)
	assert.False(t, ctx.Timestamp.IsZero())
}

func TestExtractConfig(t *testing.T) {
	cfg := &ConfigSummary{
		PackageDir: "/stow",
		TargetDir:  "/target",
		DryRun:     true,
		Verbose:    2,
	}

	ctx := ExtractConfig(cfg)

	assert.Equal(t, "", ctx.Command)
	assert.Equal(t, "/stow", ctx.Config.PackageDir)
	assert.Equal(t, "/target", ctx.Config.TargetDir)
	assert.True(t, ctx.Config.DryRun)
	assert.Equal(t, 2, ctx.Config.Verbose)
	assert.False(t, ctx.Timestamp.IsZero())
}

func TestConfigSummary_AllFields(t *testing.T) {
	cfg := ConfigSummary{
		PackageDir: "/home/user/dotfiles",
		TargetDir:  "/home/user",
		DryRun:     true,
		Verbose:    3,
	}

	assert.Equal(t, "/home/user/dotfiles", cfg.PackageDir)
	assert.Equal(t, "/home/user", cfg.TargetDir)
	assert.True(t, cfg.DryRun)
	assert.Equal(t, 3, cfg.Verbose)
}

func TestErrorContext_AllFields(t *testing.T) {
	cmd := &cobra.Command{
		Use: "manage",
	}
	cmd.SetArgs([]string{"vim"})

	cfg := &ConfigSummary{
		PackageDir: "/stow",
	}

	ctx := Extract(cmd, cfg)

	require.NotNil(t, ctx)
	assert.NotEmpty(t, ctx.Command)
	assert.NotEmpty(t, ctx.Config.PackageDir)
	assert.False(t, ctx.Timestamp.IsZero())
}
