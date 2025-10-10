package bootstrap

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

// FS defines filesystem operations required for loading bootstrap config.
type FS interface {
	ReadFile(ctx context.Context, path string) ([]byte, error)
}

// Load reads and parses a bootstrap configuration file.
//
// Returns an error if:
//   - File cannot be read
//   - YAML syntax is invalid
//   - Configuration validation fails
//
// The configuration is automatically validated after loading.
func Load(ctx context.Context, fs FS, path string) (Config, error) {
	// Read file
	data, err := fs.ReadFile(ctx, path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse YAML: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// FilterPackagesByPlatform returns packages compatible with the specified platform.
//
// Packages with no platform restrictions are included for all platforms.
// Packages with platform restrictions are included only if the platform matches.
func FilterPackagesByPlatform(packages []PackageSpec, platform string) []PackageSpec {
	filtered := make([]PackageSpec, 0, len(packages))

	for _, pkg := range packages {
		// No platform restriction - include for all platforms
		if len(pkg.Platform) == 0 {
			filtered = append(filtered, pkg)
			continue
		}

		// Check if platform matches
		for _, p := range pkg.Platform {
			if p == platform {
				filtered = append(filtered, pkg)
				break
			}
		}
	}

	return filtered
}

// GetPackageNames extracts package names from configuration.
func GetPackageNames(cfg Config) []string {
	names := make([]string, 0, len(cfg.Packages))
	for _, pkg := range cfg.Packages {
		names = append(names, pkg.Name)
	}
	return names
}

// GetProfile retrieves packages for a named profile.
//
// Returns an error if the profile does not exist.
func GetProfile(cfg Config, profileName string) ([]string, error) {
	profile, exists := cfg.Profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("profile not found: %s", profileName)
	}
	return profile.Packages, nil
}
