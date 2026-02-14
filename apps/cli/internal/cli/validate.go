package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/validator"
)

var (
	validateFormat string
	validateStrict bool
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:        "validate",
	SuggestFor: []string{"check", "verify", "lint"},
	Short:      "Lint and validate tasks",
	Long: `Validate checks task files for common errors and issues.

Validation checks include:
  - Required fields (id, title)
  - Invalid field values (status, priority, effort)
  - Duplicate task IDs
  - Missing dependencies (references to non-existent tasks)
  - Circular dependencies (cycles in dependency graph)

Use --strict to enable additional warnings for missing optional fields.

Output formats: text (default), table, json

Exit codes:
  0 - Valid (no errors)
  1 - Invalid (errors found)
  2 - Valid but with warnings (only in strict mode)

Examples:
  taskmd validate
  taskmd validate ./tasks
  taskmd validate --strict
  taskmd validate --format json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVar(&validateFormat, "format", "text", "output format (text, table, json)")
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "enable strict validation with additional warnings")
}

func runValidate(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	// Create scanner and scan for tasks
	taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	tasks := result.Tasks

	// Report scan errors if any
	if len(result.Errors) > 0 {
		if !flags.Quiet {
			fmt.Fprintf(os.Stderr, "Warning: encountered %d errors during scan:\n", len(result.Errors))
			for _, scanErr := range result.Errors {
				fmt.Fprintf(os.Stderr, "  %s: %v\n", scanErr.FilePath, scanErr.Error)
			}
			fmt.Fprintln(os.Stderr)
		}
	}

	// Run validation
	v := validator.NewValidator(validateStrict)
	validationResult := v.Validate(tasks)

	// Output results
	switch validateFormat {
	case "json":
		if err := outputValidationJSON(validationResult); err != nil {
			return err
		}
	case "text", "table":
		outputValidationText(validationResult, flags.Quiet)
	default:
		return ValidateFormat(validateFormat, []string{"text", "table", "json"})
	}

	// Determine exit code
	if !validationResult.IsValid() {
		os.Exit(ExitError)
	} else if validateStrict && validationResult.HasWarnings() {
		os.Exit(ExitValidationWarning)
	}

	return nil
}

// outputValidationText outputs validation results in human-readable text format
func outputValidationText(result *validator.ValidationResult, quiet bool) {
	r := getRenderer()

	if len(result.Issues) == 0 {
		if !quiet {
			fmt.Printf("%s All %d task(s) are valid\n", formatSuccess("✓", r), result.TaskCount)
		}
		return
	}

	// Group issues by level
	errors := []validator.ValidationIssue{}
	warnings := []validator.ValidationIssue{}

	for _, issue := range result.Issues {
		if issue.Level == validator.LevelError {
			errors = append(errors, issue)
		} else if issue.Level == validator.LevelWarning {
			warnings = append(warnings, issue)
		}
	}

	// Print errors
	if len(errors) > 0 {
		fmt.Printf("\n%s Found %d error(s):\n\n", formatError("❌", r), len(errors))
		for _, issue := range errors {
			printIssue(issue, r)
		}
	}

	// Print warnings
	if len(warnings) > 0 {
		fmt.Printf("\n%s  Found %d warning(s):\n\n", formatWarning("⚠️", r), len(warnings))
		for _, issue := range warnings {
			printIssue(issue, r)
		}
	}

	// Print summary
	fmt.Println()
	if result.Errors > 0 {
		fmt.Printf("Validated %d task(s): %s", result.TaskCount,
			formatError(fmt.Sprintf("%d error(s)", result.Errors), r))
		if result.Warnings > 0 {
			fmt.Printf(", %s", formatWarning(fmt.Sprintf("%d warning(s)", result.Warnings), r))
		}
		fmt.Println()
	} else if result.Warnings > 0 {
		fmt.Printf("Validated %d task(s) with %s\n", result.TaskCount,
			formatWarning(fmt.Sprintf("%d warning(s)", result.Warnings), r))
	}
}

// printIssue prints a single validation issue
func printIssue(issue validator.ValidationIssue, r *lipgloss.Renderer) {
	if issue.TaskID != "" {
		fmt.Printf("  [%s] %s\n", formatTaskID(issue.TaskID, r), issue.Message)
	} else {
		fmt.Printf("  %s\n", issue.Message)
	}

	if issue.FilePath != "" {
		fmt.Printf("    %s %s\n", formatLabel("File:", r), formatDim(issue.FilePath, r))
	}
}

// outputValidationJSON outputs validation results as JSON
func outputValidationJSON(result *validator.ValidationResult) error {
	return WriteJSON(os.Stdout, result)
}
