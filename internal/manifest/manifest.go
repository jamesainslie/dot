package manifest

import "time"

// Manifest tracks installed package state
type Manifest struct {
	Version   string                 `json:"version"`
	UpdatedAt time.Time              `json:"updated_at"`
	Packages  map[string]PackageInfo `json:"packages"`
	Hashes    map[string]string      `json:"hashes"`
}

// PackageInfo contains installation metadata for a package
type PackageInfo struct {
	Name        string    `json:"name"`
	InstalledAt time.Time `json:"installed_at"`
	LinkCount   int       `json:"link_count"`
	Links       []string  `json:"links"`
}

// New creates a new empty manifest
func New() Manifest {
	return Manifest{
		Version:   "1.0",
		UpdatedAt: time.Now(),
		Packages:  make(map[string]PackageInfo),
		Hashes:    make(map[string]string),
	}
}

// AddPackage adds or updates package information
func (m *Manifest) AddPackage(pkg PackageInfo) {
	m.Packages[pkg.Name] = pkg
	m.UpdatedAt = time.Now()
}

// RemovePackage removes package from manifest
func (m *Manifest) RemovePackage(name string) bool {
	if _, exists := m.Packages[name]; !exists {
		return false
	}
	delete(m.Packages, name)
	delete(m.Hashes, name)
	m.UpdatedAt = time.Now()
	return true
}

// GetPackage retrieves package information
func (m *Manifest) GetPackage(name string) (PackageInfo, bool) {
	pkg, exists := m.Packages[name]
	return pkg, exists
}

// SetHash updates content hash for package
func (m *Manifest) SetHash(name, hash string) {
	m.Hashes[name] = hash
	m.UpdatedAt = time.Now()
}

// GetHash retrieves content hash for package
func (m *Manifest) GetHash(name string) (string, bool) {
	hash, exists := m.Hashes[name]
	return hash, exists
}

// PackageList returns all packages as slice
func (m *Manifest) PackageList() []PackageInfo {
	packages := make([]PackageInfo, 0, len(m.Packages))
	for _, pkg := range m.Packages {
		packages = append(packages, pkg)
	}
	return packages
}
