package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed templates/TASKMD_SPEC.md
var specTemplate []byte

const specFilename = "TASKMD_SPEC.md"

var (
	specForce  bool
	specStdout bool
)

var specCmd = &cobra.Command{
	Use:        "spec",
	SuggestFor: []string{"specification", "format", "schema"},
	Short:      "Generate the taskmd specification file",
	Long: `Generate the taskmd specification document in your project directory.

The specification describes the task file format, including frontmatter fields,
valid values, file naming conventions, and directory structure.

Examples:
  taskmd spec                    # Writes TASKMD_SPEC.md to current directory
  taskmd spec --stdout           # Print spec to stdout
  taskmd spec --dir ./docs       # Write to docs/ directory
  taskmd spec --force            # Overwrite existing file`,
	Args: cobra.NoArgs,
	RunE: runSpec,
}

func init() {
	rootCmd.AddCommand(specCmd)

	specCmd.Flags().BoolVar(&specForce, "force", false, "overwrite existing TASKMD_SPEC.md")
	specCmd.Flags().BoolVar(&specStdout, "stdout", false, "print spec to stdout instead of writing a file")
}

func runSpec(_ *cobra.Command, _ []string) error {
	if specStdout {
		fmt.Print(string(specTemplate))
		return nil
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

	outputPath := filepath.Join(targetDir, specFilename)

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	// Check if file exists
	if !specForce {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("%s already exists at %s (use --force to overwrite)", specFilename, absPath)
		}
	}

	// Write the file
	if err := os.WriteFile(outputPath, specTemplate, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", specFilename, err)
	}

	if !GetGlobalFlags().Quiet {
		fmt.Printf("Created %s\n", absPath)
	}

	return nil
}
