# Phase 11: Manifest and State Management - Implementation Plan

## Overview

Phase 11 implements persistent state tracking for incremental operations. The manifest system maintains a record of installed packages, their content hashes, and link inventory in the target directory. This enables fast incremental restow operations and drift detection.

## Design Principles

- **Pure Domain Types**: Manifest types are pure value objects with no I/O dependencies
- **Port/Adapter Pattern**: ManifestStore is a port with filesystem adapter implementation
- **Fast Hashing**: Use efficient non-cryptographic hash for content detection
- **Graceful Degradation**: Missing or corrupt manifests don't break operations
- **Idempotent Persistence**: Saving manifests is safe to repeat
- **Atomic Updates**: Manifest writes are atomic via temp file and rename

## Dependencies

### Completed Prerequisites
- Phase 1: Domain model with Path types
- Phase 2: Infrastructure ports (FS interface)
- Phase 3: Adapters (OSFilesystem, MemoryFilesystem)

### External Dependencies
- `encoding/json`: Manifest serialization
- `crypto/sha256` or fast hash library (xxhash): Content hashing
- `time`: Timestamp tracking
- `io/fs`: Filesystem traversal for hashing

## Package Structure

```
internal/
├── manifest/
│   ├── manifest.go          # Core types (Manifest, PackageInfo)
│   ├── manifest_test.go
│   ├── store.go             # ManifestStore interface
│   ├── store_test.go
│   ├── fsstore.go           # FSManifestStore implementation
│   ├── fsstore_test.go
│   ├── hash.go              # ContentHasher implementation
│   ├── hash_test.go
│   ├── validate.go          # Validation and drift detection
│   └── validate_test.go
```

## Implementation Tasks

---

### Task 11.1.1: Define Core Manifest Types

**Objective**: Create pure domain types for manifest representation.

**Test First** (`internal/manifest/manifest_test.go`):
```go
func TestManifest_New(t *testing.T) {
    m := manifest.New()
    
    assert.Equal(t, "1.0", m.Version)
    assert.NotNil(t, m.Packages)
    assert.NotNil(t, m.Hashes)
    assert.False(t, m.UpdatedAt.IsZero())
}

func TestManifest_AddPackage(t *testing.T) {
    m := manifest.New()
    
    pkg := manifest.PackageInfo{
        Name:        "vim",
        InstalledAt: time.Now(),
        LinkCount:   5,
        Links:       []string{".vimrc", ".vim/colors"},
    }
    
    m.AddPackage(pkg)
    
    retrieved, exists := m.GetPackage("vim")
    assert.True(t, exists)
    assert.Equal(t, "vim", retrieved.Name)
    assert.Equal(t, 5, retrieved.LinkCount)
}

func TestManifest_RemovePackage(t *testing.T) {
    m := manifest.New()
    m.AddPackage(manifest.PackageInfo{Name: "vim"})
    
    removed := m.RemovePackage("vim")
    
    assert.True(t, removed)
    _, exists := m.GetPackage("vim")
    assert.False(t, exists)
}

func TestManifest_SetHash(t *testing.T) {
    m := manifest.New()
    
    m.SetHash("vim", "abc123")
    
    hash, exists := m.GetHash("vim")
    assert.True(t, exists)
    assert.Equal(t, "abc123", hash)
}

func TestManifest_PackageList(t *testing.T) {
    m := manifest.New()
    m.AddPackage(manifest.PackageInfo{Name: "vim"})
    m.AddPackage(manifest.PackageInfo{Name: "zsh"})
    
    packages := m.PackageList()
    
    assert.Len(t, packages, 2)
    names := []string{packages[0].Name, packages[1].Name}
    assert.Contains(t, names, "vim")
    assert.Contains(t, names, "zsh")
}
```

**Implementation** (`internal/manifest/manifest.go`):
```go
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
```

**Commit Message**:
```
feat(manifest): add core manifest domain types

Implement Manifest and PackageInfo types for tracking installed package
state. These pure value objects maintain package metadata, link inventory,
and content hashes for incremental change detection.

- Add Manifest type with version, packages, and hashes
- Implement PackageInfo with installation metadata
- Add immutable operations for package management
- Add hash tracking for content-based change detection
- Include timestamp tracking for audit trails

Refs: Phase 11.1
```

---

### Task 11.1.2: Define ManifestStore Interface

**Objective**: Create port interface for manifest persistence.

**Test First** (`internal/manifest/store_test.go`):
```go
func TestManifestStore_Interface(t *testing.T) {
    // Verify interface is implemented by mock
    var _ manifest.ManifestStore = (*mockManifestStore)(nil)
}

type mockManifestStore struct {
    loadFn func(context.Context, dot.TargetPath) dot.Result[manifest.Manifest]
    saveFn func(context.Context, dot.TargetPath, manifest.Manifest) error
}

func (m *mockManifestStore) Load(ctx context.Context, target dot.TargetPath) dot.Result[manifest.Manifest] {
    return m.loadFn(ctx, target)
}

func (m *mockManifestStore) Save(ctx context.Context, target dot.TargetPath, manifest manifest.Manifest) error {
    return m.saveFn(ctx, target, manifest)
}
```

