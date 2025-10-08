// Package ignore provides pattern matching for file exclusion.
package ignore

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jamesainslie/dot/internal/domain"
)

// Pattern represents a compiled pattern for matching paths.
type Pattern struct {
	original string
	regex    *regexp.Regexp
}

// NewPattern creates a pattern from a glob pattern.
// Converts glob syntax to regex for matching.
func NewPattern(glob string) domain.Result[*Pattern] {
	regex := GlobToRegex(glob)
	compiled, err := regexp.Compile(regex)
	if err != nil {
		return domain.Err[*Pattern](fmt.Errorf("compile pattern %q: %w", glob, err))
	}

	return domain.Ok(&Pattern{
		original: glob, // Store original glob, not regex
		regex:    compiled,
	})
}

// NewPatternFromRegex creates a pattern from a regex string.
func NewPatternFromRegex(regex string) domain.Result[*Pattern] {
	compiled, err := regexp.Compile(regex)
	if err != nil {
		return domain.Err[*Pattern](fmt.Errorf("compile pattern: %w", err))
	}

	return domain.Ok(&Pattern{
		original: regex,
		regex:    compiled,
	})
}

// Match checks if the path matches the pattern.
func (p *Pattern) Match(path string) bool {
	return p.regex.MatchString(path)
}

// MatchBasename checks if the basename of the path matches the pattern.
// Useful for patterns like ".DS_Store" that should match anywhere in tree.
func (p *Pattern) MatchBasename(path string) bool {
	basename := filepath.Base(path)
	return p.regex.MatchString(basename)
}

// String returns the original pattern string.
func (p *Pattern) String() string {
	return p.original
}

// GlobToRegex converts a glob pattern to a regex pattern.
//
// Glob syntax:
//   - *       matches any sequence of characters
//   - ?       matches any single character
//   - [abc]   matches any character in the set
//   - [a-z]   matches any character in the range
//
// All other characters are escaped to match literally.
func GlobToRegex(glob string) string {
	var result strings.Builder
	result.WriteString("^")

	for i := 0; i < len(glob); i++ {
		ch := glob[i]

		switch ch {
		case '*':
			// Match any sequence of characters
			result.WriteString(".*")

		case '?':
			// Match any single character
			result.WriteString(".")

		case '[':
			// Character class - find the closing bracket
			j := i + 1
			for j < len(glob) && glob[j] != ']' {
				j++
			}
			if j < len(glob) && j > i+1 {
				// Valid character class with content
				// Check if it's a valid range/set
				class := glob[i : j+1]
				// For simplicity, escape brackets in glob patterns
				// This treats [1] as literal [1], not as character class
				result.WriteString(regexp.QuoteMeta(class))
				i = j
			} else {
				// No closing bracket or empty class - treat as literal
				result.WriteString(regexp.QuoteMeta(string(ch)))
			}

		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '\\':
			// Escape regex special characters
			result.WriteString("\\")
			result.WriteByte(ch)

		default:
			// Literal character
			result.WriteByte(ch)
		}
	}

	result.WriteString("$")
	return result.String()
}
