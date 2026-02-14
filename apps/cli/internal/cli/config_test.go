package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// resetViper clears viper state for a clean test environment
func resetViper() {
	viper.Reset()
}

// resetFlags resets global flag variables to their defaults
func resetFlags() {
	cfgFile = ""
	stdin = false
	quiet = false
	verbose = false
	taskDir = "."
	webPort = 8080
	webDev = false
	webOpen = false
}

// createConfigFile creates a .taskmd.yaml file with the given content
func createConfigFile(t *testing.T, dir, content string) string {
	t.Helper()
	configPath := filepath.Join(dir, ".taskmd.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}
	return configPath
}

func TestConfigFile_ProjectLevel(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper() // Clean up after test
	defer resetFlags()

	// Create a temporary project directory with config
	projectDir := t.TempDir()
	createConfigFile(t, projectDir, `
dir: ./my-tasks
verbose: true
`)

	// Change to project directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Re-initialize cobra command to pick up new config
	rootCmd = &cobra.Command{Use: "taskmd"}
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&stdin, "stdin", false, "read from stdin")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().StringVarP(&taskDir, "task-dir", "d", ".", "task directory")

	// Bind flags to viper
	viper.BindPFlag("stdin", rootCmd.PersistentFlags().Lookup("stdin"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("task-dir", rootCmd.PersistentFlags().Lookup("task-dir"))

	// Initialize config
	initConfig()

	// Test that config values are loaded
	flags := GetGlobalFlags()
	if flags.TaskDir != "./my-tasks" {
		t.Errorf("expected dir to be './my-tasks', got '%s'", flags.TaskDir)
	}
	if !flags.Verbose {
		t.Error("expected verbose to be true from config")
	}
}

func TestConfigFile_GlobalLevel(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper()
	defer resetFlags()

	// Create a temporary home directory with config
	homeDir := t.TempDir()
	createConfigFile(t, homeDir, `
dir: ./global-tasks
verbose: true
`)

	// Set HOME to temp directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	// Create a different working directory without config
	workDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Re-initialize cobra command
	rootCmd = &cobra.Command{Use: "taskmd"}
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&stdin, "stdin", false, "read from stdin")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().StringVarP(&taskDir, "task-dir", "d", ".", "task directory")

	// Bind flags to viper
	viper.BindPFlag("stdin", rootCmd.PersistentFlags().Lookup("stdin"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("task-dir", rootCmd.PersistentFlags().Lookup("task-dir"))

	// Initialize config
	initConfig()

	// Test that global config values are loaded
	flags := GetGlobalFlags()
	if flags.TaskDir != "./global-tasks" {
		t.Errorf("expected dir to be './global-tasks', got '%s'", flags.TaskDir)
	}
	if !flags.Verbose {
		t.Error("expected verbose to be true from global config")
	}
}

func TestConfigFile_Precedence_ProjectOverGlobal(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper()
	defer resetFlags()

	// Create global config
	homeDir := t.TempDir()
	createConfigFile(t, homeDir, `
dir: ./global-tasks
verbose: true
`)

	// Create project config
	projectDir := t.TempDir()
	createConfigFile(t, projectDir, `
dir: ./project-tasks
verbose: false
`)

	// Set HOME and working directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Re-initialize cobra command
	rootCmd = &cobra.Command{Use: "taskmd"}
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&stdin, "stdin", false, "read from stdin")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().StringVarP(&taskDir, "task-dir", "d", ".", "task directory")

	// Bind flags to viper
	viper.BindPFlag("stdin", rootCmd.PersistentFlags().Lookup("stdin"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("task-dir", rootCmd.PersistentFlags().Lookup("task-dir"))

	// Initialize config
	initConfig()

	// Test that project config takes precedence
	flags := GetGlobalFlags()
	if flags.TaskDir != "./project-tasks" {
		t.Errorf("expected project config dir './project-tasks', got '%s'", flags.TaskDir)
	}
	if flags.Verbose {
		t.Error("expected verbose to be false from project config (not true from global)")
	}
}

func TestConfigFile_Precedence_FlagOverConfig(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper()
	defer resetFlags()

	// Create project config
	projectDir := t.TempDir()
	createConfigFile(t, projectDir, `
dir: ./config-tasks
verbose: true
`)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Re-initialize cobra command
	rootCmd = &cobra.Command{Use: "taskmd"}
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&stdin, "stdin", false, "read from stdin")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().StringVarP(&taskDir, "task-dir", "d", ".", "task directory")

	// Bind flags to viper
	viper.BindPFlag("stdin", rootCmd.PersistentFlags().Lookup("stdin"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("task-dir", rootCmd.PersistentFlags().Lookup("task-dir"))

	// Set flag explicitly (simulating CLI flag)
	rootCmd.PersistentFlags().Set("task-dir", "./flag-tasks")
	rootCmd.PersistentFlags().Set("verbose", "false")

	// Initialize config
	initConfig()

	// Test that CLI flags override config
	flags := GetGlobalFlags()
	if flags.TaskDir != "./flag-tasks" {
		t.Errorf("expected flag dir './flag-tasks', got '%s'", flags.TaskDir)
	}
	if flags.Verbose {
		t.Error("expected verbose to be false from flag (not true from config)")
	}
}

func TestConfigFile_WebOptions(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper()
	defer resetFlags()

	// Create project config with web options
	projectDir := t.TempDir()
	createConfigFile(t, projectDir, `
web:
  port: 3000
  auto_open_browser: true
`)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Re-initialize cobra command
	rootCmd = &cobra.Command{Use: "taskmd"}
	webCmd = &cobra.Command{Use: "web"}
	webStartCmd = &cobra.Command{Use: "start"}
	rootCmd.AddCommand(webCmd)
	webCmd.AddCommand(webStartCmd)

	cobra.OnInitialize(initConfig)

	// Setup web flags
	webStartCmd.Flags().IntVar(&webPort, "port", 8080, "server port")
	webStartCmd.Flags().BoolVar(&webDev, "dev", false, "enable dev mode")
	webStartCmd.Flags().BoolVar(&webOpen, "open", false, "open browser on start")

	// Bind web flags to viper
	viper.BindPFlag("web.port", webStartCmd.Flags().Lookup("port"))
	viper.BindPFlag("web.auto_open_browser", webStartCmd.Flags().Lookup("open"))

	// Initialize config
	initConfig()

	// Test that web config values are loaded
	port := viper.GetInt("web.port")
	open := viper.GetBool("web.auto_open_browser")

	if port != 3000 {
		t.Errorf("expected web.port to be 3000, got %d", port)
	}
	if !open {
		t.Error("expected web.auto_open_browser to be true from config")
	}
}

func TestConfigFile_Defaults(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper()
	defer resetFlags()

	// Create empty temp directory (no config file)
	workDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Set HOME to directory without config
	homeDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	// Re-initialize cobra command
	rootCmd = &cobra.Command{Use: "taskmd"}
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&stdin, "stdin", false, "read from stdin")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().StringVarP(&taskDir, "task-dir", "d", ".", "task directory")

	// Bind flags to viper
	viper.BindPFlag("stdin", rootCmd.PersistentFlags().Lookup("stdin"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("task-dir", rootCmd.PersistentFlags().Lookup("task-dir"))

	// Initialize config (no config file found)
	initConfig()

	// Test that defaults are used
	flags := GetGlobalFlags()
	if flags.TaskDir != "." {
		t.Errorf("expected default dir '.', got '%s'", flags.TaskDir)
	}
	if flags.Verbose {
		t.Error("expected default verbose to be false")
	}
}

func TestConfigFile_ExplicitConfigFile(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper()
	defer resetFlags()

	// Create a custom config file in a specific location
	configDir := t.TempDir()
	customConfigPath := filepath.Join(configDir, "custom-config.yaml")
	if err := os.WriteFile(customConfigPath, []byte(`
dir: ./custom-tasks
verbose: true
`), 0644); err != nil {
		t.Fatalf("failed to create custom config file: %v", err)
	}

	// Re-initialize cobra command
	rootCmd = &cobra.Command{Use: "taskmd"}
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&stdin, "stdin", false, "read from stdin")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().StringVarP(&taskDir, "task-dir", "d", ".", "task directory")

	// Bind flags to viper
	viper.BindPFlag("stdin", rootCmd.PersistentFlags().Lookup("stdin"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("task-dir", rootCmd.PersistentFlags().Lookup("task-dir"))

	// Set the config flag using cobra
	rootCmd.PersistentFlags().Set("config", customConfigPath)

	// Initialize config
	initConfig()

	// Test that custom config file is loaded
	flags := GetGlobalFlags()
	if flags.TaskDir != "./custom-tasks" {
		t.Errorf("expected dir to be './custom-tasks', got '%s'", flags.TaskDir)
	}
	if !flags.Verbose {
		t.Error("expected verbose to be true from custom config")
	}
}