**Implementation** (`internal/manifest/store.go`):
```go
package manifest

import (
    "context"
    
    "github.com/yourusername/dot/pkg/dot"
)

// ManifestStore provides persistence for manifests
type ManifestStore interface {
    // Load retrieves manifest from target directory
    // Returns empty manifest if file doesn't exist
    Load(ctx context.Context, targetDir dot.TargetPath) dot.Result[Manifest]
    
    // Save persists manifest to target directory
    // Write is atomic via temp file and rename
    Save(ctx context.Context, targetDir dot.TargetPath, manifest Manifest) error
}
```

**Commit Message**:
```
feat(manifest): define ManifestStore interface

Add port interface for manifest persistence operations. Interface defines
Load and Save operations with context support for cancellation.

- Add ManifestStore interface in manifest package
- Load returns Result monad for error handling
- Save performs atomic writes
- Operations accept TargetPath for type safety

Refs: Phase 11.1
```

---

### Task 11.2.1: Implement FSManifestStore Load

**Objective**: Implement filesystem-based manifest loading with graceful missing file handling.

**Test First** (`internal/manifest/fsstore_test.go`):
```go
func TestFSManifestStore_Load_MissingFile(t *testing.T) {
    fs := memfs.New()
    store := manifest.NewFSManifestStore(fs)
    targetDir := mustTargetPath(t, "/home/user")
    
    result := store.Load(context.Background(), targetDir)
    
    require.True(t, result.IsOk())
    m, _ := result.Unwrap()
    assert.Equal(t, "1.0", m.Version)
    assert.Empty(t, m.Packages)
}

func TestFSManifestStore_Load_ValidManifest(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    
    // Create manifest file
    manifestData := `{
        "version": "1.0",
        "updated_at": "2024-01-15T10:30:00Z",
        "packages": {
            "vim": {
                "name": "vim",
                "installed_at": "2024-01-15T10:00:00Z",
                "link_count": 2,
                "links": [".vimrc", ".vim/colors"]
            }
        },
        "hashes": {
            "vim": "abc123"
        }
    }`
    manifestPath := filepath.Join(targetDir.String(), ".dot-manifest.json")
    require.NoError(t, fs.WriteFile(context.Background(), manifestPath, []byte(manifestData), 0644))
    
    store := manifest.NewFSManifestStore(fs)
    result := store.Load(context.Background(), targetDir)
    
    require.True(t, result.IsOk())
    m, _ := result.Unwrap()
    assert.Equal(t, "1.0", m.Version)
    assert.Len(t, m.Packages, 1)
    
    vim, exists := m.GetPackage("vim")
    assert.True(t, exists)
    assert.Equal(t, 2, vim.LinkCount)
}

func TestFSManifestStore_Load_CorruptManifest(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    
    // Write invalid JSON
    manifestPath := filepath.Join(targetDir.String(), ".dot-manifest.json")
    require.NoError(t, fs.WriteFile(context.Background(), manifestPath, []byte("invalid json"), 0644))
    
    store := manifest.NewFSManifestStore(fs)
    result := store.Load(context.Background(), targetDir)
    
    assert.False(t, result.IsOk())
    _, err := result.Unwrap()
    assert.Error(t, err)
}

func TestFSManifestStore_Load_WithContext(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    store := manifest.NewFSManifestStore(fs)
    
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    
    result := store.Load(ctx, targetDir)
    
    assert.False(t, result.IsOk())
}
```

**Implementation** (`internal/manifest/fsstore.go`):
```go
package manifest

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/yourusername/dot/pkg/dot"
)

const manifestFileName = ".dot-manifest.json"

// FSManifestStore implements ManifestStore using filesystem
type FSManifestStore struct {
    fs dot.FS
}

// NewFSManifestStore creates filesystem-based manifest store
func NewFSManifestStore(fs dot.FS) *FSManifestStore {
    return &FSManifestStore{fs: fs}
}

// Load retrieves manifest from target directory
func (s *FSManifestStore) Load(ctx context.Context, targetDir dot.TargetPath) dot.Result[Manifest] {
    if ctx.Err() != nil {
        return dot.Err[Manifest](ctx.Err())
    }
    
    manifestPath := filepath.Join(targetDir.String(), manifestFileName)
    
    data, err := s.fs.ReadFile(ctx, manifestPath)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            // Missing manifest is not an error - return empty manifest
            return dot.Ok(New())
        }
        return dot.Err[Manifest](fmt.Errorf("failed to read manifest: %w", err))
    }
    
    var m Manifest
    if err := json.Unmarshal(data, &m); err != nil {
        return dot.Err[Manifest](fmt.Errorf("failed to parse manifest: %w", err))
    }
    
    return dot.Ok(m)
}

// Save persists manifest to target directory
func (s *FSManifestStore) Save(ctx context.Context, targetDir dot.TargetPath, manifest Manifest) error {
    // Implementation in next task
    return nil
}
```

