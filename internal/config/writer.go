package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// Writer handles writing configuration to files.
type Writer struct {
	path string
}

// NewWriter creates a configuration writer.
func NewWriter(path string) *Writer {
	return &Writer{
		path: path,
	}
}

// Write writes configuration to file.
func (w *Writer) Write(cfg *ExtendedConfig, opts WriteOptions) error {
	// Ensure directory exists
	dir := filepath.Dir(w.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Marshal config based on format
	data, err := w.marshal(cfg, opts)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(w.path, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// WriteDefault writes default configuration with comments.
func (w *Writer) WriteDefault(opts WriteOptions) error {
	cfg := DefaultExtended()
	opts.IncludeComments = opts.IncludeComments || opts.Format == "yaml"
	return w.Write(cfg, opts)
}

// Update updates specific value in configuration file.
func (w *Writer) Update(key string, value interface{}) error {
	// Load existing config
	var cfg *ExtendedConfig
	var err error

	if fileExists(w.path) {
		cfg, err = LoadExtendedFromFile(w.path)
		if err != nil {
			return fmt.Errorf("load existing config: %w", err)
		}
	} else {
		// File doesn't exist, create with default
		cfg = DefaultExtended()
	}

	// Update value
	if err := w.setValue(cfg, key, value); err != nil {
		return fmt.Errorf("set value: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Write back
	opts := WriteOptions{
		Format:          w.DetectFormat(),
		IncludeComments: false,
	}
	return w.Write(cfg, opts)
}

// WriteOptions controls configuration file output.
type WriteOptions struct {
	Format          string // yaml, json, toml
	IncludeComments bool
	Indent          int
}

// marshal converts config to bytes in specified format.
func (w *Writer) marshal(cfg *ExtendedConfig, opts WriteOptions) ([]byte, error) {
	format := opts.Format
	if format == "" {
		format = w.DetectFormat()
	}

	switch format {
	case "yaml", "yml":
		return w.marshalYAML(cfg, opts)
	case "json":
		return w.marshalJSON(cfg, opts)
	case "toml":
		return w.marshalTOML(cfg, opts)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// marshalYAML marshals config to YAML.
func (w *Writer) marshalYAML(cfg *ExtendedConfig, opts WriteOptions) ([]byte, error) {
	if opts.IncludeComments {
		return w.marshalYAMLWithComments(cfg)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal yaml: %w", err)
	}

	return data, nil
}

// marshalYAMLWithComments creates YAML with helpful comments.
func (w *Writer) marshalYAMLWithComments(cfg *ExtendedConfig) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("# Dot Configuration File\n")
	buf.WriteString("# Location: " + w.path + "\n")
	buf.WriteString("# Documentation: https://github.com/jamesainslie/dot/docs/configuration.md\n\n")

	buf.WriteString("# Core Directories\n")
	buf.WriteString("directories:\n")
	buf.WriteString("  # Stow directory containing packages\n")
	buf.WriteString(fmt.Sprintf("  stow: %s\n", cfg.Directories.Stow))
	buf.WriteString("  # Target directory for symlinks\n")
	buf.WriteString(fmt.Sprintf("  target: %s\n", cfg.Directories.Target))
	buf.WriteString("  # Manifest directory for tracking\n")
	buf.WriteString(fmt.Sprintf("  manifest: %s\n\n", cfg.Directories.Manifest))

	buf.WriteString("# Logging Configuration\n")
	buf.WriteString("logging:\n")
	buf.WriteString("  # Log level: DEBUG, INFO, WARN, ERROR\n")
	buf.WriteString(fmt.Sprintf("  level: %s\n", cfg.Logging.Level))
	buf.WriteString("  # Log format: text, json\n")
	buf.WriteString(fmt.Sprintf("  format: %s\n", cfg.Logging.Format))
	buf.WriteString("  # Log destination: stderr, stdout, file\n")
	buf.WriteString(fmt.Sprintf("  destination: %s\n", cfg.Logging.Destination))
	buf.WriteString("  # Log file path (only used if destination is file)\n")
	buf.WriteString(fmt.Sprintf("  file: %s\n\n", cfg.Logging.File))

	buf.WriteString("# Symlink Behavior\n")
	buf.WriteString("symlinks:\n")
	buf.WriteString("  # Link mode: relative, absolute\n")
	buf.WriteString(fmt.Sprintf("  mode: %s\n", cfg.Symlinks.Mode))
	buf.WriteString("  # Enable directory folding optimization\n")
	buf.WriteString(fmt.Sprintf("  folding: %t\n", cfg.Symlinks.Folding))
	buf.WriteString("  # Overwrite existing files when conflicts occur\n")
	buf.WriteString(fmt.Sprintf("  overwrite: %t\n", cfg.Symlinks.Overwrite))
	buf.WriteString("  # Create backup of overwritten files\n")
	buf.WriteString(fmt.Sprintf("  backup: %t\n", cfg.Symlinks.Backup))
	buf.WriteString("  # Backup suffix when backups enabled\n")
	buf.WriteString(fmt.Sprintf("  backup_suffix: %s\n\n", cfg.Symlinks.BackupSuffix))

	buf.WriteString("# Ignore Patterns\n")
	buf.WriteString("ignore:\n")
	buf.WriteString("  # Use default ignore patterns\n")
	buf.WriteString(fmt.Sprintf("  use_defaults: %t\n", cfg.Ignore.UseDefaults))
	buf.WriteString("  # Additional patterns to ignore (glob format)\n")
	w.writeYAMLList(&buf, "patterns", cfg.Ignore.Patterns, 2)
	buf.WriteString("  # Patterns to override (force include even if ignored)\n")
	w.writeYAMLList(&buf, "overrides", cfg.Ignore.Overrides, 2)
	buf.WriteString("\n")

	buf.WriteString("# Dotfile Translation\n")
	buf.WriteString("dotfile:\n")
	buf.WriteString("  # Enable dot- to . translation\n")
	buf.WriteString(fmt.Sprintf("  translate: %t\n", cfg.Dotfile.Translate))
	buf.WriteString("  # Prefix for dotfile translation\n")
	buf.WriteString(fmt.Sprintf("  prefix: %s\n\n", cfg.Dotfile.Prefix))

	buf.WriteString("# Output Configuration\n")
	buf.WriteString("output:\n")
	buf.WriteString("  # Default output format: text, json, yaml, table\n")
	buf.WriteString(fmt.Sprintf("  format: %s\n", cfg.Output.Format))
	buf.WriteString("  # Enable colored output: auto, always, never\n")
	buf.WriteString(fmt.Sprintf("  color: %s\n", cfg.Output.Color))
	buf.WriteString("  # Show progress indicators\n")
	buf.WriteString(fmt.Sprintf("  progress: %t\n", cfg.Output.Progress))
	buf.WriteString("  # Verbosity level: 0 (quiet), 1 (normal), 2 (verbose), 3 (debug)\n")
	buf.WriteString(fmt.Sprintf("  verbosity: %d\n", cfg.Output.Verbosity))
	buf.WriteString("  # Terminal width for text wrapping (0 = auto-detect)\n")
	buf.WriteString(fmt.Sprintf("  width: %d\n\n", cfg.Output.Width))

	buf.WriteString("# Operation Defaults\n")
	buf.WriteString("operations:\n")
	buf.WriteString("  # Enable dry-run mode by default\n")
	buf.WriteString(fmt.Sprintf("  dry_run: %t\n", cfg.Operations.DryRun))
	buf.WriteString("  # Enable atomic operations with rollback\n")
	buf.WriteString(fmt.Sprintf("  atomic: %t\n", cfg.Operations.Atomic))
	buf.WriteString("  # Maximum number of parallel operations (0 = auto)\n")
	buf.WriteString(fmt.Sprintf("  max_parallel: %d\n\n", cfg.Operations.MaxParallel))

	buf.WriteString("# Package Management\n")
	buf.WriteString("packages:\n")
	buf.WriteString("  # Default sort order: name, links, date\n")
	buf.WriteString(fmt.Sprintf("  sort_by: %s\n", cfg.Packages.SortBy))
	buf.WriteString("  # Automatically scan for new packages\n")
	buf.WriteString(fmt.Sprintf("  auto_discover: %t\n", cfg.Packages.AutoDiscover))
	buf.WriteString("  # Package naming convention validation\n")
	buf.WriteString(fmt.Sprintf("  validate_names: %t\n\n", cfg.Packages.ValidateNames))

	buf.WriteString("# Doctor Configuration\n")
	buf.WriteString("doctor:\n")
	buf.WriteString("  # Auto-fix issues when possible\n")
	buf.WriteString(fmt.Sprintf("  auto_fix: %t\n", cfg.Doctor.AutoFix))
	buf.WriteString("  # Check manifest integrity\n")
	buf.WriteString(fmt.Sprintf("  check_manifest: %t\n", cfg.Doctor.CheckManifest))
	buf.WriteString("  # Check for broken symlinks\n")
	buf.WriteString(fmt.Sprintf("  check_broken_links: %t\n", cfg.Doctor.CheckBrokenLinks))
	buf.WriteString("  # Check for orphaned links\n")
	buf.WriteString(fmt.Sprintf("  check_orphaned: %t\n", cfg.Doctor.CheckOrphaned))
	buf.WriteString("  # Check file permissions\n")
	buf.WriteString(fmt.Sprintf("  check_permissions: %t\n\n", cfg.Doctor.CheckPermissions))

	buf.WriteString("# Experimental Features\n")
	buf.WriteString("experimental:\n")
	buf.WriteString("  # Enable parallel operations\n")
	buf.WriteString(fmt.Sprintf("  parallel: %t\n", cfg.Experimental.Parallel))
	buf.WriteString("  # Enable performance profiling\n")
	buf.WriteString(fmt.Sprintf("  profiling: %t\n", cfg.Experimental.Profiling))

	return buf.Bytes(), nil
}

// writeYAMLList writes a YAML list with proper indentation.
func (w *Writer) writeYAMLList(buf *bytes.Buffer, key string, items []string, indent int) {
	prefix := strings.Repeat(" ", indent)

	if len(items) == 0 {
		buf.WriteString(fmt.Sprintf("%s%s: []\n", prefix, key))
		return
	}

	buf.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
	for _, item := range items {
		buf.WriteString(fmt.Sprintf("%s  - %q\n", prefix, item))
	}
}

// marshalJSON marshals config to JSON.
func (w *Writer) marshalJSON(cfg *ExtendedConfig, opts WriteOptions) ([]byte, error) {
	indent := opts.Indent
	if indent == 0 {
		indent = 2
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", strings.Repeat(" ", indent))

	if err := encoder.Encode(cfg); err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	return buf.Bytes(), nil
}

// marshalTOML marshals config to TOML.
func (w *Writer) marshalTOML(cfg *ExtendedConfig, opts WriteOptions) ([]byte, error) {
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)

	if err := encoder.Encode(cfg); err != nil {
		return nil, fmt.Errorf("encode toml: %w", err)
	}

	return buf.Bytes(), nil
}

// DetectFormat detects format from file extension.
func (w *Writer) DetectFormat() string {
	ext := filepath.Ext(w.path)
	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".toml":
		return "toml"
	default:
		return "yaml"
	}
}

// setValue sets a configuration value by dotted key path.
func (w *Writer) setValue(cfg *ExtendedConfig, key string, value interface{}) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid key: %s (must be section.field)", key)
	}

	section := parts[0]
	field := parts[1]

	switch section {
	case "directories":
		return setDirectoriesValue(&cfg.Directories, field, value)
	case "logging":
		return setLoggingValue(&cfg.Logging, field, value)
	case "symlinks":
		return setSymlinksValue(&cfg.Symlinks, field, value)
	case "ignore":
		return setIgnoreValue(&cfg.Ignore, field, value)
	case "dotfile":
		return setDotfileValue(&cfg.Dotfile, field, value)
	case "output":
		return setOutputValue(&cfg.Output, field, value)
	case "operations":
		return setOperationsValue(&cfg.Operations, field, value)
	case "packages":
		return setPackagesValue(&cfg.Packages, field, value)
	case "doctor":
		return setDoctorValue(&cfg.Doctor, field, value)
	case "experimental":
		return setExperimentalValue(&cfg.Experimental, field, value)
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
}

func setDirectoriesValue(cfg *DirectoriesConfig, field string, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("directories.%s: value must be string", field)
	}

	switch field {
	case "stow":
		cfg.Stow = str
	case "target":
		cfg.Target = str
	case "manifest":
		cfg.Manifest = str
	default:
		return fmt.Errorf("unknown field: directories.%s", field)
	}

	return nil
}

func setLoggingValue(cfg *LoggingConfig, field string, value interface{}) error {
	switch field {
	case "level", "format", "destination", "file":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("logging.%s: value must be string", field)
		}

		switch field {
		case "level":
			cfg.Level = str
		case "format":
			cfg.Format = str
		case "destination":
			cfg.Destination = str
		case "file":
			cfg.File = str
		}
	default:
		return fmt.Errorf("unknown field: logging.%s", field)
	}

	return nil
}

func setSymlinksValue(cfg *SymlinksConfig, field string, value interface{}) error {
	switch field {
	case "mode", "backup_suffix":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("symlinks.%s: value must be string", field)
		}

		switch field {
		case "mode":
			cfg.Mode = str
		case "backup_suffix":
			cfg.BackupSuffix = str
		}

	case "folding", "overwrite", "backup":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("symlinks.%s: value must be bool", field)
		}

		switch field {
		case "folding":
			cfg.Folding = b
		case "overwrite":
			cfg.Overwrite = b
		case "backup":
			cfg.Backup = b
		}

	default:
		return fmt.Errorf("unknown field: symlinks.%s", field)
	}

	return nil
}

