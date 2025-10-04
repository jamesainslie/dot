package planner

import "fmt"

// generateSuggestions creates actionable suggestions for conflicts
func generateSuggestions(c Conflict) []Suggestion {
	switch c.Type {
	case ConflictFileExists:
		return generateFileExistsSuggestions(c)
	case ConflictWrongLink:
		return generateWrongLinkSuggestions(c)
	case ConflictPermission:
		return generatePermissionSuggestions(c)
	case ConflictCircular:
		return generateCircularSuggestions(c)
	case ConflictFileExpected, ConflictDirExpected:
		return generateTypeMismatchSuggestions(c)
	default:
		return []Suggestion{}
	}
}

// generateFileExistsSuggestions provides suggestions for existing files
func generateFileExistsSuggestions(c Conflict) []Suggestion {
	return []Suggestion{
		{
			Action:      "Use --backup flag to preserve existing file",
			Explanation: "Moves conflicting file to backup location before linking",
			Example:     "dot stow --backup <package>",
		},
		{
			Action:      "Use dot adopt to move file into package",
			Explanation: "Incorporates existing file into package management",
			Example:     fmt.Sprintf("dot adopt <package> %s", c.Path.String()),
		},
		{
			Action:      "Remove conflicting file manually",
			Explanation: "Delete file if no longer needed",
			Example:     fmt.Sprintf("rm %s", c.Path.String()),
		},
	}
}

// generateWrongLinkSuggestions provides suggestions for incorrect symlinks
func generateWrongLinkSuggestions(c Conflict) []Suggestion {
	return []Suggestion{
		{
			Action:      "Unstow the other package first",
			Explanation: "Removes conflicting symlink from different package",
			Example:     "dot unstow <other-package>",
		},
		{
			Action:      "Use --overwrite to replace the link",
			Explanation: "Forces link to point to new package",
			Example:     "dot stow --overwrite <package>",
		},
		{
			Action:      "Check which package owns the link",
			Explanation: "Identify the conflicting package to decide which to keep",
			Example:     fmt.Sprintf("ls -l %s", c.Path.String()),
		},
	}
}

// generatePermissionSuggestions provides suggestions for permission errors
func generatePermissionSuggestions(c Conflict) []Suggestion {
	parentPath := c.Path.Parent()
	var parentStr string
	if parentPath.IsOk() {
		parentStr = parentPath.Unwrap().String()
	} else {
		parentStr = c.Path.String()
	}

	return []Suggestion{
		{
			Action:      "Check file permissions on target directory",
			Explanation: "Ensure you have write access to the target location",
			Example:     fmt.Sprintf("ls -ld %s", parentStr),
		},
		{
			Action:      "Run with appropriate permissions",
			Explanation: "May need elevated privileges for system directories",
			Example:     "sudo dot stow <package>",
		},
		{
			Action:      "Change ownership of target directory",
			Explanation: "Make yourself owner of the directory",
			Example:     fmt.Sprintf("sudo chown -R $USER %s", parentStr),
		},
	}
}

// generateCircularSuggestions provides suggestions for circular dependencies
func generateCircularSuggestions(c Conflict) []Suggestion {
	return []Suggestion{
		{
			Action:      "Check for symlinks pointing to symlinks",
			Explanation: "Identify the circular link chain",
			Example:     fmt.Sprintf("ls -l %s", c.Path.String()),
		},
		{
			Action:      "Remove the circular symlink manually",
			Explanation: "Break the circular dependency",
			Example:     fmt.Sprintf("rm %s", c.Path.String()),
		},
		{
			Action:      "Review package structure for recursive links",
			Explanation: "Ensure packages do not contain circular references",
		},
	}
}

// generateTypeMismatchSuggestions provides suggestions for type conflicts
func generateTypeMismatchSuggestions(c Conflict) []Suggestion {
	var expected, found string
	if c.Type == ConflictFileExpected {
		expected = "file"
		found = "directory"
	} else {
		expected = "directory"
		found = "file"
	}

	return []Suggestion{
		{
			Action:      fmt.Sprintf("Remove the conflicting %s", found),
			Explanation: fmt.Sprintf("Package expects a %s at this location", expected),
			Example:     fmt.Sprintf("rm -r %s", c.Path.String()),
		},
		{
			Action:      "Review package contents",
			Explanation: "Verify package structure matches your target directory layout",
		},
		{
			Action:      "Backup and remove conflict",
			Explanation: "Preserve existing structure before resolving",
			Example:     fmt.Sprintf("mv %s %s.backup", c.Path.String(), c.Path.String()),
		},
	}
}

// enrichConflictWithSuggestions adds suggestions to a conflict
func enrichConflictWithSuggestions(c Conflict) Conflict {
	suggestions := generateSuggestions(c)
	for _, s := range suggestions {
		c = c.WithSuggestion(s)
	}
	return c
}