**Commit Message**:
```
feat(manifest): implement FSManifestStore Load operation

Add filesystem-based manifest loading with graceful handling of missing
files. Returns empty manifest when file doesn't exist rather than failing,
enabling new installations to work seamlessly.

- Implement NewFSManifestStore constructor
- Add Load operation with JSON deserialization
- Handle missing manifest gracefully
- Return errors for corrupt manifests
- Support context cancellation

Refs: Phase 11.2
```

---

### Task 11.2.2: Implement FSManifestStore Save

**Objective**: Implement atomic manifest persistence via temp file and rename.

**Test First** (`internal/manifest/fsstore_test.go`):
```go
func TestFSManifestStore_Save_NewManifest(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    
    store := manifest.NewFSManifestStore(fs)
    m := manifest.New()
    m.AddPackage(manifest.PackageInfo{
        Name:      "vim",
        LinkCount: 2,
        Links:     []string{".vimrc", ".vim/colors"},
    })
    
    err := store.Save(context.Background(), targetDir, m)
    
    require.NoError(t, err)
    
    // Verify file exists and is readable
    manifestPath := filepath.Join(targetDir.String(), ".dot-manifest.json")
    exists := fs.Exists(context.Background(), manifestPath)
    assert.True(t, exists)
    
    // Verify content
    result := store.Load(context.Background(), targetDir)
    require.True(t, result.IsOk())
    loaded, _ := result.Unwrap()
    assert.Len(t, loaded.Packages, 1)
}

func TestFSManifestStore_Save_UpdatesTimestamp(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    
    store := manifest.NewFSManifestStore(fs)
    m := manifest.New()
    originalTime := m.UpdatedAt
    
    time.Sleep(10 * time.Millisecond)
    err := store.Save(context.Background(), targetDir, m)
    require.NoError(t, err)
    
    result := store.Load(context.Background(), targetDir)
    require.True(t, result.IsOk())
    loaded, _ := result.Unwrap()
    assert.True(t, loaded.UpdatedAt.After(originalTime))
}

func TestFSManifestStore_Save_AtomicWrite(t *testing.T) {
    // This test verifies atomic write by checking temp file is cleaned up
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    
    store := manifest.NewFSManifestStore(fs)
    m := manifest.New()
    
    err := store.Save(context.Background(), targetDir, m)
    require.NoError(t, err)
    
    // Verify no temp files left behind
    entries, err := fs.ReadDir(context.Background(), targetDir.String())
    require.NoError(t, err)
    
    for _, entry := range entries {
        assert.NotContains(t, entry.Name(), ".tmp")
        assert.NotContains(t, entry.Name(), "~")
    }
}

func TestFSManifestStore_Save_WithContext(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    
    store := manifest.NewFSManifestStore(fs)
    m := manifest.New()
    
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    
    err := store.Save(ctx, targetDir, m)
    
    assert.Error(t, err)
}

func TestFSManifestStore_Save_PermissionDenied(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0555)) // Read-only
    
    store := manifest.NewFSManifestStore(fs)
    m := manifest.New()
    
    err := store.Save(context.Background(), targetDir, m)
    
    assert.Error(t, err)
}
```

**Implementation** (add to `internal/manifest/fsstore.go`):
```go
func (s *FSManifestStore) Save(ctx context.Context, targetDir dot.TargetPath, manifest Manifest) error {
    if ctx.Err() != nil {
        return ctx.Err()
    }
    
    // Update timestamp
    manifest.UpdatedAt = time.Now()
    
    // Marshal to JSON with indentation
    data, err := json.MarshalIndent(manifest, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal manifest: %w", err)
    }
    
    manifestPath := filepath.Join(targetDir.String(), manifestFileName)
    
    // Atomic write via temp file and rename
    tempPath := manifestPath + ".tmp"
    
    // Write to temp file
    if err := s.fs.WriteFile(ctx, tempPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write temp manifest: %w", err)
    }
    
    // Atomic rename
    if err := s.fs.Rename(ctx, tempPath, manifestPath); err != nil {
        // Clean up temp file on failure
        _ = s.fs.Remove(ctx, tempPath)
        return fmt.Errorf("failed to rename manifest: %w", err)
    }
    
    return nil
}
```

**Commit Message**:
```
feat(manifest): implement FSManifestStore Save operation

Add atomic manifest persistence using temp file and rename strategy.
Ensures manifest writes are atomic and never leave corrupt files.

- Implement Save with JSON serialization
- Use temp file and rename for atomic writes
- Update timestamp on save
- Clean up temp files on failure
- Add permission error handling

Refs: Phase 11.2
```

---