func setIgnoreValue(cfg *IgnoreConfig, field string, value interface{}) error {
	switch field {
	case "use_defaults":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ignore.%s: value must be bool", field)
		}
		cfg.UseDefaults = b

	case "patterns", "overrides":
		// Accept both []string and string
		var arr []string
		switch v := value.(type) {
		case []string:
			arr = v
		case string:
			// Split comma-separated string
			arr = strings.Split(v, ",")
			for i := range arr {
				arr[i] = strings.TrimSpace(arr[i])
			}
		default:
			return fmt.Errorf("ignore.%s: value must be []string or string", field)
		}

		switch field {
		case "patterns":
			cfg.Patterns = arr
		case "overrides":
			cfg.Overrides = arr
		}

	default:
		return fmt.Errorf("unknown field: ignore.%s", field)
	}

	return nil
}

func setDotfileValue(cfg *DotfileConfig, field string, value interface{}) error {
	switch field {
	case "translate":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("dotfile.%s: value must be bool", field)
		}
		cfg.Translate = b

	case "prefix":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("dotfile.%s: value must be string", field)
		}
		cfg.Prefix = str

	default:
		return fmt.Errorf("unknown field: dotfile.%s", field)
	}

	return nil
}

func setOutputValue(cfg *OutputConfig, field string, value interface{}) error {
	switch field {
	case "format", "color":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("output.%s: value must be string", field)
		}

		switch field {
		case "format":
			cfg.Format = str
		case "color":
			cfg.Color = str
		}

	case "progress":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("output.%s: value must be bool", field)
		}
		cfg.Progress = b

	case "verbosity", "width":
		var i int
		switch v := value.(type) {
		case int:
			i = v
		case float64:
			i = int(v)
		default:
			return fmt.Errorf("output.%s: value must be int", field)
		}

		switch field {
		case "verbosity":
			cfg.Verbosity = i
		case "width":
			cfg.Width = i
		}

	default:
		return fmt.Errorf("unknown field: output.%s", field)
	}

	return nil
}

