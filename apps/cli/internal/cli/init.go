package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed templates/CLAUDE.md
var claudeTemplate []byte

//go:embed templates/GEMINI.md
var geminiTemplate []byte

//go:embed templates/CODEX.md
var codexTemplate []byte

var (
	initForce  bool
	initStdout bool
	initClaude bool
	initGemini bool
	initCodex  bool
)

type agentConfig struct {
	name     string
	filename string
	template []byte
}

var initCmd = &cobra.Command{
	Use:        "init",
	SuggestFor: []string{"setup", "create", "new"},
	Short:      "Initialize agent configuration files for taskmd",
	Long: `Initialize creates configuration files for AI agents in the target directory.

By default, creates a CLAUDE.md file for Claude Code. Gemini uses GEMINI.md,
and Codex uses AGENTS.md.

The templates include task file format documentation, common CLI commands,
and workflow guidelines.

Examples:
  taskmd agents init                    # Creates CLAUDE.md (default)
  taskmd agents init --claude           # Explicitly create CLAUDE.md
  taskmd agents init --gemini           # Create GEMINI.md
  taskmd agents init --codex            # Create AGENTS.md
  taskmd agents init --claude --gemini  # Create both CLAUDE.md and GEMINI.md
  taskmd agents init --gemini --codex   # Create GEMINI.md and AGENTS.md
  taskmd agents init --force            # Overwrite existing files
  taskmd agents init --stdout           # Print to stdout instead of files`,
	Args: cobra.NoArgs,
	RunE: runInit,
}

// Deprecated init command for backward compatibility
var deprecatedInitCmd = &cobra.Command{
	Use:        "init",
	SuggestFor: []string{"setup", "create", "new"},
	Short:      "Initialize taskmd (deprecated - use 'taskmd agents init')",
	Long:       `This command is deprecated. Please use 'taskmd agents init' instead.`,
	Args:       cobra.NoArgs,
	Deprecated: "use 'taskmd agents init' instead",
	RunE:       runDeprecatedInit,
}

func init() {
	agentsCmd.AddCommand(initCmd)
	rootCmd.AddCommand(deprecatedInitCmd)

	// Flags for agents init
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing configuration files")
	initCmd.Flags().BoolVar(&initStdout, "stdout", false, "print templates to stdout instead of writing files")
	initCmd.Flags().BoolVar(&initClaude, "claude", false, "initialize for Claude Code")
	initCmd.Flags().BoolVar(&initGemini, "gemini", false, "initialize for Gemini")
	initCmd.Flags().BoolVar(&initCodex, "codex", false, "initialize for Codex")

	// Flags for deprecated init (only the old ones for compatibility)
	deprecatedInitCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing CLAUDE.md")
	deprecatedInitCmd.Flags().BoolVar(&initStdout, "stdout", false, "print template to stdout instead of writing a file")
}

func runDeprecatedInit(_ *cobra.Command, _ []string) error {
	// Set Claude flag to true for backward compatibility
	initClaude = true
	initGemini = false
	initCodex = false

	return runInit(nil, nil)
}

func runInit(_ *cobra.Command, _ []string) error {
	// Determine which agents to initialize
	agents := getSelectedAgents()

	if initStdout {
		return printAgentsToStdout(agents)
	}

	targetDir := GetGlobalFlags().Dir

	// Verify target directory exists
	info, err := os.Stat(targetDir)
	if err != nil {
		return fmt.Errorf("directory does not exist: %s", targetDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", targetDir)
	}

	// Write config files for each agent
	for _, agent := range agents {
		if err := writeAgentConfig(targetDir, agent); err != nil {
			return err
		}
	}

	return nil
}

func getSelectedAgents() []agentConfig {
	agents := []agentConfig{}

	// If no flags specified, default to Claude
	if !initClaude && !initGemini && !initCodex {
		initClaude = true
	}

	if initClaude {
		agents = append(agents, agentConfig{
			name:     "Claude Code",
			filename: "CLAUDE.md",
			template: claudeTemplate,
		})
	}

	if initGemini {
		agents = append(agents, agentConfig{
			name:     "Gemini",
			filename: "GEMINI.md",
			template: geminiTemplate,
		})
	}

	if initCodex {
		agents = append(agents, agentConfig{
			name:     "Codex",
			filename: "AGENTS.md",
			template: codexTemplate,
		})
	}

	return agents
}

func printAgentsToStdout(agents []agentConfig) error {
	for i, agent := range agents {
		if i > 0 {
			fmt.Print("\n---\n")
			fmt.Printf("# %s\n", agent.filename)
			fmt.Print("---\n\n")
		}
		fmt.Print(string(agent.template))
	}
	return nil
}

func writeAgentConfig(targetDir string, agent agentConfig) error {
	outputPath := filepath.Join(targetDir, agent.filename)

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	// Check if file exists
	if !initForce {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("%s already exists at %s (use --force to overwrite)", agent.filename, absPath)
		}
	}

	// Write the file
	if err := os.WriteFile(outputPath, agent.template, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", agent.filename, err)
	}

	if !GetGlobalFlags().Quiet {
		fmt.Printf("Created %s\n", absPath)
	}

	return nil
}
