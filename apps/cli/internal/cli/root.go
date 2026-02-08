package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version information (set via build flags)
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"

	// Global flags
	cfgFile string
	stdin   bool
	format  string
	quiet   bool
	verbose bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "taskmd",
	Short: "A markdown-based task tracker CLI",
	Long: `taskmd is a command-line tool for managing tasks stored in markdown files.
It supports reading from files or stdin, multiple output formats, and various
commands for listing, validating, and visualizing your tasks.`,
	SilenceUsage:  true,
	SilenceErrors: false,
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
	rootCmd.PersistentFlags().StringVar(&format, "format", "table", "output format (table, json, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")

	// Bind flags to viper
	viper.BindPFlag("stdin", rootCmd.PersistentFlags().Lookup("stdin"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}
			return
		}

		// Search config in home directory with name ".taskmd" (without extension)
		viper.AddConfigPath(home)
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
	return GlobalFlags{
		Stdin:   stdin,
		Format:  format,
		Quiet:   quiet,
		Verbose: verbose,
	}
}

// GlobalFlags holds global flag values
type GlobalFlags struct {
	Stdin   bool
	Format  string
	Quiet   bool
	Verbose bool
}