### Task 11.3.1: Implement ContentHasher

**Objective**: Fast content hashing for package change detection.

**Test First** (`internal/manifest/hash_test.go`):
```go
func TestContentHasher_HashPackage_EmptyPackage(t *testing.T) {
    fs := memfs.New()
    pkgPath := mustPackagePath(t, "/stow/vim")
    require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))
    
    hasher := manifest.NewContentHasher(fs)
    
    hash, err := hasher.HashPackage(context.Background(), pkgPath)
    
    require.NoError(t, err)
    assert.NotEmpty(t, hash)
}

func TestContentHasher_HashPackage_SingleFile(t *testing.T) {
    fs := memfs.New()
    pkgPath := mustPackagePath(t, "/stow/vim")
    require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))
    
    vimrcPath := filepath.Join(pkgPath.String(), "dot-vimrc")
    require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("set number\n"), 0644))
    
    hasher := manifest.NewContentHasher(fs)
    
    hash, err := hasher.HashPackage(context.Background(), pkgPath)
    
    require.NoError(t, err)
    assert.NotEmpty(t, hash)
    assert.Len(t, hash, 64) // SHA256 hex length
}

func TestContentHasher_HashPackage_Deterministic(t *testing.T) {
    fs := memfs.New()
    pkgPath := mustPackagePath(t, "/stow/vim")
    require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))
    
    vimrcPath := filepath.Join(pkgPath.String(), "dot-vimrc")
    require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("set number\n"), 0644))
    
    hasher := manifest.NewContentHasher(fs)
    
    hash1, err := hasher.HashPackage(context.Background(), pkgPath)
    require.NoError(t, err)
    
    hash2, err := hasher.HashPackage(context.Background(), pkgPath)
    require.NoError(t, err)
    
    assert.Equal(t, hash1, hash2)
}

func TestContentHasher_HashPackage_DifferentContent(t *testing.T) {
    fs := memfs.New()
    pkgPath := mustPackagePath(t, "/stow/vim")
    require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))
    
    vimrcPath := filepath.Join(pkgPath.String(), "dot-vimrc")
    hasher := manifest.NewContentHasher(fs)
    
    // Hash with initial content
    require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("set number\n"), 0644))
    hash1, err := hasher.HashPackage(context.Background(), pkgPath)
    require.NoError(t, err)
    
    // Hash with different content
    require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("set relativenumber\n"), 0644))
    hash2, err := hasher.HashPackage(context.Background(), pkgPath)
    require.NoError(t, err)
    
    assert.NotEqual(t, hash1, hash2)
}

func TestContentHasher_HashPackage_NestedDirectories(t *testing.T) {
    fs := memfs.New()
    pkgPath := mustPackagePath(t, "/stow/vim")
    
    // Create nested structure
    colorsPath := filepath.Join(pkgPath.String(), "dot-vim", "colors")
    require.NoError(t, fs.MkdirAll(context.Background(), colorsPath, 0755))
    require.NoError(t, fs.WriteFile(context.Background(), 
        filepath.Join(colorsPath, "molokai.vim"), []byte("colorscheme"), 0644))
    
    hasher := manifest.NewContentHasher(fs)
    
    hash, err := hasher.HashPackage(context.Background(), pkgPath)
    
    require.NoError(t, err)
    assert.NotEmpty(t, hash)
}

func TestContentHasher_HashPackage_IgnoresSymlinks(t *testing.T) {
    fs := memfs.New()
    pkgPath := mustPackagePath(t, "/stow/vim")
    require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))
    
    // Create file and symlink
    vimrcPath := filepath.Join(pkgPath.String(), "dot-vimrc")
    linkPath := filepath.Join(pkgPath.String(), "link-to-vimrc")
    require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("content"), 0644))
    require.NoError(t, fs.Symlink(context.Background(), vimrcPath, linkPath))
    
    hasher := manifest.NewContentHasher(fs)
    
    hash, err := hasher.HashPackage(context.Background(), pkgPath)
    
    require.NoError(t, err)
    assert.NotEmpty(t, hash)
    // Symlink should not affect hash (only real files)
}
```

