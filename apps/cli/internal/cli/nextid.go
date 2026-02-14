package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/nextid"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var nextIDFormat string

var nextIDCmd = &cobra.Command{
	Use:   "next-id [directory]",
	Short: "Show the next available task ID",
	Long: `Next-id scans task files and outputs the next available sequential ID.

It finds the highest numeric ID among existing tasks and returns max + 1,
preserving any common prefix and zero-padding.

By default outputs just the ID string (ideal for scripting). Use --format json
for structured output with additional metadata.

Examples:
  taskmd next-id
  taskmd next-id ./tasks/cli
  taskmd next-id --format json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNextID,
}

func init() {
	rootCmd.AddCommand(nextIDCmd)

	nextIDCmd.Flags().StringVar(&nextIDFormat, "format", "plain", "output format (plain, json)")
}

func runNextID(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if flags.Verbose && len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarning: encountered %d errors during scan:\n", len(result.Errors))
		for _, scanErr := range result.Errors {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", scanErr.FilePath, scanErr.Error)
		}
		fmt.Fprintln(os.Stderr)
	}

	ids := make([]string, len(result.Tasks))
	for i, task := range result.Tasks {
		ids[i] = task.ID
	}

	nextResult := nextid.Calculate(ids)

	switch nextIDFormat {
	case "plain":
		fmt.Println(nextResult.NextID)
		return nil
	case "json":
		return WriteJSON(os.Stdout, nextResult)
	default:
		return ValidateFormat(nextIDFormat, []string{"plain", "json"})
	}
}
