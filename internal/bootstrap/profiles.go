package bootstrap

import (
	"fmt"
	"slices"
	"strings"
)

// ResolveProfilePackages returns the packages of a profile with inheritance applied.
//
// A profile may declare a single parent through the extends field. The resolved
// package list is the union of the whole parent chain and the profile's own
// packages, ordered from the root ancestor down to the profile itself. Duplicate
// names are removed, keeping the earliest (most ancestral) occurrence.
//
// Returns an error if:
//   - The profile does not exist
//   - Any profile in the chain extends an unknown profile
//   - The chain contains a cycle
func ResolveProfilePackages(profiles map[string]Profile, name string) ([]string, error) {
	chain, err := profileChain(profiles, name)
	if err != nil {
		return nil, err
	}

	resolved := make([]string, 0)
	seen := make(map[string]struct{})

	for _, profileName := range chain {
		for _, pkg := range profiles[profileName].Packages {
			if _, duplicate := seen[pkg]; duplicate {
				continue
			}
			seen[pkg] = struct{}{}
			resolved = append(resolved, pkg)
		}
	}

	return resolved, nil
}

// profileChain returns the inheritance chain for a profile, root ancestor first.
//
// For a profile "work" extending "dev" extending "base", the chain is
// [base, dev, work].
func profileChain(profiles map[string]Profile, name string) ([]string, error) {
	if _, exists := profiles[name]; !exists {
		return nil, fmt.Errorf("profile not found: %s", name)
	}

	// Walk parents, collecting the chain from the profile upwards.
	descending := make([]string, 0, len(profiles))
	visited := make(map[string]struct{}, len(profiles))

	for current := name; ; {
		if _, seen := visited[current]; seen {
			cycle := strings.Join(append(descending, current), " -> ")
			return nil, fmt.Errorf("profile %q has a circular extends chain: %s", name, cycle)
		}
		visited[current] = struct{}{}
		descending = append(descending, current)

		parent := profiles[current].Extends
		if parent == "" {
			break
		}
		if _, exists := profiles[parent]; !exists {
			return nil, fmt.Errorf("profile %q extends unknown profile: %s", current, parent)
		}
		current = parent
	}

	slices.Reverse(descending)
	return descending, nil
}
