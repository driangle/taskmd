package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version information (set via build flags)
	Version   = "0.0.3"
	GitCommit = "unknown"
	BuildDate = "unknown"

	// Global flags
	cfgFile string
	stdin   bool
	quiet   bool
	verbose bool
	debug   bool
	noColor bool
	dir     string
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

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

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
	rootCmd.PersistentFlags().StringVarP(&dir, "dir", "d", ".", "task directory to scan")

	// Bind flags to viper
	viper.BindPFlag("stdin", rootCmd.PersistentFlags().Lookup("stdin"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
	viper.BindPFlag("dir", rootCmd.PersistentFlags().Lookup("dir"))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Add current directory (project-level config takes precedence)
		viper.AddConfigPath(".")

		// Add home directory (global config)
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}

		viper.SetConfigType("yaml")
		viper.SetConfigName(".taskmd")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("TASKMD")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// GetGlobalFlags returns a struct with all global flag values
func GetGlobalFlags() GlobalFlags {
	// Try to get from viper first (supports config files)
	// Fall back to flag variables if viper doesn't have the value (for tests)
	dirVal := viper.GetString("dir")
	if dirVal == "" && dir != "" {
		dirVal = dir
	} else if dirVal == "" {
		dirVal = "."
	}

	return GlobalFlags{
		Stdin:   viper.GetBool("stdin") || stdin,
		Quiet:   viper.GetBool("quiet") || quiet,
		Verbose: viper.GetBool("verbose") || verbose,
		Debug:   viper.GetBool("debug") || debug,
		NoColor: viper.GetBool("no-color") || noColor,
		Dir:     dirVal,
	}
}

// GlobalFlags holds global flag values
type GlobalFlags struct {
	Stdin   bool
	Quiet   bool
	Verbose bool
	Debug   bool
	NoColor bool
	Dir     string
}

// ResolveScanDir returns the scan directory from positional arg or --dir flag.
// Positional arg takes precedence for backward compatibility.
func ResolveScanDir(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return GetGlobalFlags().Dir
}
