package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubRelease represents a GitHub release from the API.
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PreRelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
}

// Version represents a semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Raw        string
}

// VersionChecker checks for new versions from GitHub releases.
type VersionChecker struct {
	httpClient *http.Client
	repository string
}

// NewVersionChecker creates a new version checker.
func NewVersionChecker(repository string) *VersionChecker {
	return &VersionChecker{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		repository: repository,
	}
}

// GetLatestVersion fetches the latest release from GitHub.
func (vc *VersionChecker) GetLatestVersion(includePrerelease bool) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", vc.repository)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "dot-updater")

	resp, err := vc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned status %d: %s", resp.StatusCode, string(body))
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decode releases: %w", err)
	}

	// Find the latest non-draft release
	for _, release := range releases {
		if release.Draft {
			continue
		}
		if release.PreRelease && !includePrerelease {
			continue
		}
		return &release, nil
	}

	return nil, fmt.Errorf("no suitable release found")
}

// ParseVersion parses a version string into a Version struct.
func ParseVersion(versionStr string) (*Version, error) {
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	v := &Version{Raw: versionStr}

	// Split on '-' to separate version from pre-release
	parts := strings.SplitN(versionStr, "-", 2)
	if len(parts) == 2 {
		v.PreRelease = parts[1]
	}

	// Parse major.minor.patch
	versionParts := strings.Split(parts[0], ".")
	if len(versionParts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	if _, err := fmt.Sscanf(versionParts[0], "%d", &v.Major); err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}
	if _, err := fmt.Sscanf(versionParts[1], "%d", &v.Minor); err != nil {
		return nil, fmt.Errorf("invalid minor version: %w", err)
	}
	if _, err := fmt.Sscanf(versionParts[2], "%d", &v.Patch); err != nil {
		return nil, fmt.Errorf("invalid patch version: %w", err)
	}

	return v, nil
}

// IsNewerThan returns true if v is newer than other.
func (v *Version) IsNewerThan(other *Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch > other.Patch
	}

	// If versions are equal, consider pre-release
	// A release version is newer than a pre-release version
	if v.PreRelease == "" && other.PreRelease != "" {
		return true
	}
	if v.PreRelease != "" && other.PreRelease == "" {
		return false
	}

	// Both are pre-releases or both are releases - consider equal
	return false
}

// String returns the version as a string.
func (v *Version) String() string {
	if v.PreRelease != "" {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.PreRelease)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// CheckForUpdate checks if there's a newer version available.
func (vc *VersionChecker) CheckForUpdate(currentVersion string, includePrerelease bool) (newVersion *GitHubRelease, hasUpdate bool, err error) {
	current, err := ParseVersion(currentVersion)
	if err != nil {
		return nil, false, fmt.Errorf("parse current version: %w", err)
	}

	latest, err := vc.GetLatestVersion(includePrerelease)
	if err != nil {
		return nil, false, fmt.Errorf("get latest version: %w", err)
	}

	latestVersion, err := ParseVersion(latest.TagName)
	if err != nil {
		return nil, false, fmt.Errorf("parse latest version: %w", err)
	}

	if latestVersion.IsNewerThan(current) {
		return latest, true, nil
	}

	return latest, false, nil
}
