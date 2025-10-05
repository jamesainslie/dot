package dot

import "time"

// Status represents the installation state of packages.
type Status struct {
	Packages []PackageInfo `json:"packages" yaml:"packages"`
}

// PackageInfo contains metadata about an installed package.
type PackageInfo struct {
	Name        string    `json:"name" yaml:"name"`
	InstalledAt time.Time `json:"installed_at" yaml:"installed_at"`
	LinkCount   int       `json:"link_count" yaml:"link_count"`
	Links       []string  `json:"links" yaml:"links"`
}
