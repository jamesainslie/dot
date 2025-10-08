package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultValueConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant interface{}
		expected interface{}
		desc     string
	}{
		// Logging defaults
		{name: "DefaultLogLevel", constant: DefaultLogLevel, expected: "INFO", desc: "default log level"},
		{name: "DefaultLogFormat", constant: DefaultLogFormat, expected: "text", desc: "default log format"},
		{name: "DefaultLogDestination", constant: DefaultLogDestination, expected: "stderr", desc: "default log destination"},

		// Symlink defaults
		{name: "DefaultSymlinkMode", constant: DefaultSymlinkMode, expected: "relative", desc: "default symlink mode"},
		{name: "DefaultSymlinkFolding", constant: DefaultSymlinkFolding, expected: true, desc: "default symlink folding"},
		{name: "DefaultSymlinkOverwrite", constant: DefaultSymlinkOverwrite, expected: false, desc: "default symlink overwrite"},
		{name: "DefaultSymlinkBackup", constant: DefaultSymlinkBackup, expected: false, desc: "default symlink backup"},
		{name: "DefaultSymlinkBackupSuffix", constant: DefaultSymlinkBackupSuffix, expected: ".bak", desc: "default backup suffix"},

		// Dotfile defaults
		{name: "DefaultDotfileTranslate", constant: DefaultDotfileTranslate, expected: true, desc: "default dotfile translation"},
		{name: "DefaultDotfilePrefix", constant: DefaultDotfilePrefix, expected: "dot-", desc: "default dotfile prefix"},

		// Output defaults
		{name: "DefaultOutputFormat", constant: DefaultOutputFormat, expected: "text", desc: "default output format"},
		{name: "DefaultOutputColor", constant: DefaultOutputColor, expected: "auto", desc: "default output color"},
		{name: "DefaultOutputProgress", constant: DefaultOutputProgress, expected: true, desc: "default output progress"},
		{name: "DefaultOutputVerbosity", constant: DefaultOutputVerbosity, expected: 1, desc: "default output verbosity"},
		{name: "DefaultOutputWidth", constant: DefaultOutputWidth, expected: 0, desc: "default output width (auto-detect)"},

		// Operations defaults
		{name: "DefaultOperationsDryRun", constant: DefaultOperationsDryRun, expected: false, desc: "default dry run mode"},
		{name: "DefaultOperationsAtomic", constant: DefaultOperationsAtomic, expected: true, desc: "default atomic operations"},
		{name: "DefaultOperationsMaxParallel", constant: DefaultOperationsMaxParallel, expected: 0, desc: "default max parallel (auto)"},

		// Packages defaults
		{name: "DefaultPackagesSortBy", constant: DefaultPackagesSortBy, expected: "name", desc: "default package sort"},
		{name: "DefaultPackagesAutoDiscover", constant: DefaultPackagesAutoDiscover, expected: false, desc: "default auto-discover"},
		{name: "DefaultPackagesValidateNames", constant: DefaultPackagesValidateNames, expected: true, desc: "default validate names"},

		// Doctor defaults
		{name: "DefaultDoctorAutoFix", constant: DefaultDoctorAutoFix, expected: false, desc: "default auto-fix"},
		{name: "DefaultDoctorCheckManifest", constant: DefaultDoctorCheckManifest, expected: true, desc: "default check manifest"},
		{name: "DefaultDoctorCheckBrokenLinks", constant: DefaultDoctorCheckBrokenLinks, expected: true, desc: "default check broken links"},
		{name: "DefaultDoctorCheckOrphaned", constant: DefaultDoctorCheckOrphaned, expected: false, desc: "default check orphaned"},
		{name: "DefaultDoctorOrphanScanMode", constant: DefaultDoctorOrphanScanMode, expected: "off", desc: "default orphan scan mode"},
		{name: "DefaultDoctorOrphanScanDepth", constant: DefaultDoctorOrphanScanDepth, expected: 0, desc: "default orphan scan depth"},
		{name: "DefaultDoctorCheckPermissions", constant: DefaultDoctorCheckPermissions, expected: false, desc: "default check permissions"},

		// Ignore defaults
		{name: "DefaultIgnoreUseDefaults", constant: DefaultIgnoreUseDefaults, expected: true, desc: "default use ignore defaults"},

		// Experimental defaults
		{name: "DefaultExperimentalParallel", constant: DefaultExperimentalParallel, expected: false, desc: "default experimental parallel"},
		{name: "DefaultExperimentalProfiling", constant: DefaultExperimentalProfiling, expected: false, desc: "default experimental profiling"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant,
				"constant %s should equal %v (%s)", tt.name, tt.expected, tt.desc)
		})
	}
}

