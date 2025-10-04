package main

import (
	"fmt"
	"os"
)

// Build information populated via ldflags during compilation.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("dot version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		os.Exit(0)
	}

	// Handle help flag
	if len(os.Args) > 1 && (os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		os.Exit(0)
	}

	// Default message when no command provided
	if len(os.Args) == 1 {
		printHelp()
		os.Exit(0)
	}

	// Unknown command
	fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n\n", os.Args[1])
	printHelp()
	os.Exit(1)
}

func printHelp() {
	fmt.Println("dot - dotfile manager")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dot <command> [arguments]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  version    Display version information")
	fmt.Println("  help       Display this help message")
	fmt.Println()
	fmt.Println("Commands under development:")
	fmt.Println("  manage     Install dotfile packages to target directory")
	fmt.Println("  unmanage   Remove managed dotfiles from target directory")
	fmt.Println("  remanage   Update managed dotfiles with latest changes")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/jamesainslie/dot")
}
