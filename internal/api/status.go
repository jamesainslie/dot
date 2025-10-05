package api

import (
	"context"

	"github.com/jamesainslie/dot/pkg/dot"
)

// Status reports the current installation state for packages.
func (c *client) Status(ctx context.Context, packages ...string) (dot.Status, error) {
	targetPathResult := dot.NewTargetPath(c.config.TargetDir)
	if !targetPathResult.IsOk() {
		return dot.Status{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	manifestResult := c.manifest.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		// No manifest means nothing installed
		return dot.Status{Packages: []dot.PackageInfo{}}, nil
	}
	m := manifestResult.Unwrap()

	// Filter to requested packages if specified
	var pkgInfos []dot.PackageInfo
	if len(packages) == 0 {
		// Return all packages
		for _, info := range m.Packages {
			pkgInfos = append(pkgInfos, dot.PackageInfo{
				Name:        info.Name,
				InstalledAt: info.InstalledAt,
				LinkCount:   info.LinkCount,
				Links:       info.Links,
			})
		}
	} else {
		// Return only specified packages
		for _, pkg := range packages {
			if info, exists := m.Packages[pkg]; exists {
				pkgInfos = append(pkgInfos, dot.PackageInfo{
					Name:        info.Name,
					InstalledAt: info.InstalledAt,
					LinkCount:   info.LinkCount,
					Links:       info.Links,
				})
			}
		}
	}

	return dot.Status{
		Packages: pkgInfos,
	}, nil
}

// List returns all installed packages from the manifest.
func (c *client) List(ctx context.Context) ([]dot.PackageInfo, error) {
	status, err := c.Status(ctx)
	if err != nil {
		return nil, err
	}
	return status.Packages, nil
}
