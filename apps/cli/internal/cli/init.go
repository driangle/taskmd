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

var (
	initForce  bool
	initStdout bool
)

var initCmd = &cobra.Command{
	Use:        "init",
	SuggestFor: []string{"setup", "create", "new"},
	Short:      "Initialize taskmd in a project by creating a CLAUDE.md template",
	Long: `Initialize creates a CLAUDE.md file in the target directory with a template
that helps Claude Code understand and work with your taskmd task files.

The template includes task file format documentation, common CLI commands,
and workflow guidelines.

Examples:
  taskmd init
  taskmd init --dir ./my-project
  taskmd init --force
  taskmd init --stdout`,
	Args: cobra.NoArgs,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing CLAUDE.md")
	initCmd.Flags().BoolVar(&initStdout, "stdout", false, "print template to stdout instead of writing a file")
}

func runInit(cmd *cobra.Command, args []string) error {
	if initStdout {
		fmt.Print(string(claudeTemplate))
		return nil
	}

	targetDir := GetGlobalFlags().Dir
	outputPath := filepath.Join(targetDir, "CLAUDE.md")

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	info, err := os.Stat(targetDir)
	if err != nil {
		return fmt.Errorf("directory does not exist: %s", targetDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", targetDir)
	}

	if !initForce {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("CLAUDE.md already exists at %s (use --force to overwrite)", absPath)
		}
	}

	if err := os.WriteFile(outputPath, claudeTemplate, 0644); err != nil {
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}

	if !GetGlobalFlags().Quiet {
		fmt.Printf("Created %s\n", absPath)
	}

	return nil
}
