package config

// Default configuration values for consistent initialization.
// These constants define sensible, safe defaults for all configuration options.

const (
	// Logging defaults
	DefaultLogLevel       = "INFO"   // Default log level (DEBUG, INFO, WARN, ERROR)
	DefaultLogFormat      = "text"   // Default log format (text, json)
	DefaultLogDestination = "stderr" // Default log destination (stderr, stdout, file)

	// Symlink defaults
	DefaultSymlinkMode         = "relative" // Default symlink mode (relative, absolute)
	DefaultSymlinkFolding      = true       // Enable directory folding optimization
	DefaultSymlinkOverwrite    = false      // Do not overwrite existing files (safe default)
	DefaultSymlinkBackup       = false      // Do not create backups (explicit opt-in)
	DefaultSymlinkBackupSuffix = ".bak"     // Default backup file suffix

	// Dotfile translation defaults
	DefaultDotfileTranslate = true   // Enable dot- to . translation
	DefaultDotfilePrefix    = "dot-" // Prefix for dotfile translation

	// Output defaults
	DefaultOutputFormat    = "text" // Default output format (text, json, yaml, table)
	DefaultOutputColor     = "auto" // Default color mode (auto, always, never)
	DefaultOutputProgress  = true   // Show progress indicators
	DefaultOutputVerbosity = 1      // Default verbosity (0=quiet, 1=normal, 2=verbose, 3=debug)
	DefaultOutputWidth     = 0      // Terminal width (0 = auto-detect)

	// Operations defaults
	DefaultOperationsDryRun      = false // Execute operations (not dry-run)
	DefaultOperationsAtomic      = true  // Enable atomic operations with rollback
	DefaultOperationsMaxParallel = 0     // Max parallel operations (0 = auto-detect CPU count)

	// Packages defaults
	DefaultPackagesSortBy        = "name" // Default sort order (name, links, date)
	DefaultPackagesAutoDiscover  = false  // Do not auto-discover packages
	DefaultPackagesValidateNames = true   // Validate package naming conventions

	// Doctor defaults
	DefaultDoctorAutoFix          = false // Do not auto-fix issues (require explicit action)
	DefaultDoctorCheckManifest    = true  // Check manifest integrity
	DefaultDoctorCheckBrokenLinks = true  // Check for broken symlinks
	DefaultDoctorCheckOrphaned    = false // Do not check for orphaned links (opt-in)
	DefaultDoctorOrphanScanMode   = "off" // Orphan scan mode (off, scoped, deep)
	DefaultDoctorOrphanScanDepth  = 0     // Orphan scan depth (0 = unlimited)
	DefaultDoctorCheckPermissions = false // Do not check file permissions

	// Ignore defaults
	DefaultIgnoreUseDefaults = true // Use default ignore patterns

	// Experimental defaults
	DefaultExperimentalParallel  = false // Experimental parallel operations disabled
	DefaultExperimentalProfiling = false // Performance profiling disabled
)
