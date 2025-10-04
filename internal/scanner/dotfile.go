package scanner

import (
	"path/filepath"
	"strings"
)

// TranslateDotfile converts "dot-filename" to ".filename".
// Files with "dot-" prefix become dotfiles in the target directory.
//
// Examples:
//   - "dot-vimrc" -> ".vimrc"
//   - "dot-bashrc" -> ".bashrc"
//   - "README.md" -> "README.md" (no change)
func TranslateDotfile(name string) string {
	if strings.HasPrefix(name, "dot-") {
		return "." + name[4:] // Replace "dot-" with "."
	}
	return name
}

// UntranslateDotfile converts ".filename" to "dot-filename".
// This is the reverse operation of TranslateDotfile.
//
// Examples:
//   - ".vimrc" -> "dot-vimrc"
//   - ".bashrc" -> "dot-bashrc"
//   - "README.md" -> "README.md" (no change)
func UntranslateDotfile(name string) string {
	if strings.HasPrefix(name, ".") && len(name) > 1 {
		return "dot-" + name[1:] // Replace "." with "dot-"
	}
	return name
}

// TranslatePath translates the last component of a path if it has dot- prefix.
// This handles paths like "vim/dot-vimrc" -> "vim/.vimrc".
//
// The function only translates the final component (base name), leaving
// directory components unchanged.
func TranslatePath(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	translated := TranslateDotfile(base)

	if dir == "." {
		return translated
	}

	return filepath.Join(dir, translated)
}

// UntranslatePath translates the last component of a path if it starts with dot.
// This is the reverse of TranslatePath.
func UntranslatePath(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	untranslated := UntranslateDotfile(base)

	if dir == "." {
		return untranslated
	}

	return filepath.Join(dir, untranslated)
}