**Implementation** (`internal/manifest/hash.go`):
```go
package manifest

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "io/fs"
    "path/filepath"
    "sort"
    
    "github.com/yourusername/dot/pkg/dot"
)

// ContentHasher computes content hashes for packages
type ContentHasher struct {
    fs dot.FS
}

// NewContentHasher creates a new content hasher
func NewContentHasher(fs dot.FS) *ContentHasher {
    return &ContentHasher{fs: fs}
}

// HashPackage computes content hash for entire package
// Hash is deterministic and based on file contents and paths
func (h *ContentHasher) HashPackage(ctx context.Context, pkgPath dot.PackagePath) (string, error) {
    if ctx.Err() != nil {
        return "", ctx.Err()
    }
    
    hasher := sha256.New()
    
    // Collect all files in sorted order for determinism
    var files []string
    err := h.walkPackage(ctx, pkgPath.String(), &files)
    if err != nil {
        return "", fmt.Errorf("failed to walk package: %w", err)
    }
    
    sort.Strings(files)
    
    // Hash each file's path and content
    for _, relPath := range files {
        fullPath := filepath.Join(pkgPath.String(), relPath)
        
        // Write path to hash
        if _, err := hasher.Write([]byte(relPath)); err != nil {
            return "", fmt.Errorf("failed to hash path: %w", err)
        }
        
        // Write content to hash
        data, err := h.fs.ReadFile(ctx, fullPath)
        if err != nil {
            return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
        }
        
        if _, err := hasher.Write(data); err != nil {
            return "", fmt.Errorf("failed to hash content: %w", err)
        }
    }
    
    return hex.EncodeToString(hasher.Sum(nil)), nil
}

// walkPackage collects regular files recursively
func (h *ContentHasher) walkPackage(ctx context.Context, root string, files *[]string) error {
    entries, err := h.fs.ReadDir(ctx, root)
    if err != nil {
        return err
    }
    
    for _, entry := range entries {
        if ctx.Err() != nil {
            return ctx.Err()
        }
        
        fullPath := filepath.Join(root, entry.Name())
        
        if entry.IsDir() {
            if err := h.walkPackage(ctx, fullPath, files); err != nil {
                return err
            }
        } else if entry.Type().IsRegular() {
            // Store relative path for determinism
            relPath, err := filepath.Rel(root, fullPath)
            if err != nil {
                return err
            }
            *files = append(*files, relPath)
        }
        // Skip symlinks and other non-regular files
    }
    
    return nil
}
```

**Commit Message**:
```
feat(manifest): implement content hashing for packages

Add ContentHasher for fast, deterministic package content hashing.
Uses SHA256 over sorted file paths and contents to detect changes.

- Implement ContentHasher with HashPackage operation
- Use SHA256 for cryptographic-quality hashing
- Sort files for deterministic hash computation
- Hash both paths and contents
- Skip symlinks and non-regular files
- Support context cancellation

Refs: Phase 11.3
```

---

### Task 11.4.1: Implement Manifest Validation

**Objective**: Validate manifest consistency with filesystem state.

**Test First** (`internal/manifest/validate_test.go`):
```go
func TestValidator_Validate_EmptyManifest(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    
    m := manifest.New()
    validator := manifest.NewValidator(fs)
    
    result := validator.Validate(context.Background(), targetDir, m)
    
    require.True(t, result.IsValid)
    assert.Empty(t, result.Issues)
}

func TestValidator_Validate_ValidManifest(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    packageDir := mustStowPath(t, "/stow")
    
    // Create package structure
    pkgPath := filepath.Join(packageDir.String(), "vim")
    vimrcSrc := filepath.Join(pkgPath, "dot-vimrc")
    require.NoError(t, fs.MkdirAll(context.Background(), pkgPath, 0755))
    require.NoError(t, fs.WriteFile(context.Background(), vimrcSrc, []byte("content"), 0644))
    
    // Create target link
    vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    require.NoError(t, fs.Symlink(context.Background(), vimrcSrc, vimrcTarget))
    
    // Create manifest
    m := manifest.New()
    m.AddPackage(manifest.PackageInfo{
        Name:      "vim",
        LinkCount: 1,
        Links:     []string{".vimrc"},
    })
    
    validator := manifest.NewValidator(fs)
    result := validator.Validate(context.Background(), targetDir, m)
    
    assert.True(t, result.IsValid)
    assert.Empty(t, result.Issues)
}

func TestValidator_Validate_BrokenLink(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    
    // Create broken symlink
    vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    require.NoError(t, fs.Symlink(context.Background(), "/nonexistent", vimrcTarget))
    
    m := manifest.New()
    m.AddPackage(manifest.PackageInfo{
        Name:      "vim",
        LinkCount: 1,
        Links:     []string{".vimrc"},
    })
    
    validator := manifest.NewValidator(fs)
    result := validator.Validate(context.Background(), targetDir, m)
    
    assert.False(t, result.IsValid)
    assert.Len(t, result.Issues, 1)
    assert.Equal(t, manifest.IssueBrokenLink, result.Issues[0].Type)
}

func TestValidator_Validate_MissingLink(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    
    m := manifest.New()
    m.AddPackage(manifest.PackageInfo{
        Name:      "vim",
        LinkCount: 1,
        Links:     []string{".vimrc"},
    })
    
    validator := manifest.NewValidator(fs)
    result := validator.Validate(context.Background(), targetDir, m)
    
    assert.False(t, result.IsValid)
    assert.Len(t, result.Issues, 1)
    assert.Equal(t, manifest.IssueMissingLink, result.Issues[0].Type)
}

func TestValidator_Validate_ExtraLink(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    packageDir := mustStowPath(t, "/stow")
    
    // Create package and link not in manifest
    pkgPath := filepath.Join(packageDir.String(), "vim")
    vimrcSrc := filepath.Join(pkgPath, "dot-vimrc")
    require.NoError(t, fs.MkdirAll(context.Background(), pkgPath, 0755))
    require.NoError(t, fs.WriteFile(context.Background(), vimrcSrc, []byte("content"), 0644))
    
    vimrcTarget := filepath.Join(targetDir.String(), ".vimrc")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    require.NoError(t, fs.Symlink(context.Background(), vimrcSrc, vimrcTarget))
    
    m := manifest.New()
    // Manifest is empty but link exists
    
    validator := manifest.NewValidator(fs)
    result := validator.Validate(context.Background(), targetDir, m)
    
    assert.False(t, result.IsValid)
    assert.Len(t, result.Issues, 1)
    assert.Equal(t, manifest.IssueExtraLink, result.Issues[0].Type)
}
```

