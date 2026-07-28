package bootstrap

import (
	"fmt"
	"path"
	"strings"
)

// MachineRule maps a host pattern to a profile name.
//
// Rules are an ordered list rather than a map because evaluation order is part
// of the semantics: patterns are allowed to overlap and the first matching rule
// wins. YAML mappings do not preserve order in Go, so a list is the only way to
// express that intent.
type MachineRule struct {
	// Host is a glob pattern matched against the machine hostname.
	// Pattern syntax is path.Match: * and ? and [class] ranges, with no
	// special treatment of dots.
	Host string `yaml:"host"`

	// Profile names the profile applied on matching hosts.
	Profile string `yaml:"profile"`
}

// ResolveMachineProfile returns the first machine rule matching the hostname.
//
// Each pattern is matched against both the full hostname and its first label,
// so a rule for "hephaestus" matches "hephaestus.example.com" as well.
// Matching is case-insensitive. Rules are evaluated in declaration order and
// the first match wins; a trailing "*" rule therefore acts as a catch-all.
//
// Returns false when the hostname is empty or no rule matches.
func ResolveMachineProfile(machines []MachineRule, hostname string) (MachineRule, bool) {
	if hostname == "" {
		return MachineRule{}, false
	}

	full := strings.ToLower(hostname)
	short := full
	if idx := strings.Index(full, "."); idx > 0 {
		short = full[:idx]
	}

	for _, rule := range machines {
		pattern := strings.ToLower(rule.Host)
		if matchHost(pattern, full) || matchHost(pattern, short) {
			return rule, true
		}
	}

	return MachineRule{}, false
}

// matchHost reports whether a host pattern matches a hostname.
// Malformed patterns never match; Validate rejects them at load time.
func matchHost(pattern, hostname string) bool {
	matched, err := path.Match(pattern, hostname)
	if err != nil {
		return false
	}
	return matched
}

// validateMachines validates the machines section.
//
// Each entry must carry a non-empty, well-formed host pattern and name a
// profile that exists.
func (c Config) validateMachines() error {
	for _, rule := range c.Machines {
		if rule.Host == "" {
			return fmt.Errorf("machine host pattern cannot be empty")
		}

		if _, err := path.Match(rule.Host, ""); err != nil {
			return fmt.Errorf("invalid machine host pattern %q: %w", rule.Host, err)
		}

		if rule.Profile == "" {
			return fmt.Errorf("machine entry for host %q has no profile", rule.Host)
		}

		if _, exists := c.Profiles[rule.Profile]; !exists {
			return fmt.Errorf("machine entry for host %q references unknown profile: %s", rule.Host, rule.Profile)
		}
	}

	return nil
}
