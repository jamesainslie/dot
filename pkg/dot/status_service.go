package dot

import (
	"context"
)

// StatusService handles status and listing operations.
type StatusService struct {
	manifestSvc *ManifestService
	targetDir   string
}

// newStatusService creates a new status service.
func newStatusService(manifestSvc *ManifestService, targetDir string) *StatusService {
	return &StatusService{
		manifestSvc: manifestSvc,
		targetDir:   targetDir,
	}
}

// Status reports the current installation state for packages.
func (s *StatusService) Status(ctx context.Context, packages ...string) (Status, error) {
	targetPathResult := NewTargetPath(s.targetDir)
	if !targetPathResult.IsOk() {
		return Status{}, targetPathResult.UnwrapErr()
	}
	targetPath := targetPathResult.Unwrap()

	// Load manifest
	manifestResult := s.manifestSvc.Load(ctx, targetPath)
	if !manifestResult.IsOk() {
		// No manifest means nothing installed
		return Status{Packages: []PackageInfo{}}, nil
	}

	m := manifestResult.Unwrap()

	// Filter to requested packages if specified
	pkgInfos := make([]PackageInfo, 0)
	if len(packages) == 0 {
		// Return all packages
		for _, info := range m.Packages {
			pkgInfos = append(pkgInfos, PackageInfo{
				Name:        info.Name,
				InstalledAt: info.InstalledAt,
				LinkCount:   info.LinkCount,
				Links:       info.Links,
			})
		}
	} else {
		// Return only specified packages
		for _, pkg := range packages {
			if info, exists := m.GetPackage(pkg); exists {
				pkgInfos = append(pkgInfos, PackageInfo{
					Name:        info.Name,
					InstalledAt: info.InstalledAt,
					LinkCount:   info.LinkCount,
					Links:       info.Links,
				})
			}
		}
	}
	return Status{
		Packages: pkgInfos,
	}, nil
}

// List returns all installed packages from the manifest.
func (s *StatusService) List(ctx context.Context) ([]PackageInfo, error) {
	status, err := s.Status(ctx)
	if err != nil {
		return nil, err
	}
	return status.Packages, nil
}