func setOperationsValue(cfg *OperationsConfig, field string, value interface{}) error {
	switch field {
	case "dry_run", "atomic":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("operations.%s: value must be bool", field)
		}

		switch field {
		case "dry_run":
			cfg.DryRun = b
		case "atomic":
			cfg.Atomic = b
		}

	case "max_parallel":
		var i int
		switch v := value.(type) {
		case int:
			i = v
		case float64:
			i = int(v)
		default:
			return fmt.Errorf("operations.%s: value must be int", field)
		}
		cfg.MaxParallel = i

	default:
		return fmt.Errorf("unknown field: operations.%s", field)
	}

	return nil
}

func setPackagesValue(cfg *PackagesConfig, field string, value interface{}) error {
	switch field {
	case "sort_by":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("packages.%s: value must be string", field)
		}
		cfg.SortBy = str

	case "auto_discover", "validate_names":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("packages.%s: value must be bool", field)
		}

		switch field {
		case "auto_discover":
			cfg.AutoDiscover = b
		case "validate_names":
			cfg.ValidateNames = b
		}

	default:
		return fmt.Errorf("unknown field: packages.%s", field)
	}

	return nil
}

func setDoctorValue(cfg *DoctorConfig, field string, value interface{}) error {
	b, ok := value.(bool)
	if !ok {
		return fmt.Errorf("doctor.%s: value must be bool", field)
	}

	switch field {
	case "auto_fix":
		cfg.AutoFix = b
	case "check_manifest":
		cfg.CheckManifest = b
	case "check_broken_links":
		cfg.CheckBrokenLinks = b
	case "check_orphaned":
		cfg.CheckOrphaned = b
	case "check_permissions":
		cfg.CheckPermissions = b
	default:
		return fmt.Errorf("unknown field: doctor.%s", field)
	}

	return nil
}

func setExperimentalValue(cfg *ExperimentalConfig, field string, value interface{}) error {
	b, ok := value.(bool)
	if !ok {
		return fmt.Errorf("experimental.%s: value must be bool", field)
	}

	switch field {
	case "parallel":
		cfg.Parallel = b
	case "profiling":
		cfg.Profiling = b
	default:
		return fmt.Errorf("unknown field: experimental.%s", field)
	}

	return nil
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
