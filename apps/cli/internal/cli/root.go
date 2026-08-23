package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/validator"
)

var (
	// Version information (set via build flags)
	Version   = "0.4.1"
	GitCommit = "unknown"
	BuildDate = "unknown"
	GitDirty  = ""

	// Global flags
	cfgFile         string
	stdin           bool
	quiet           bool
	verbose         bool
	debug           bool
	noColor         bool
	taskDir         string
	projectFlag     string
	allProjectsFlag bool
	worktreesFlag   string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "taskmd",
	Short: "A markdown-based task tracker CLI",
	Long: `taskmd is a command-line tool for managing tasks stored in markdown files.
It supports reading from files or stdin, multiple output formats, and various
commands for listing, validating, and visualizing your tasks.

Exit codes:
  0 - Success
  1 - Error (invalid input, scan failure, etc.)
  2 - Validation warnings (with --strict)`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       Version,
}

// FullVersion returns the display version string.
// Examples: "0.0.3", "0.0.3-abc1234", "0.0.3-abc1234*"
func FullVersion() string {
	v := Version
	if GitCommit != "unknown" && GitCommit != "" {
		short := GitCommit
		if len(short) > 7 {
			short = short[:7]
		}
		v += "-" + short
	}
	if GitDirty == "true" {
		v += "*"
	}
	return v
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Cobra short-circuits to the help renderer before OnInitialize runs, so
	// config-derived usage text (the --effort vocabulary) would print stale
	// defaults. Load config and refresh it here instead. Subcommands inherit
	// this help func from the root.
	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		initConfig()
		refreshEffortUsage(cmd)
		defaultHelp(cmd, args)
	})

	// Set version template with detailed info
	versionTemplate := fmt.Sprintf("taskmd version %s\n  Git commit: %s\n  Built:      %s\n", Version, GitCommit, BuildDate)
	rootCmd.SetVersionTemplate(versionTemplate)

	// Global flags available to all subcommands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.taskmd.yaml)")
	rootCmd.PersistentFlags().BoolVar(&stdin, "stdin", false, "read input from stdin instead of file")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output (prints to stderr)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringVarP(&taskDir, "task-dir", "d", ".", "task directory to scan")

	rootCmd.PersistentFlags().StringVar(&worktreesFlag, "worktrees", worktreeModeAuto, "cross-worktree overlay: auto (active in multi-worktree repos), true, or false")

	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "", "operate on a registered project by id")
	rootCmd.PersistentFlags().BoolVar(&allProjectsFlag, "all-projects", false, "aggregate tasks from all registered projects")
	rootCmd.MarkFlagsMutuallyExclusive("project", "all-projects")

	// Deprecated alias: --dir still works but is hidden
	rootCmd.PersistentFlags().StringVar(&taskDir, "dir", "", "task directory (deprecated: use --task-dir)")
	_ = rootCmd.PersistentFlags().MarkHidden("dir")

	// Bind flags to viper
	viper.BindPFlag("stdin", rootCmd.PersistentFlags().Lookup("stdin"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
	viper.BindPFlag("task-dir", rootCmd.PersistentFlags().Lookup("task-dir"))
	viper.BindPFlag("worktrees", rootCmd.PersistentFlags().Lookup("worktrees"))
	viper.BindPFlag("dir", rootCmd.PersistentFlags().Lookup("dir"))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Walk up from cwd, adding each ancestor as a config search path.
		// Nearest directory is added first (highest priority in viper).
		// Stop at .git boundary or filesystem root.
		dir, err := filepath.Abs(".")
		if err == nil {
			for {
				viper.AddConfigPath(dir)
				// Stop if we hit a .git directory (project boundary)
				if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
					break
				}
				parent := filepath.Dir(dir)
				if parent == dir {
					break
				}
				dir = parent
			}
		}

		// Add home directory as final fallback
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}

		viper.SetConfigType("yaml")
		viper.SetConfigName(".taskmd")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("TASKMD")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// GetGlobalFlags returns a struct with all global flag values
func GetGlobalFlags() GlobalFlags {
	// Resolve task directory with proper precedence:
	// 1. Explicit --task-dir or --dir CLI flag (if changed from default)
	// 2. Config file value (supports both "task-dir" and "dir" keys)
	// 3. taskDir variable (for tests that set it directly)
	// 4. Default "."
	dirVal := resolveTaskDir()

	return GlobalFlags{
		Stdin:      viper.GetBool("stdin") || stdin,
		Quiet:      viper.GetBool("quiet") || quiet,
		Verbose:    viper.GetBool("verbose") || verbose,
		Debug:      viper.GetBool("debug") || debug,
		NoColor:    viper.GetBool("no-color") || noColor,
		TaskDir:    dirVal,
		IgnoreDirs: viper.GetStringSlice("ignore"),
		Workflow:   resolveWorkflow(),
	}
}

// resolveRelativeToConfig resolves a relative path against the config file's directory.
// If the path is absolute, or no config file is loaded, it is returned unchanged.
func resolveRelativeToConfig(dir string) string {
	if filepath.IsAbs(dir) {
		return dir
	}
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return dir
	}
	configDir, err := filepath.Abs(filepath.Dir(configFile))
	if err != nil {
		return dir
	}
	return filepath.Join(configDir, dir)
}

