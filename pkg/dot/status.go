package dot

import "time"

// Status represents the installation state of packages.
type Status struct {
	Packages []PackageInfo
}

// PackageInfo contains metadata about an installed package.
type PackageInfo struct {
	Name        string
	InstalledAt time.Time
	LinkCount   int
	Links       []string
}