**Implementation** (`internal/manifest/validate.go`):
```go
package manifest

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/yourusername/dot/pkg/dot"
)

// Validator checks manifest consistency with filesystem
type Validator struct {
    fs dot.FS
}

// NewValidator creates a new manifest validator
func NewValidator(fs dot.FS) *Validator {
    return &Validator{fs: fs}
}

// ValidationResult contains validation outcome and issues
type ValidationResult struct {
    IsValid bool
    Issues  []ValidationIssue
}

// ValidationIssue describes a specific problem found
type ValidationIssue struct {
    Type        IssueType
    Path        string
    Package     string
    Description string
}

// IssueType categorizes validation problems
type IssueType int

const (
    IssueBrokenLink IssueType = iota
    IssueMissingLink
    IssueExtraLink
    IssueWrongTarget
    IssueNotSymlink
)

// Validate checks manifest consistency with filesystem
func (v *Validator) Validate(ctx context.Context, targetDir dot.TargetPath, manifest Manifest) ValidationResult {
    result := ValidationResult{
        IsValid: true,
        Issues:  []ValidationIssue{},
    }
    
    // Check all links in manifest exist and are valid
    for _, pkg := range manifest.Packages {
        for _, linkPath := range pkg.Links {
            if ctx.Err() != nil {
                break
            }
            
            fullPath := filepath.Join(targetDir.String(), linkPath)
            issue := v.validateLink(ctx, fullPath, pkg.Name)
            if issue != nil {
                result.IsValid = false
                result.Issues = append(result.Issues, *issue)
            }
        }
    }
    
    // Check for extra links not in manifest
    extraIssues := v.findExtraLinks(ctx, targetDir, manifest)
    if len(extraIssues) > 0 {
        result.IsValid = false
        result.Issues = append(result.Issues, extraIssues...)
    }
    
    return result
}

// validateLink checks if a specific link is valid
func (v *Validator) validateLink(ctx context.Context, linkPath, pkgName string) *ValidationIssue {
    // Check if link exists
    exists := v.fs.Exists(ctx, linkPath)
    if !exists {
        return &ValidationIssue{
            Type:        IssueMissingLink,
            Path:        linkPath,
            Package:     pkgName,
            Description: "Link specified in manifest does not exist",
        }
    }
    
    // Check if it's a symlink
    isSymlink := v.fs.IsSymlink(ctx, linkPath)
    if !isSymlink {
        return &ValidationIssue{
            Type:        IssueNotSymlink,
            Path:        linkPath,
            Package:     pkgName,
            Description: "Path is not a symlink",
        }
    }
    
    // Check if link target exists
    target, err := v.fs.ReadLink(ctx, linkPath)
    if err != nil {
        return &ValidationIssue{
            Type:        IssueBrokenLink,
            Path:        linkPath,
            Package:     pkgName,
            Description: fmt.Sprintf("Cannot read link: %v", err),
        }
    }
    
    // Check if target exists
    targetExists := v.fs.Exists(ctx, target)
    if !targetExists {
        return &ValidationIssue{
            Type:        IssueBrokenLink,
            Path:        linkPath,
            Package:     pkgName,
            Description: fmt.Sprintf("Link target does not exist: %s", target),
        }
    }
    
    return nil
}

// findExtraLinks identifies symlinks not tracked in manifest
func (v *Validator) findExtraLinks(ctx context.Context, targetDir dot.TargetPath, manifest Manifest) []ValidationIssue {
    var issues []ValidationIssue
    
    // Build set of expected links
    expected := make(map[string]bool)
    for _, pkg := range manifest.Packages {
        for _, link := range pkg.Links {
            expected[link] = true
        }
    }
    
    // Walk target directory looking for unexpected symlinks
    v.walkForLinks(ctx, targetDir.String(), targetDir.String(), expected, &issues)
    
    return issues
}

// walkForLinks recursively finds symlinks
func (v *Validator) walkForLinks(ctx context.Context, root, current string, expected map[string]bool, issues *[]ValidationIssue) {
    entries, err := v.fs.ReadDir(ctx, current)
    if err != nil {
        return
    }
    
    for _, entry := range entries {
        if ctx.Err() != nil {
            return
        }
        
        fullPath := filepath.Join(current, entry.Name())
        relPath, _ := filepath.Rel(root, fullPath)
        
        if v.fs.IsSymlink(ctx, fullPath) {
            if !expected[relPath] {
                *issues = append(*issues, ValidationIssue{
                    Type:        IssueExtraLink,
                    Path:        relPath,
                    Description: "Symlink exists but not tracked in manifest",
                })
            }
        } else if entry.IsDir() {
            v.walkForLinks(ctx, root, fullPath, expected, issues)
        }
    }
}
```

