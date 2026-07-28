package main

import "runtime/debug"

// effectiveVersion resolves the version reported by dot --version.
//
// Precedence: an ldflags-stamped version (goreleaser and Makefile builds)
// wins; otherwise the module version embedded by the Go toolchain covers
// go install builds; a source build inside the repo, where the module
// version is "(devel)", degrades to "dev".
func effectiveVersion(ldflagsVersion string, info *debug.BuildInfo) string {
	if ldflagsVersion != "" && ldflagsVersion != "dev" {
		return ldflagsVersion
	}
	if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	if ldflagsVersion == "" {
		return "dev"
	}
	return ldflagsVersion
}

// resolveVersion applies effectiveVersion to the running binary's build info.
func resolveVersion(ldflagsVersion string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		info = nil
	}
	return effectiveVersion(ldflagsVersion, info)
}
