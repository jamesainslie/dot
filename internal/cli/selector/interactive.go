// Package selector provides package selection interfaces and implementations.
package selector

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// PackageSelector defines the interface for selecting packages.
type PackageSelector interface {
	// Select prompts the user to select packages from the provided list.
	// Returns the selected package names.
	Select(ctx context.Context, packages []string) ([]string, error)
}

// InteractiveSelector implements PackageSelector with interactive prompts.
type InteractiveSelector struct {
	input  io.Reader
	output io.Writer
}

// NewInteractiveSelector creates a new interactive selector.
func NewInteractiveSelector(input io.Reader, output io.Writer) *InteractiveSelector {
	return &InteractiveSelector{
		input:  input,
		output: output,
	}
}

// Select prompts the user to select packages interactively.
func (s *InteractiveSelector) Select(ctx context.Context, packages []string) ([]string, error) {
	// Handle empty package list
	if len(packages) == 0 {
		return []string{}, nil
	}

	// Display available packages
	fmt.Fprintln(s.output, "Available packages:")
	for i, pkg := range packages {
		fmt.Fprintf(s.output, "  %d) %s\n", i+1, pkg)
	}
	fmt.Fprintln(s.output, "")
	fmt.Fprintln(s.output, "Select packages to install:")
	fmt.Fprintln(s.output, "  - Enter numbers (e.g., 1,2,3)")
	fmt.Fprintln(s.output, "  - Enter ranges (e.g., 1-3)")
	fmt.Fprintln(s.output, "  - Enter 'all' to select all packages")
	fmt.Fprintln(s.output, "  - Enter 'none' to skip installation")
	fmt.Fprint(s.output, "Selection: ")

	// Read user input
	scanner := bufio.NewScanner(s.input)
	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Read input
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("read input: %w", err)
			}
			return nil, fmt.Errorf("unexpected end of input")
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			fmt.Fprint(s.output, "Selection: ")
			continue
		}

		// Parse selection
		indices, err := parseSelection(input, len(packages))
		if err != nil {
			fmt.Fprintf(s.output, "Invalid selection: %v\n", err)
			fmt.Fprint(s.output, "Selection: ")
			continue
		}

		// Build selected package list
		selected := make([]string, 0, len(indices))
		for _, idx := range indices {
			selected = append(selected, packages[idx])
		}

		return selected, nil
	}
}

// parseSelection parses user input into package indices.
//
// Supported formats:
//   - "1" - single number
//   - "1,3,5" - comma-separated numbers
//   - "1-3" - range
//   - "1, 3-5, 7" - mixed
//   - "all" - all packages
//   - "none" - no packages
//
// Returns zero-based indices.
func parseSelection(input string, maxIndex int) ([]int, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	// Handle special keywords
	if input == "all" {
		indices := make([]int, maxIndex)
		for i := range indices {
			indices[i] = i
		}
		return indices, nil
	}

	if input == "none" {
		return []int{}, nil
	}

	// Parse comma-separated parts
	parts := strings.Split(input, ",")
	indices := make(map[int]bool) // Use map to deduplicate

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check if it's a range
		if strings.Contains(part, "-") {
			rangeIndices, err := parseRange(part, maxIndex)
			if err != nil {
				return nil, err
			}
			for _, idx := range rangeIndices {
				indices[idx] = true
			}
			continue
		}

		// Parse single number
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", part)
		}

		if num < 1 || num > maxIndex {
			return nil, fmt.Errorf("number %d out of range (1-%d)", num, maxIndex)
		}

		indices[num-1] = true // Convert to zero-based index
	}

	// Convert map to sorted slice
	result := make([]int, 0, len(indices))
	for idx := range indices {
		result = append(result, idx)
	}
	sort.Ints(result)

	return result, nil
}

// parseRange parses a range like "1-3" into indices.
func parseRange(rangeStr string, maxIndex int) ([]int, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format: %s", rangeStr)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid range start: %s", parts[0])
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid range end: %s", parts[1])
	}

	if start < 1 || start > maxIndex {
		return nil, fmt.Errorf("range start %d out of range (1-%d)", start, maxIndex)
	}

	if end < 1 || end > maxIndex {
		return nil, fmt.Errorf("range end %d out of range (1-%d)", end, maxIndex)
	}

	if start > end {
		return nil, fmt.Errorf("invalid range: start %d is greater than end %d", start, end)
	}

	indices := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		indices = append(indices, i-1) // Convert to zero-based index
	}

	return indices, nil
}
