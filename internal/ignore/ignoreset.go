package ignore

// IgnoreSet is a collection of patterns for ignoring files.
type IgnoreSet struct {
	patterns []*Pattern
}

// NewIgnoreSet creates a new empty ignore set.
func NewIgnoreSet() *IgnoreSet {
	return &IgnoreSet{
		patterns: make([]*Pattern, 0),
	}
}

// NewDefaultIgnoreSet creates an ignore set with default patterns.
func NewDefaultIgnoreSet() *IgnoreSet {
	set := NewIgnoreSet()

	for _, glob := range DefaultIgnorePatterns() {
		set.Add(glob)
	}

	return set
}

// Add adds a glob pattern to the ignore set.
func (s *IgnoreSet) Add(glob string) error {
	result := NewPattern(glob)
	if result.IsErr() {
		return result.UnwrapErr()
	}

	s.patterns = append(s.patterns, result.Unwrap())
	return nil
}

// AddPattern adds a compiled pattern to the ignore set.
func (s *IgnoreSet) AddPattern(pattern *Pattern) {
	s.patterns = append(s.patterns, pattern)
}

// ShouldIgnore checks if a path should be ignored.
// Returns true if the path matches any pattern.
//
// The function checks both full path match and basename match
// to support patterns like ".DS_Store" matching anywhere in the tree.
func (s *IgnoreSet) ShouldIgnore(path string) bool {
	for _, pattern := range s.patterns {
		// Check full path match
		if pattern.Match(path) {
			return true
		}

		// Check basename match for patterns like ".git"
		if pattern.MatchBasename(path) {
			return true
		}
	}

	return false
}

// Size returns the number of patterns in the set.
func (s *IgnoreSet) Size() int {
	return len(s.patterns)
}

// DefaultIgnorePatterns returns the default set of patterns to ignore.
// These are common files that should not be managed.
func DefaultIgnorePatterns() []string {
	return []string{
		".git",
		".svn",
		".hg",
		".DS_Store",
		"Thumbs.db",
		"desktop.ini",
		".Trash",
		".Spotlight-V100",
		".TemporaryItems",
	}
}
