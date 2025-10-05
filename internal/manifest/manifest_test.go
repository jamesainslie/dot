package manifest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManifest_New(t *testing.T) {
	m := New()

	assert.Equal(t, "1.0", m.Version)
	assert.NotNil(t, m.Packages)
	assert.NotNil(t, m.Hashes)
	assert.False(t, m.UpdatedAt.IsZero())
}

func TestManifest_AddPackage(t *testing.T) {
	m := New()

	pkg := PackageInfo{
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
	assert.Len(t, retrieved.Links, 2)
}

func TestManifest_RemovePackage(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim"})

	removed := m.RemovePackage("vim")

	assert.True(t, removed)
	_, exists := m.GetPackage("vim")
	assert.False(t, exists)
}

func TestManifest_RemovePackage_NotExists(t *testing.T) {
	m := New()

	removed := m.RemovePackage("nonexistent")

	assert.False(t, removed)
}

func TestManifest_SetHash(t *testing.T) {
	m := New()

	m.SetHash("vim", "abc123")

	hash, exists := m.GetHash("vim")
	assert.True(t, exists)
	assert.Equal(t, "abc123", hash)
}

func TestManifest_GetHash_NotExists(t *testing.T) {
	m := New()

	hash, exists := m.GetHash("nonexistent")

	assert.False(t, exists)
	assert.Empty(t, hash)
}

func TestManifest_PackageList(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim"})
	m.AddPackage(PackageInfo{Name: "zsh"})

	packages := m.PackageList()

	assert.Len(t, packages, 2)
	names := []string{packages[0].Name, packages[1].Name}
	assert.Contains(t, names, "vim")
	assert.Contains(t, names, "zsh")
}

func TestManifest_PackageList_Empty(t *testing.T) {
	m := New()

	packages := m.PackageList()

	assert.Empty(t, packages)
}

func TestManifest_RemovePackage_RemovesHash(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim"})
	m.SetHash("vim", "abc123")

	removed := m.RemovePackage("vim")

	assert.True(t, removed)
	_, exists := m.GetHash("vim")
	assert.False(t, exists)
}

func TestManifest_UpdatesTimestamp_OnAdd(t *testing.T) {
	m := New()
	originalTime := m.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	m.AddPackage(PackageInfo{Name: "vim"})

	assert.True(t, m.UpdatedAt.After(originalTime))
}

func TestManifest_UpdatesTimestamp_OnRemove(t *testing.T) {
	m := New()
	m.AddPackage(PackageInfo{Name: "vim"})
	time.Sleep(10 * time.Millisecond)
	originalTime := m.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	m.RemovePackage("vim")

	assert.True(t, m.UpdatedAt.After(originalTime))
}

func TestManifest_UpdatesTimestamp_OnSetHash(t *testing.T) {
	m := New()
	originalTime := m.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	m.SetHash("vim", "abc123")

	assert.True(t, m.UpdatedAt.After(originalTime))
}

