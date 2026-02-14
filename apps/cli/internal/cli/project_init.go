package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	projectInitForce   bool
	projectInitStdout  bool
	projectInitClaude  bool
	projectInitGemini  bool
	projectInitCodex   bool
	projectInitNoSpec  bool
	projectInitNoAgent bool
)

var projectInitCmd = &cobra.Command{
	Use:        "init",
	SuggestFor: []string{"setup", "create", "new"},
	Short:      "Initialize a project with agent configuration and spec files",
	Long: `Initialize sets up a project directory with agent configuration files and the
taskmd specification document in a single step.

By default, creates a CLAUDE.md agent config and TASKMD_SPEC.md. Use agent flags
to select which agent configs to generate.

If a file already exists and --force is not set, it is skipped with a warning.

Examples:
  taskmd init                        # Writes CLAUDE.md + TASKMD_SPEC.md
  taskmd init --gemini               # Writes GEMINI.md + TASKMD_SPEC.md
  taskmd init --claude --gemini      # Writes CLAUDE.md + GEMINI.md + TASKMD_SPEC.md
  taskmd init --no-spec              # Writes CLAUDE.md only
  taskmd init --no-agent             # Writes TASKMD_SPEC.md only
  taskmd init --force                # Overwrite existing files
  taskmd init --stdout               # Print all content to stdout
  taskmd init --dir ./my-project     # Write to a specific directory`,
	Args: cobra.NoArgs,
	RunE: runProjectInit,
}

func init() {
	rootCmd.AddCommand(projectInitCmd)

	projectInitCmd.Flags().BoolVar(&projectInitForce, "force", false, "overwrite existing files")
	projectInitCmd.Flags().BoolVar(&projectInitStdout, "stdout", false, "print all content to stdout instead of writing files")
	projectInitCmd.Flags().BoolVar(&projectInitClaude, "claude", false, "initialize for Claude Code")
	projectInitCmd.Flags().BoolVar(&projectInitGemini, "gemini", false, "initialize for Gemini")
	projectInitCmd.Flags().BoolVar(&projectInitCodex, "codex", false, "initialize for Codex")
	projectInitCmd.Flags().BoolVar(&projectInitNoSpec, "no-spec", false, "skip generating TASKMD_SPEC.md")
	projectInitCmd.Flags().BoolVar(&projectInitNoAgent, "no-agent", false, "skip generating agent configuration files")
}

// fileToWrite represents a file that the init command will create.
type fileToWrite struct {
	filename string
	content  []byte
}

func runProjectInit(_ *cobra.Command, _ []string) error {
	if projectInitNoSpec && projectInitNoAgent {
		return fmt.Errorf("--no-spec and --no-agent cannot both be set (nothing to do)")
	}

	files := collectFilesToWrite()

	if projectInitStdout {
		return printFilesToStdout(files)
	}

	targetDir := GetGlobalFlags().TaskDir

	info, err := os.Stat(targetDir)
	if err != nil {
		return fmt.Errorf("directory does not exist: %s", targetDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", targetDir)
	}

	return writeInitFiles(targetDir, files)
}

func writeInitFiles(targetDir string, files []fileToWrite) error {
	var created []string
	quiet := GetGlobalFlags().Quiet

	for _, f := range files {
		absPath, skipped, err := writeInitFile(targetDir, f)
		if err != nil {
			return err
		}
		if skipped {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Skipped %s (already exists, use --force to overwrite)\n", absPath)
			}
			continue
		}
		created = append(created, absPath)
	}

	if !quiet {
		for _, path := range created {
			fmt.Printf("Created %s\n", path)
		}
	}

	return nil
}

func writeInitFile(targetDir string, f fileToWrite) (absPath string, skipped bool, err error) {
	outputPath := filepath.Join(targetDir, f.filename)
	absPath, err = filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	if !projectInitForce {
		if _, err := os.Stat(outputPath); err == nil {
			return absPath, true, nil
		}
	}

	if err := os.WriteFile(outputPath, f.content, 0644); err != nil {
		return absPath, false, fmt.Errorf("failed to write %s: %w", f.filename, err)
	}

	return absPath, false, nil
}

func collectFilesToWrite() []fileToWrite {
	var files []fileToWrite

	if !projectInitNoAgent {
		agents := getProjectInitAgents()
		for _, agent := range agents {
			files = append(files, fileToWrite{
				filename: agent.filename,
				content:  agent.template,
			})
		}
	}

	if !projectInitNoSpec {
		files = append(files, fileToWrite{
			filename: specFilename,
			content:  specTemplate,
		})
	}

	return files
}

func getProjectInitAgents() []agentConfig {
	var agents []agentConfig

	// If no agent flags specified, default to Claude
	if !projectInitClaude && !projectInitGemini && !projectInitCodex {
		projectInitClaude = true
	}

	if projectInitClaude {
		agents = append(agents, agentConfig{
			name:     "Claude Code",
			filename: "CLAUDE.md",
			template: claudeTemplate,
		})
	}

	if projectInitGemini {
		agents = append(agents, agentConfig{
			name:     "Gemini",
			filename: "GEMINI.md",
			template: geminiTemplate,
		})
	}

	if projectInitCodex {
		agents = append(agents, agentConfig{
			name:     "Codex",
			filename: "AGENTS.md",
			template: codexTemplate,
		})
	}

	return agents
}

func printFilesToStdout(files []fileToWrite) error {
	for i, f := range files {
		if i > 0 {
			fmt.Print("\n---\n")
			fmt.Printf("# %s\n", f.filename)
			fmt.Print("---\n\n")
		}
		fmt.Print(string(f.content))
	}
	return nil
}
