package main

import (
	"os"

	_ "github.com/jamesainslie/dot/internal/api" // Register Client implementation
)

// Version information (set via ldflags at build time)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := NewRootCommand(version, commit, date)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
