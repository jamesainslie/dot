package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandPaths resolves shell-style references in every path-typed
// configuration value, in place.
//
// Two substitutions are applied, in this order:
//
//  1. Environment variables, so "$HOME/.dotfiles" and "${XDG_DATA_HOME}/dot"
//     resolve. Unlike a shell, a reference to a variable that is not set is
//     an error rather than an empty string: silently turning
//     "$DOTFILES_ROOT/repo" into the root-anchored "/repo" is never what the
//     author meant.
//  2. A leading tilde, so "~" becomes the current user's home directory and
//     "~/.dotfiles" becomes "<home>/.dotfiles".
//
// Only path-typed values are touched: directories.package,
// directories.target, directories.manifest, symlinks.backup_dir, and
// logging.file. Pattern and prefix values such as ignore.patterns and
// dotfile.prefix are left verbatim, because a tilde or dollar sign there is
// part of the pattern rather than a path reference.
//
// Expansion is applied once per configuration source: file values are
// expanded as the file is loaded, and environment or flag overlays are
// expanded before they are merged. Running it again over already expanded
// values would re-interpret a "$" or a leading "~" that a resolved path
// legitimately contains.
//
// On error the configuration is left as written, and the message names both
// the offending key and the reference that could not be resolved.
func (c *ExtendedConfig) ExpandPaths() error {
	fields := c.pathFields()

	expanded := make([]string, len(fields))
	for i, field := range fields {
		value, err := expandPath(*field.value)
		if err != nil {
			return fmt.Errorf("%s: %w", field.key, err)
		}
		expanded[i] = value
	}

	for i, field := range fields {
		*field.value = expanded[i]
	}

	return nil
}

// pathField pairs a path-typed configuration key with its value slot.
type pathField struct {
	key   string
	value *string
}

// pathFields returns every path-typed configuration value in a stable order.
// It is the single list both ExpandPaths and the environment override path
// consume, so a new path-typed field is expanded everywhere or nowhere.
func (c *ExtendedConfig) pathFields() []pathField {
	return []pathField{
		{"directories.package", &c.Directories.Package},
		{"directories.target", &c.Directories.Target},
		{"directories.manifest", &c.Directories.Manifest},
		{"symlinks.backup_dir", &c.Symlinks.BackupDir},
		{"logging.file", &c.Logging.File},
	}
}

// expandPath expands environment variables and a leading tilde in a single
// path value. An empty value stays empty, so callers can distinguish "unset"
// from "set to the home directory".
func expandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	expanded, err := expandVariables(path)
	if err != nil {
		return "", err
	}

	return expandTilde(expanded)
}

// expandVariables replaces "$VAR" and "${VAR}" with their environment values.
// A reference to a variable that is not set is reported rather than expanded
// to the empty string, because an empty expansion yields a path that looks
// plausible ("/repo") while pointing somewhere the author never named.
func expandVariables(path string) (string, error) {
	var missing []string

	expanded := os.Expand(path, func(name string) string {
		value, found := os.LookupEnv(name)
		if !found {
			if !contains(missing, name) {
				missing = append(missing, name)
			}
			return ""
		}
		return value
	})

	if len(missing) > 0 {
		names := make([]string, 0, len(missing))
		for _, name := range missing {
			names = append(names, "$"+name)
		}
		return "", fmt.Errorf("undefined environment variable %s in %q",
			strings.Join(names, ", "), path)
	}

	return expanded, nil
}

// expandTilde replaces a leading "~" or "~/" with the current user's home
// directory. Other forms, including "~user" and a tilde anywhere but the
// start, are returned unchanged: dot does not resolve other users' home
// directories, and a tilde inside a path is a legitimate directory name.
//
// A home directory that cannot be determined is an error, so the caller sees
// which key could not be resolved rather than a later failure on a literal
// path named "~".
func expandTilde(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", path, err)
	}
	if home == "" {
		return "", fmt.Errorf("resolve %q: home directory is empty", path)
	}

	if path == "~" {
		return home, nil
	}

	return filepath.Join(home, path[len("~/"):]), nil
}
