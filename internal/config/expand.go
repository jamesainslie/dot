package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPaths resolves shell-style references in every path-typed
// configuration value, in place.
//
// Two substitutions are applied, in this order:
//
//  1. Environment variables, using os.ExpandEnv, so "$HOME/.dotfiles" and
//     "${XDG_DATA_HOME}/dot" resolve. A variable that is not set expands to
//     the empty string, matching shell behaviour.
//  2. A leading tilde, so "~" becomes the current user's home directory and
//     "~/.dotfiles" becomes "<home>/.dotfiles".
//
// Only path-typed values are touched: directories.package,
// directories.target, directories.manifest, symlinks.backup_dir, and
// logging.file. Pattern and prefix values such as ignore.patterns and
// dotfile.prefix are left verbatim, because a tilde or dollar sign there is
// part of the pattern rather than a path reference.
//
// The operation is idempotent: expanding an already expanded configuration
// leaves it unchanged.
func (c *ExtendedConfig) ExpandPaths() {
	c.Directories.Package = expandPath(c.Directories.Package)
	c.Directories.Target = expandPath(c.Directories.Target)
	c.Directories.Manifest = expandPath(c.Directories.Manifest)
	c.Symlinks.BackupDir = expandPath(c.Symlinks.BackupDir)
	c.Logging.File = expandPath(c.Logging.File)
}

// expandPath expands environment variables and a leading tilde in a single
// path value. An empty value stays empty, so callers can distinguish "unset"
// from "set to the home directory".
func expandPath(path string) string {
	if path == "" {
		return ""
	}

	return expandTilde(os.ExpandEnv(path))
}

// expandTilde replaces a leading "~" or "~/" with the current user's home
// directory. Other forms, including "~user" and a tilde anywhere but the
// start, are returned unchanged: dot does not resolve other users' home
// directories, and a tilde inside a path is a legitimate directory name.
//
// When the home directory cannot be determined the value is returned as
// written, leaving the caller to fail on the literal path rather than
// silently relocating it.
func expandTilde(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}

	if path == "~" {
		return home
	}

	return filepath.Join(home, path[len("~/"):])
}
