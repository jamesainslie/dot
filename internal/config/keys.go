package config

// Configuration key constants for consistent configuration access.
// Keys follow the format "category.field" using dot notation.
// All keys use lowercase with underscores for multi-word fields.

const (
	// Directory configuration keys
	KeyDirPackage  = "directories.package"
	KeyDirTarget   = "directories.target"
	KeyDirManifest = "directories.manifest"

	// Logging configuration keys
	KeyLogLevel       = "logging.level"
	KeyLogFormat      = "logging.format"
	KeyLogDestination = "logging.destination"
	KeyLogFile        = "logging.file"

	// Symlink configuration keys
	KeySymlinkMode         = "symlinks.mode"
	KeySymlinkFolding      = "symlinks.folding"
	KeySymlinkOverwrite    = "symlinks.overwrite"
	KeySymlinkBackup       = "symlinks.backup"
	KeySymlinkBackupSuffix = "symlinks.backup_suffix"
	KeySymlinkBackupDir    = "symlinks.backup_dir"

	// Ignore pattern configuration keys
	KeyIgnoreUseDefaults = "ignore.use_defaults"
	KeyIgnorePatterns    = "ignore.patterns"
	KeyIgnoreOverrides   = "ignore.overrides"

	// Dotfile translation configuration keys
	KeyDotfileTranslate = "dotfile.translate"
	KeyDotfilePrefix    = "dotfile.prefix"

	// Output configuration keys
	KeyOutputFormat    = "output.format"
	KeyOutputColor     = "output.color"
	KeyOutputProgress  = "output.progress"
	KeyOutputVerbosity = "output.verbosity"
	KeyOutputWidth     = "output.width"

	// Operations configuration keys
	KeyOperationsDryRun      = "operations.dry_run"
	KeyOperationsAtomic      = "operations.atomic"
	KeyOperationsMaxParallel = "operations.max_parallel"

	// Packages configuration keys
	KeyPackagesSortBy        = "packages.sort_by"
	KeyPackagesAutoDiscover  = "packages.auto_discover"
	KeyPackagesValidateNames = "packages.validate_names"

	// Doctor configuration keys
	KeyDoctorAutoFix            = "doctor.auto_fix"
	KeyDoctorCheckManifest      = "doctor.check_manifest"
	KeyDoctorCheckBrokenLinks   = "doctor.check_broken_links"
	KeyDoctorCheckOrphaned      = "doctor.check_orphaned"
	KeyDoctorOrphanScanMode     = "doctor.orphan_scan_mode"
	KeyDoctorOrphanScanDepth    = "doctor.orphan_scan_depth"
	KeyDoctorOrphanSkipPatterns = "doctor.orphan_skip_patterns"
)