func TestDefaultValueTypes(t *testing.T) {
	t.Run("string defaults", func(t *testing.T) {
		stringDefaults := []string{
			DefaultLogLevel,
			DefaultLogFormat,
			DefaultLogDestination,
			DefaultSymlinkMode,
			DefaultSymlinkBackupSuffix,
			DefaultDotfilePrefix,
			DefaultOutputFormat,
			DefaultOutputColor,
			DefaultPackagesSortBy,
			DefaultDoctorOrphanScanMode,
		}

		for _, v := range stringDefaults {
			assert.IsType(t, "", v, "should be string type")
			assert.NotEmpty(t, v, "string default should not be empty")
		}
	})

	t.Run("boolean defaults", func(t *testing.T) {
		boolDefaults := []bool{
			DefaultSymlinkFolding,
			DefaultSymlinkOverwrite,
			DefaultSymlinkBackup,
			DefaultDotfileTranslate,
			DefaultOutputProgress,
			DefaultOperationsDryRun,
			DefaultOperationsAtomic,
			DefaultPackagesAutoDiscover,
			DefaultPackagesValidateNames,
			DefaultDoctorAutoFix,
			DefaultDoctorCheckManifest,
			DefaultDoctorCheckBrokenLinks,
			DefaultDoctorCheckOrphaned,
			DefaultDoctorCheckPermissions,
			DefaultIgnoreUseDefaults,
			DefaultExperimentalParallel,
			DefaultExperimentalProfiling,
		}

		for _, v := range boolDefaults {
			assert.IsType(t, false, v, "should be bool type")
		}
	})

	t.Run("integer defaults", func(t *testing.T) {
		intDefaults := []int{
			DefaultOutputVerbosity,
			DefaultOutputWidth,
			DefaultOperationsMaxParallel,
			DefaultDoctorOrphanScanDepth,
		}

		for _, v := range intDefaults {
			assert.IsType(t, 0, v, "should be int type")
			assert.GreaterOrEqual(t, v, 0, "int default should be non-negative")
		}
	})
}

func TestDefaultValueSemantics(t *testing.T) {
	t.Run("safe defaults", func(t *testing.T) {
		// Defaults should be safe/conservative
		assert.False(t, DefaultOperationsDryRun, "dry-run should default to false (normal operations)")
		assert.True(t, DefaultOperationsAtomic, "operations should default to atomic (safe)")
		assert.False(t, DefaultDoctorAutoFix, "auto-fix should default to false (explicit action)")
		assert.False(t, DefaultSymlinkOverwrite, "overwrite should default to false (safe)")
		assert.False(t, DefaultSymlinkBackup, "backup should default to false (explicit opt-in)")
	})

	t.Run("user-friendly defaults", func(t *testing.T) {
		// Defaults should be user-friendly
		assert.True(t, DefaultSymlinkFolding, "folding should be enabled (convenience)")
		assert.True(t, DefaultDotfileTranslate, "dotfile translation should be enabled (convenience)")
		assert.True(t, DefaultOutputProgress, "progress should be enabled (feedback)")
		assert.Equal(t, "auto", DefaultOutputColor, "color should auto-detect (smart default)")
	})

	t.Run("verbosity levels", func(t *testing.T) {
		// Verbosity: 0 = quiet, 1 = normal, 2 = verbose, 3 = debug
		assert.Equal(t, 1, DefaultOutputVerbosity, "verbosity should default to normal (1)")
		assert.GreaterOrEqual(t, DefaultOutputVerbosity, 0, "verbosity should be non-negative")
		assert.LessOrEqual(t, DefaultOutputVerbosity, 3, "verbosity should be <= 3")
	})

	t.Run("auto-detection defaults", func(t *testing.T) {
		// 0 means auto-detect
		assert.Equal(t, 0, DefaultOutputWidth, "width 0 means auto-detect terminal width")
		assert.Equal(t, 0, DefaultOperationsMaxParallel, "max parallel 0 means auto-detect CPU count")
		assert.Equal(t, 0, DefaultDoctorOrphanScanDepth, "scan depth 0 means unlimited depth")
	})

	t.Run("string format values", func(t *testing.T) {
		// Validate string format values
		validFormats := []string{"text", "json", "yaml", "table"}
		assert.Contains(t, validFormats, DefaultOutputFormat, "output format should be valid")

		validModes := []string{"relative", "absolute"}
		assert.Contains(t, validModes, DefaultSymlinkMode, "symlink mode should be valid")

		validColors := []string{"auto", "always", "never"}
		assert.Contains(t, validColors, DefaultOutputColor, "color mode should be valid")

		validScanModes := []string{"off", "scoped", "deep"}
		assert.Contains(t, validScanModes, DefaultDoctorOrphanScanMode, "scan mode should be valid")
	})
}

func TestDefaultValueDocumentation(t *testing.T) {
	t.Run("usage documentation", func(t *testing.T) {
		defaults := map[string]interface{}{
			"DefaultLogLevel":             DefaultLogLevel,
			"DefaultOutputVerbosity":      DefaultOutputVerbosity,
			"DefaultSymlinkFolding":       DefaultSymlinkFolding,
			"DefaultOperationsAtomic":     DefaultOperationsAtomic,
			"DefaultDoctorCheckManifest":  DefaultDoctorCheckManifest,
			"DefaultDoctorOrphanScanMode": DefaultDoctorOrphanScanMode,
			"DefaultExperimentalParallel": DefaultExperimentalParallel,
		}

		for name, value := range defaults {
			t.Logf("%s: %v", name, value)
		}
	})
}
