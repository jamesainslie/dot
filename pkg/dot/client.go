package dot

import "context"

// Client provides the high-level API for dot operations.
//
// This interface abstracts the internal pipeline and executor orchestration,
// providing a simple facade for library consumers.
//
// The implementation resides in internal/api to avoid import cycles between
// pkg/dot (which contains domain types) and internal packages (which import
// those domain types).
//
// All operations are safe for concurrent use from multiple goroutines.
type Client interface {
	// Manage installs the specified packages by creating symlinks.
	// If DryRun is enabled in Config, shows the plan without executing.
	//
	// Returns an error if:
	//   - Package does not exist in PackageDir
	//   - Conflicts are detected and policy is PolicyFail
	//   - Filesystem operations fail
	//   - Context is cancelled
	Manage(ctx context.Context, packages ...string) error

	// PlanManage computes the execution plan for managing packages
	// without applying changes. Useful for dry-run preview or validation.
	PlanManage(ctx context.Context, packages ...string) (Plan, error)

	// Unmanage removes the specified packages by deleting their symlinks.
	// Only removes links that point to the PackageDir. Never touches user files.
	//
	// Returns an error if:
	//   - Package is not currently installed
	//   - Filesystem operations fail
	//   - Context is cancelled
	Unmanage(ctx context.Context, packages ...string) error

	// PlanUnmanage computes the execution plan for unmanaging packages.
	PlanUnmanage(ctx context.Context, packages ...string) (Plan, error)

	// Remanage reinstalls packages by unmanaging then managing.
	// Uses incremental planning to skip unchanged packages when possible.
	//
	// Returns an error if:
	//   - Package does not exist in PackageDir
	//   - Conflicts are detected
	//   - Filesystem operations fail
	//   - Context is cancelled
	Remanage(ctx context.Context, packages ...string) error

	// PlanRemanage computes the execution plan for remanaging packages.
	// Uses manifest and content hashing to detect changes.
	PlanRemanage(ctx context.Context, packages ...string) (Plan, error)

	// Adopt moves existing files from TargetDir into a package then creates symlinks.
	// The files are moved to the package directory and replaced with symlinks.
	//
	// Returns an error if:
	//   - Files do not exist in TargetDir
	//   - Package does not exist in PackageDir
	//   - Filesystem operations fail
	//   - Context is cancelled
	Adopt(ctx context.Context, files []string, pkg string) error

	// PlanAdopt computes the execution plan for adopting files.
	PlanAdopt(ctx context.Context, files []string, pkg string) (Plan, error)

	// Status reports the current installation state for the specified packages.
	// If no packages are specified, reports status for all installed packages.
	//
	// Returns PackageInfo for each package including:
	//   - Installation timestamp
	//   - Number of links created
	//   - List of link paths
	Status(ctx context.Context, packages ...string) (Status, error)

	// List returns information about all installed packages.
	// Equivalent to Status() with no package filter.
	List(ctx context.Context) ([]PackageInfo, error)

	// Doctor performs health checks on the installation.
	// Detects broken links, orphaned links, permission issues, and inconsistencies.
	//
	// The scanCfg parameter controls orphaned link detection behavior.
	// Use DefaultScanConfig() for no orphan scanning (backward compatible),
	// ScopedScanConfig() for smart scanning, or DeepScanConfig(depth) for full scan.
	//
	// Returns a DiagnosticReport with:
	//   - Overall health status
	//   - List of issues found
	//   - Summary statistics
	//
	// TODO: BREAKING CHANGE in v0.2.0 - Added scanCfg parameter.
	// Consider adding transitional path in v0.3.0:
	//   - Add DoctorWithScan(ctx, scanCfg) method
	//   - Deprecate Doctor() and make it call DoctorWithScan(ctx, DefaultScanConfig())
	//   - Remove deprecated method in v1.0.0
	// This would allow gradual migration for library consumers.
	Doctor(ctx context.Context, scanCfg ScanConfig) (DiagnosticReport, error)

	// Config returns a copy of the client's configuration.
	Config() Config
}

// NewClient creates a new Client with the given configuration.
//
// Returns an error if:
//   - Configuration is invalid (see Config.Validate)
//   - Required dependencies are missing (FS, Logger)
//
// The returned Client is safe for concurrent use from multiple goroutines.
func NewClient(cfg Config) (Client, error) {
	if newClientImpl == nil {
		panic("Client implementation not registered. This indicates internal/api package " +
			"was not imported to trigger init() registration. " +
			"Import path: github.com/jamesainslie/dot/internal/api")
	}
	return newClientImpl(cfg)
}

// newClientImpl holds the actual constructor function.
// This is set by internal/api during package initialization.
var newClientImpl func(Config) (Client, error)

// RegisterClientImpl registers the Client implementation.
// This function should only be called by internal/api during init().
func RegisterClientImpl(fn func(Config) (Client, error)) {
	newClientImpl = fn
}

// GetClientImpl returns the current client implementation function.
// This is primarily for testing purposes.
func GetClientImpl() func(Config) (Client, error) {
	return newClientImpl
}
