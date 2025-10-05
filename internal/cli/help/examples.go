package help

// Example represents a command usage example.
type Example struct {
	Description string
	Command     string
	Output      string // Optional expected output
}

// Command examples organized by command name.
var (
	// ManageExamples demonstrates the manage command.
	ManageExamples = []Example{
		{
			Description: "Install a single package",
			Command:     "dot manage vim",
		},
		{
			Description: "Install multiple packages",
			Command:     "dot manage vim tmux zsh",
		},
		{
			Description: "Preview installation without applying changes",
			Command:     "dot manage --dry-run vim",
		},
		{
			Description: "Install with absolute symlinks",
			Command:     "dot manage --absolute vim",
		},
		{
			Description: "Install from specific stow directory",
			Command:     "dot manage --dir /path/to/stow vim",
		},
	}

	// UnmanageExamples demonstrates the unmanage command.
	UnmanageExamples = []Example{
		{
			Description: "Uninstall a single package",
			Command:     "dot unmanage vim",
		},
		{
			Description: "Uninstall multiple packages",
			Command:     "dot unmanage vim tmux",
		},
		{
			Description: "Preview uninstallation without applying",
			Command:     "dot unmanage --dry-run vim",
		},
	}

	// RemanageExamples demonstrates the remanage command.
	RemanageExamples = []Example{
		{
			Description: "Reinstall a package after configuration changes",
			Command:     "dot remanage vim",
		},
		{
			Description: "Reinstall with absolute symlinks",
			Command:     "dot remanage --absolute vim",
		},
		{
			Description: "Reinstall all managed packages",
			Command:     "dot remanage --all",
		},
	}

	// AdoptExamples demonstrates the adopt command.
	AdoptExamples = []Example{
		{
			Description: "Adopt existing dotfile into package",
			Command:     "dot adopt vim ~/.vimrc",
		},
		{
			Description: "Adopt with custom filename in package",
			Command:     "dot adopt vim ~/.vimrc --as vimrc.backup",
		},
	}

	// StatusExamples demonstrates the status command.
	StatusExamples = []Example{
		{
			Description: "Show status of all packages",
			Command:     "dot status",
		},
		{
			Description: "Show status of specific packages",
			Command:     "dot status vim tmux",
		},
		{
			Description: "Show status in JSON format",
			Command:     "dot status --format json",
		},
		{
			Description: "Show detailed status",
			Command:     "dot status --verbose",
		},
	}

	// DoctorExamples demonstrates the doctor command.
	DoctorExamples = []Example{
		{
			Description: "Check system health",
			Command:     "dot doctor",
		},
		{
			Description: "Check specific packages",
			Command:     "dot doctor vim tmux",
		},
		{
			Description: "Output diagnosis in JSON format",
			Command:     "dot doctor --format json",
		},
	}

	// ListExamples demonstrates the list command.
	ListExamples = []Example{
		{
			Description: "List all available packages",
			Command:     "dot list",
		},
		{
			Description: "List with detailed information",
			Command:     "dot list --verbose",
		},
		{
			Description: "List in JSON format",
			Command:     "dot list --format json",
		},
		{
			Description: "Sort packages by name",
			Command:     "dot list --sort name",
		},
	}
)

// FormatExamples formats examples for display.
func FormatExamples(examples []Example) string {
	if len(examples) == 0 {
		return ""
	}

	result := "Examples:\n"
	for _, ex := range examples {
		result += "  # " + ex.Description + "\n"
		result += "  $ " + ex.Command + "\n"
		if ex.Output != "" {
			result += "  " + ex.Output + "\n"
		}
		result += "\n"
	}

	return result
}
