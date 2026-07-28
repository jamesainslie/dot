package gitsync

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// RootLabel is the display name used for changed files that live directly in
// the repository root rather than inside a package directory.
const RootLabel = "(repository root)"

// unknownHost is the fallback used when the machine reports no usable hostname.
const unknownHost = "unknown-host"

// StatusEntry is a single line of `git status --porcelain` output.
// Index holds the staged status code, Worktree the unstaged status code.
// OrigPath is populated only for renames and copies.
type StatusEntry struct {
	Index    byte
	Worktree byte
	Path     string
	OrigPath string
}

// Code returns the two-character porcelain status code, for example "??" or " M".
func (e StatusEntry) Code() string {
	return string([]byte{e.Index, e.Worktree})
}

// PackageGroup collects the changed entries that belong to one top-level
// directory of the package repository. Package is empty for repository-root files.
type PackageGroup struct {
	Package string
	Entries []StatusEntry
}

// Label returns the group name for display, substituting RootLabel for
// repository-root files.
func (g PackageGroup) Label() string {
	if g.Package == "" {
		return RootLabel
	}
	return g.Package
}

// ParsePorcelain parses `git status --porcelain` (v1) output into entries.
// Lines that are too short to carry a status code and a path are skipped.
func ParsePorcelain(out string) []StatusEntry {
	entries := make([]StatusEntry, 0)

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 4 {
			continue
		}

		entry := StatusEntry{
			Index:    line[0],
			Worktree: line[1],
		}

		rest := line[3:]
		if entry.Index == 'R' || entry.Index == 'C' {
			if orig, newPath, ok := splitRename(rest); ok {
				entry.OrigPath = unquotePath(orig)
				entry.Path = unquotePath(newPath)
				entries = append(entries, entry)
				continue
			}
		}

		entry.Path = unquotePath(rest)
		if entry.Path == "" {
			continue
		}
		entries = append(entries, entry)
	}

	return entries
}

// splitRename splits a porcelain rename or copy payload of the form
// "old -> new" into its two paths.
func splitRename(rest string) (orig, newPath string, ok bool) {
	const arrow = " -> "
	idx := strings.Index(rest, arrow)
	if idx < 0 {
		return "", "", false
	}
	return rest[:idx], rest[idx+len(arrow):], true
}

// unquotePath removes git's C-style quoting from a path when present.
// Paths that fail to unquote are returned with the surrounding quotes stripped.
func unquotePath(path string) string {
	if !strings.HasPrefix(path, `"`) {
		return path
	}
	if unquoted, err := strconv.Unquote(path); err == nil {
		return unquoted
	}
	return strings.Trim(path, `"`)
}

// GroupByPackage groups entries by their top-level directory. Groups are
// sorted by package name with repository-root files first; entries inside a
// group are sorted by path.
func GroupByPackage(entries []StatusEntry) []PackageGroup {
	byPackage := make(map[string][]StatusEntry)
	for _, entry := range entries {
		pkg := topLevelDir(entry.Path)
		byPackage[pkg] = append(byPackage[pkg], entry)
	}

	names := make([]string, 0, len(byPackage))
	for name := range byPackage {
		names = append(names, name)
	}
	sort.Strings(names)

	groups := make([]PackageGroup, 0, len(names))
	for _, name := range names {
		grouped := byPackage[name]
		sort.Slice(grouped, func(i, j int) bool {
			return grouped[i].Path < grouped[j].Path
		})
		groups = append(groups, PackageGroup{Package: name, Entries: grouped})
	}

	return groups
}

// topLevelDir returns the first path segment of a repository-relative path,
// or an empty string when the path names a file in the repository root.
func topLevelDir(path string) string {
	path = strings.TrimPrefix(path, "./")
	idx := strings.Index(path, "/")
	if idx < 0 {
		return ""
	}
	return path[:idx]
}

// PackageLabels returns the display labels of the given groups, in order.
func PackageLabels(groups []PackageGroup) []string {
	labels := make([]string, 0, len(groups))
	for _, group := range groups {
		labels = append(labels, group.Label())
	}
	return labels
}

// ShortHostname reduces a hostname to its first label, dropping any domain
// suffix. It returns unknownHost when no usable label is present.
func ShortHostname(hostname string) string {
	short, _, _ := strings.Cut(hostname, ".")
	if short == "" {
		return unknownHost
	}
	return short
}

// CommitMessage builds the commit message used by `dot sync --push`.
func CommitMessage(hostname string, groups []PackageGroup) string {
	base := fmt.Sprintf("sync: %s local edits", ShortHostname(hostname))
	if len(groups) == 0 {
		return base
	}
	return fmt.Sprintf("%s (%s)", base, strings.Join(PackageLabels(groups), ", "))
}
