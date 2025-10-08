package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigurationKeyConstants(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
		category string
	}{
		// Directory keys
		{name: "KeyDirPackage", key: KeyDirPackage, expected: "directories.package", category: "directories"},
		{name: "KeyDirTarget", key: KeyDirTarget, expected: "directories.target", category: "directories"},
		{name: "KeyDirManifest", key: KeyDirManifest, expected: "directories.manifest", category: "directories"},

		// Logging keys
		{name: "KeyLogLevel", key: KeyLogLevel, expected: "logging.level", category: "logging"},
		{name: "KeyLogFormat", key: KeyLogFormat, expected: "logging.format", category: "logging"},
		{name: "KeyLogDestination", key: KeyLogDestination, expected: "logging.destination", category: "logging"},
		{name: "KeyLogFile", key: KeyLogFile, expected: "logging.file", category: "logging"},

		// Symlink keys
		{name: "KeySymlinkMode", key: KeySymlinkMode, expected: "symlinks.mode", category: "symlinks"},
		{name: "KeySymlinkFolding", key: KeySymlinkFolding, expected: "symlinks.folding", category: "symlinks"},
		{name: "KeySymlinkOverwrite", key: KeySymlinkOverwrite, expected: "symlinks.overwrite", category: "symlinks"},
		{name: "KeySymlinkBackup", key: KeySymlinkBackup, expected: "symlinks.backup", category: "symlinks"},
		{name: "KeySymlinkBackupSuffix", key: KeySymlinkBackupSuffix, expected: "symlinks.backup_suffix", category: "symlinks"},
		{name: "KeySymlinkBackupDir", key: KeySymlinkBackupDir, expected: "symlinks.backup_dir", category: "symlinks"},

		// Ignore keys
		{name: "KeyIgnoreUseDefaults", key: KeyIgnoreUseDefaults, expected: "ignore.use_defaults", category: "ignore"},
		{name: "KeyIgnorePatterns", key: KeyIgnorePatterns, expected: "ignore.patterns", category: "ignore"},
		{name: "KeyIgnoreOverrides", key: KeyIgnoreOverrides, expected: "ignore.overrides", category: "ignore"},

		// Dotfile keys
		{name: "KeyDotfileTranslate", key: KeyDotfileTranslate, expected: "dotfile.translate", category: "dotfile"},
		{name: "KeyDotfilePrefix", key: KeyDotfilePrefix, expected: "dotfile.prefix", category: "dotfile"},

		// Output keys
		{name: "KeyOutputFormat", key: KeyOutputFormat, expected: "output.format", category: "output"},
		{name: "KeyOutputColor", key: KeyOutputColor, expected: "output.color", category: "output"},
		{name: "KeyOutputProgress", key: KeyOutputProgress, expected: "output.progress", category: "output"},
		{name: "KeyOutputVerbosity", key: KeyOutputVerbosity, expected: "output.verbosity", category: "output"},
		{name: "KeyOutputWidth", key: KeyOutputWidth, expected: "output.width", category: "output"},

		// Operations keys
		{name: "KeyOperationsDryRun", key: KeyOperationsDryRun, expected: "operations.dry_run", category: "operations"},
		{name: "KeyOperationsAtomic", key: KeyOperationsAtomic, expected: "operations.atomic", category: "operations"},
		{name: "KeyOperationsMaxParallel", key: KeyOperationsMaxParallel, expected: "operations.max_parallel", category: "operations"},

		// Packages keys
		{name: "KeyPackagesSortBy", key: KeyPackagesSortBy, expected: "packages.sort_by", category: "packages"},
		{name: "KeyPackagesAutoDiscover", key: KeyPackagesAutoDiscover, expected: "packages.auto_discover", category: "packages"},
		{name: "KeyPackagesValidateNames", key: KeyPackagesValidateNames, expected: "packages.validate_names", category: "packages"},

		// Doctor keys
		{name: "KeyDoctorAutoFix", key: KeyDoctorAutoFix, expected: "doctor.auto_fix", category: "doctor"},
		{name: "KeyDoctorCheckManifest", key: KeyDoctorCheckManifest, expected: "doctor.check_manifest", category: "doctor"},
		{name: "KeyDoctorCheckBrokenLinks", key: KeyDoctorCheckBrokenLinks, expected: "doctor.check_broken_links", category: "doctor"},
		{name: "KeyDoctorCheckOrphaned", key: KeyDoctorCheckOrphaned, expected: "doctor.check_orphaned", category: "doctor"},
		{name: "KeyDoctorOrphanScanMode", key: KeyDoctorOrphanScanMode, expected: "doctor.orphan_scan_mode", category: "doctor"},
		{name: "KeyDoctorOrphanScanDepth", key: KeyDoctorOrphanScanDepth, expected: "doctor.orphan_scan_depth", category: "doctor"},
		{name: "KeyDoctorOrphanSkipPatterns", key: KeyDoctorOrphanSkipPatterns, expected: "doctor.orphan_skip_patterns", category: "doctor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.key,
				"key constant %s should equal %s", tt.name, tt.expected)
		})
	}
}

