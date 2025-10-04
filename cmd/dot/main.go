package main

import (
	"fmt"
	"io"
	"os"
)

// Build information populated via ldflags during compilation.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	exitCode := run(os.Args, os.Stdout, os.Stderr)
	os.Exit(exitCode)
}

// run executes the CLI logic and returns an exit code.
// This function is extracted for testability.
func run(args []string, stdout, stderr io.Writer) int {
	// Handle version flag
	if len(args) > 1 && (args[1] == "version" || args[1] == "--version" || args[1] == "-v") {
		printVersion(stdout)
		return 0
	}

	// Handle help flag
	if len(args) > 1 && (args[1] == "help" || args[1] == "--help" || args[1] == "-h") {
		printHelp(stdout)
		return 0
	}

	// Default message when no command provided
	if len(args) == 1 {
		printHelp(stdout)
		return 0
	}

	// Unknown command
	fmt.Fprintf(stderr, "Error: unknown command '%s'\n\n", args[1])
	printHelp(stdout)
	return 1
}

// printVersion prints the version information.
func printVersion(w io.Writer) {
	fmt.Fprintf(w, "dot version %s\n", version)
	fmt.Fprintf(w, "commit: %s\n", commit)
	fmt.Fprintf(w, "built: %s\n", date)
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "dot - dotfile manager")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  dot <command> [arguments]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Available commands:")
	fmt.Fprintln(w, "  version    Display version information")
	fmt.Fprintln(w, "  help       Display this help message")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands under development:")
	fmt.Fprintln(w, "  manage     Install dotfile packages to target directory")
	fmt.Fprintln(w, "  unmanage   Remove managed dotfiles from target directory")
	fmt.Fprintln(w, "  remanage   Update managed dotfiles with latest changes")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "For more information, see: https://github.com/jamesainslie/dot")
}