**Commit Message**:
```
feat(manifest): implement manifest validation

Add validation to check manifest consistency with filesystem state.
Detects broken links, missing links, and untracked symlinks.

- Implement Validator with Validate operation
- Define validation issue types and results
- Check all manifest links exist and are valid
- Detect broken symlinks pointing to nonexistent targets
- Find extra symlinks not tracked in manifest
- Support context cancellation

Refs: Phase 11.4
```

---

### Task 11.4.2: Integration with IncrementalPlanner

**Objective**: Connect ContentHasher with IncrementalPlanner for change detection.

**Test First** (`internal/planner/incremental_test.go` - new tests):
```go
func TestIncrementalPlanner_DetectChangedPackages_NoManifest(t *testing.T) {
    fs := memfs.New()
    manifestStore := manifest.NewFSManifestStore(fs)
    hasher := manifest.NewContentHasher(fs)
    planner := planner.NewIncrementalPlanner(manifestStore, hasher)
    
    packageDir := mustStowPath(t, "/stow")
    packages := []string{"vim", "zsh"}
    
    changed := planner.DetectChangedPackages(context.Background(), packageDir, packages, manifest.New())
    
    // No previous manifest means all packages are changed
    assert.ElementsMatch(t, packages, changed)
}

func TestIncrementalPlanner_DetectChangedPackages_UnchangedPackage(t *testing.T) {
    fs := memfs.New()
    
    // Create package
    packageDir := mustStowPath(t, "/stow")
    vimPath := filepath.Join(packageDir.String(), "vim")
    require.NoError(t, fs.MkdirAll(context.Background(), vimPath, 0755))
    require.NoError(t, fs.WriteFile(context.Background(), 
        filepath.Join(vimPath, "dot-vimrc"), []byte("content"), 0644))
    
    // Compute initial hash
    hasher := manifest.NewContentHasher(fs)
    hash, err := hasher.HashPackage(context.Background(), mustPackagePath(t, vimPath))
    require.NoError(t, err)
    
    // Create manifest with hash
    m := manifest.New()
    m.SetHash("vim", hash)
    
    manifestStore := manifest.NewFSManifestStore(fs)
    planner := planner.NewIncrementalPlanner(manifestStore, hasher)
    
    changed := planner.DetectChangedPackages(context.Background(), packageDir, []string{"vim"}, m)
    
    // Package unchanged
    assert.Empty(t, changed)
}

func TestIncrementalPlanner_DetectChangedPackages_ChangedPackage(t *testing.T) {
    fs := memfs.New()
    
    // Create package with initial content
    packageDir := mustStowPath(t, "/stow")
    vimPath := filepath.Join(packageDir.String(), "vim")
    vimrcPath := filepath.Join(vimPath, "dot-vimrc")
    require.NoError(t, fs.MkdirAll(context.Background(), vimPath, 0755))
    require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("initial"), 0644))
    
    hasher := manifest.NewContentHasher(fs)
    initialHash, err := hasher.HashPackage(context.Background(), mustPackagePath(t, vimPath))
    require.NoError(t, err)
    
    // Create manifest
    m := manifest.New()
    m.SetHash("vim", initialHash)
    
    // Modify package
    require.NoError(t, fs.WriteFile(context.Background(), vimrcPath, []byte("modified"), 0644))
    
    manifestStore := manifest.NewFSManifestStore(fs)
    planner := planner.NewIncrementalPlanner(manifestStore, hasher)
    
    changed := planner.DetectChangedPackages(context.Background(), packageDir, []string{"vim"}, m)
    
    // Package changed
    assert.Equal(t, []string{"vim"}, changed)
}
```

**Implementation** (update `internal/planner/incremental.go`):
```go
// DetectChangedPackages identifies packages with content changes
func (p *IncrementalPlanner) DetectChangedPackages(
    ctx context.Context,
    packageDir dot.StowPath,
    packages []string,
    manifest manifest.Manifest,
) []string {
    var changed []string
    
    for _, pkgName := range packages {
        if ctx.Err() != nil {
            break
        }
        
        pkgPath := packageDir.Join(pkgName)
        
        // Compute current hash
        currentHash, err := p.hasher.HashPackage(ctx, pkgPath)
        if err != nil {
            // Error computing hash - treat as changed
            changed = append(changed, pkgName)
            continue
        }
        
        // Get previous hash
        prevHash, exists := manifest.GetHash(pkgName)
        if !exists || prevHash != currentHash {
            changed = append(changed, pkgName)
        }
    }
    
    return changed
}
```