func TestKeyFormatConsistency(t *testing.T) {
	keys := []string{
		KeyDirPackage, KeyDirTarget, KeyDirManifest,
		KeyLogLevel, KeyLogFormat, KeyLogDestination, KeyLogFile,
		KeySymlinkMode, KeySymlinkFolding, KeySymlinkOverwrite, KeySymlinkBackup,
		KeySymlinkBackupSuffix, KeySymlinkBackupDir,
		KeyIgnoreUseDefaults, KeyIgnorePatterns, KeyIgnoreOverrides,
		KeyDotfileTranslate, KeyDotfilePrefix,
		KeyOutputFormat, KeyOutputColor, KeyOutputProgress, KeyOutputVerbosity, KeyOutputWidth,
		KeyOperationsDryRun, KeyOperationsAtomic, KeyOperationsMaxParallel,
		KeyPackagesSortBy, KeyPackagesAutoDiscover, KeyPackagesValidateNames,
		KeyDoctorAutoFix, KeyDoctorCheckManifest, KeyDoctorCheckBrokenLinks,
		KeyDoctorCheckOrphaned, KeyDoctorOrphanScanMode, KeyDoctorOrphanScanDepth,
		KeyDoctorOrphanSkipPatterns,
	}

	for _, key := range keys {
		t.Run("key format: "+key, func(t *testing.T) {
			// All keys should contain exactly one dot separator
			assert.Equal(t, 1, strings.Count(key, "."),
				"key %s should have exactly one dot separator", key)

			// Keys should not start or end with dot
			assert.False(t, strings.HasPrefix(key, "."),
				"key %s should not start with dot", key)
			assert.False(t, strings.HasSuffix(key, "."),
				"key %s should not end with dot", key)

			// Keys should be lowercase with underscores
			assert.Equal(t, strings.ToLower(key), key,
				"key %s should be lowercase", key)
			assert.NotContains(t, key, "-",
				"key %s should use underscores not hyphens", key)
		})
	}
}

func TestKeyCategoryGrouping(t *testing.T) {
	categories := map[string][]string{
		"directories": {KeyDirPackage, KeyDirTarget, KeyDirManifest},
		"logging":     {KeyLogLevel, KeyLogFormat, KeyLogDestination, KeyLogFile},
		"symlinks":    {KeySymlinkMode, KeySymlinkFolding, KeySymlinkOverwrite, KeySymlinkBackup, KeySymlinkBackupSuffix, KeySymlinkBackupDir},
		"ignore":      {KeyIgnoreUseDefaults, KeyIgnorePatterns, KeyIgnoreOverrides},
		"dotfile":     {KeyDotfileTranslate, KeyDotfilePrefix},
		"output":      {KeyOutputFormat, KeyOutputColor, KeyOutputProgress, KeyOutputVerbosity, KeyOutputWidth},
		"operations":  {KeyOperationsDryRun, KeyOperationsAtomic, KeyOperationsMaxParallel},
		"packages":    {KeyPackagesSortBy, KeyPackagesAutoDiscover, KeyPackagesValidateNames},
		"doctor":      {KeyDoctorAutoFix, KeyDoctorCheckManifest, KeyDoctorCheckBrokenLinks, KeyDoctorCheckOrphaned, KeyDoctorOrphanScanMode, KeyDoctorOrphanScanDepth, KeyDoctorOrphanSkipPatterns},
	}

	for category, keys := range categories {
		t.Run("category: "+category, func(t *testing.T) {
			for _, key := range keys {
				assert.True(t, strings.HasPrefix(key, category+"."),
					"key %s should start with category prefix %s", key, category)
			}
		})
	}
}

func TestKeyUniqueness(t *testing.T) {
	keys := []string{
		KeyDirPackage, KeyDirTarget, KeyDirManifest,
		KeyLogLevel, KeyLogFormat, KeyLogDestination, KeyLogFile,
		KeySymlinkMode, KeySymlinkFolding, KeySymlinkOverwrite, KeySymlinkBackup,
		KeySymlinkBackupSuffix, KeySymlinkBackupDir,
		KeyIgnoreUseDefaults, KeyIgnorePatterns, KeyIgnoreOverrides,
		KeyDotfileTranslate, KeyDotfilePrefix,
		KeyOutputFormat, KeyOutputColor, KeyOutputProgress, KeyOutputVerbosity, KeyOutputWidth,
		KeyOperationsDryRun, KeyOperationsAtomic, KeyOperationsMaxParallel,
		KeyPackagesSortBy, KeyPackagesAutoDiscover, KeyPackagesValidateNames,
		KeyDoctorAutoFix, KeyDoctorCheckManifest, KeyDoctorCheckBrokenLinks,
		KeyDoctorCheckOrphaned, KeyDoctorOrphanScanMode, KeyDoctorOrphanScanDepth,
		KeyDoctorOrphanSkipPatterns,
	}

	seen := make(map[string]bool)
	for _, key := range keys {
		assert.False(t, seen[key], "key %s appears multiple times", key)
		seen[key] = true
	}

	assert.Equal(t, len(keys), len(seen),
		"all keys should be unique")
}