// resolveTaskDir determines the task directory using proper precedence.
func resolveTaskDir() string {
	// --project flag takes highest precedence
	if projectFlag != "" {
		dir, err := resolveProjectDir(projectFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return dir
	}

	// Check if --task-dir or --dir was explicitly passed on the CLI
	taskDirFlag := rootCmd.PersistentFlags().Lookup("task-dir")
	dirFlag := rootCmd.PersistentFlags().Lookup("dir")

	if taskDirFlag != nil && taskDirFlag.Changed {
		return taskDirFlag.Value.String()
	}
	if dirFlag != nil && dirFlag.Changed {
		return dirFlag.Value.String()
	}

	// Check config file: support both "task-dir" and "dir" YAML keys.
	// We must bypass viper's pflag binding (which returns the flag default)
	// by checking the config file values directly via viper.InConfig.
	if viper.InConfig("task-dir") {
		return resolveRelativeToConfig(viper.GetString("task-dir"))
	}
	if viper.InConfig("dir") {
		return resolveRelativeToConfig(viper.GetString("dir"))
	}

	// Fall back to the taskDir variable (set directly in tests)
	if taskDir != "" {
		return taskDir
	}

	// No local config found — check for default_project in global config
	if dir := resolveDefaultProject(); dir != "" {
		return dir
	}

	return "."
}

// GlobalFlags holds global flag values
type GlobalFlags struct {
	Stdin      bool
	Quiet      bool
	Verbose    bool
	Debug      bool
	NoColor    bool
	TaskDir    string
	IgnoreDirs []string
	Workflow   string
}

// resolveWorkflow returns the configured workflow mode ("solo" or "pr-review").
// Defaults to "solo" when not set.
func resolveWorkflow() string {
	if w := viper.GetString("workflow"); w != "" {
		return w
	}
	return "solo"
}

// ResolveScanDir returns the scan directory from positional arg or --task-dir flag.
// Positional arg takes precedence for backward compatibility.
func ResolveScanDir(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return GetGlobalFlags().TaskDir
}

// resolveProjectDir looks up a project in the global registry and returns its task directory.
// It also reloads viper config from the project's .taskmd.yaml.
func resolveProjectDir(projectID string) (string, error) {
	entries, err := LoadGlobalRegistry()
	if err != nil {
		return "", fmt.Errorf("load global registry: %w", err)
	}

	entry, found := findProjectEntry(entries, projectID)
	if !found {
		return "", fmt.Errorf("project %q not found in global registry", projectID)
	}

	info, err := os.Stat(entry.Path)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("project %q path does not exist: %s", projectID, entry.Path)
	}

	projectPath := entry.Path
	// When the cwd is inside another worktree of the registered repo, scope to
	// the current checkout's files, never the primary's (ADR 0005 §1).
	if local := localWorktreePathFor(projectPath); local != "" {
		projectPath = local
	}

	return resolveProjectTaskDir(projectPath)
}

// findProjectEntry searches for a project by ID in the registry entries.
func findProjectEntry(entries []GlobalProjectEntry, id string) (GlobalProjectEntry, bool) {
	for _, e := range entries {
		if e.ID == id {
			return e, true
		}
	}
	return GlobalProjectEntry{}, false
}

// resolveProjectTaskDir loads the project's .taskmd.yaml and returns the task directory.
func resolveProjectTaskDir(projectPath string) (string, error) {
	projectConfig := filepath.Join(projectPath, ".taskmd.yaml")
	if _, err := os.Stat(projectConfig); err == nil {
		viper.SetConfigFile(projectConfig)
		if err := viper.ReadInConfig(); err != nil {
			return "", fmt.Errorf("read project config: %w", err)
		}
	}

	if viper.InConfig("task-dir") {
		return resolveRelativeToConfig(viper.GetString("task-dir")), nil
	}
	if viper.InConfig("dir") {
		return resolveRelativeToConfig(viper.GetString("dir")), nil
	}

	return projectPath, nil
}

// resolveDefaultProject checks the global config for a default_project setting
// and resolves it to a task directory. Returns empty string if not set or invalid.
func resolveDefaultProject() string {
	defaultID := LoadDefaultProject()
	if defaultID == "" {
		return ""
	}

	dir, err := resolveProjectDir(defaultID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: default_project %q: %v\n", defaultID, err)
		return ""
	}
	return dir
}

// resolveIDConfig returns the ID generation config with defaults applied.
func resolveIDConfig() validator.IDConfig {
	cfg := validator.IDConfig{
		Strategy: "sequential",
		Length:   6,
		Padding:  3,
	}

	raw := viper.Get("id")
	if raw == nil {
		return cfg
	}
	parsed := parseIDConfig(raw)
	if parsed == nil {
		return cfg
	}

	if parsed.Strategy != "" {
		cfg.Strategy = parsed.Strategy
	}
	if parsed.Prefix != "" {
		cfg.Prefix = parsed.Prefix
	}
	if parsed.Length != 0 {
		cfg.Length = parsed.Length
	}
	if parsed.Padding != 0 {
		cfg.Padding = parsed.Padding
	}

	return cfg
}