**Commit Message**:
```
feat(planner): integrate content hashing with incremental planner

Connect ContentHasher with IncrementalPlanner for efficient change
detection. Only packages with content changes are reprocessed.

- Add DetectChangedPackages to IncrementalPlanner
- Compare current package hashes with manifest
- Treat missing hashes as changed
- Treat hash computation errors as changed
- Enable fast incremental restow operations

Refs: Phase 11.3, Phase 11.4
```

---

## Testing Strategy

### Unit Tests
- Test each type and function independently
- Use memory filesystem for fast, isolated tests
- Cover edge cases: empty manifests, corrupt data, missing files
- Verify error handling and context cancellation

### Integration Tests
```go
func TestManifest_Integration_FullWorkflow(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    
    store := manifest.NewFSManifestStore(fs)
    
    // Create and save manifest
    m := manifest.New()
    m.AddPackage(manifest.PackageInfo{
        Name:      "vim",
        LinkCount: 2,
        Links:     []string{".vimrc", ".vim/colors"},
    })
    m.SetHash("vim", "abc123")
    
    err := store.Save(context.Background(), targetDir, m)
    require.NoError(t, err)
    
    // Load and verify
    result := store.Load(context.Background(), targetDir)
    require.True(t, result.IsOk())
    
    loaded, _ := result.Unwrap()
    assert.Equal(t, "1.0", loaded.Version)
    assert.Len(t, loaded.Packages, 1)
    
    vim, exists := loaded.GetPackage("vim")
    assert.True(t, exists)
    assert.Equal(t, 2, vim.LinkCount)
    
    hash, exists := loaded.GetHash("vim")
    assert.True(t, exists)
    assert.Equal(t, "abc123", hash)
}

func TestManifest_Integration_ConcurrentAccess(t *testing.T) {
    fs := memfs.New()
    targetDir := mustTargetPath(t, "/home/user")
    require.NoError(t, fs.MkdirAll(context.Background(), targetDir.String(), 0755))
    
    store := manifest.NewFSManifestStore(fs)
    
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            m := manifest.New()
            m.AddPackage(manifest.PackageInfo{
                Name: fmt.Sprintf("pkg%d", id),
            })
            
            err := store.Save(context.Background(), targetDir, m)
            assert.NoError(t, err)
        }(i)
    }
    
    wg.Wait()
}
```

### Property-Based Tests
```go
func TestManifest_Properties_HashDeterminism(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("hash is deterministic", prop.ForAll(
        func(content []byte) bool {
            fs := memfs.New()
            pkgPath := mustPackagePath(t, "/stow/pkg")
            require.NoError(t, fs.MkdirAll(context.Background(), pkgPath.String(), 0755))
            require.NoError(t, fs.WriteFile(context.Background(),
                filepath.Join(pkgPath.String(), "file"), content, 0644))
            
            hasher := manifest.NewContentHasher(fs)
            
            hash1, err := hasher.HashPackage(context.Background(), pkgPath)
            require.NoError(t, err)
            
            hash2, err := hasher.HashPackage(context.Background(), pkgPath)
            require.NoError(t, err)
            
            return hash1 == hash2
        },
        gen.SliceOfN(100, gen.Byte()),
    ))
    
    properties.TestingRun(t)
}
```

## Success Criteria

- [ ] All manifest types implemented and tested
- [ ] FSManifestStore loads and saves manifests atomically
- [ ] ContentHasher computes deterministic package hashes
- [ ] Validator detects manifest-filesystem inconsistencies
- [ ] IncrementalPlanner uses hashes for change detection
- [ ] Test coverage ≥ 80% for manifest package
- [ ] All linters pass
- [ ] Integration tests verify full workflows
- [ ] Property tests verify hash determinism
- [ ] Documentation updated

## Dependencies Update

After Phase 11 completion, update:
- `go.mod`: Ensure crypto/sha256 is available (standard library)
- `CHANGELOG.md`: Add manifest system features
- `README.md`: Document manifest file location and format

## Future Enhancements

Post-v0.1.0 considerations:
- Manifest versioning and migration
- Compressed manifest storage
- Manifest signing for integrity
- Remote manifest synchronization
- Manifest diff and merge operations

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Corrupt manifest breaks operations | High | Graceful degradation, rebuild from filesystem |
| Hash collisions | Low | Use SHA256 (cryptographic quality) |
| Large package hashing slow | Medium | Consider incremental hashing, parallel processing |
| Concurrent manifest writes | Medium | Atomic writes via temp file and rename |
| Manifest file permissions | Low | Document required permissions (0644) |

