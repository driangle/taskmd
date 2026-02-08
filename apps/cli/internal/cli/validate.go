package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
	"github.com/driangle/md-task-tracker/apps/cli/internal/validator"
	"github.com/spf13/cobra"
)

var (
	validateStrict bool
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [directory]",
	Short: "Lint and validate tasks",
	Long: `Validate checks task files for common errors and issues.

Validation checks include:
  - Required fields (id, title)
  - Invalid field values (status, priority, effort)
  - Duplicate task IDs
  - Missing dependencies (references to non-existent tasks)
  - Circular dependencies (cycles in dependency graph)

Use --strict to enable additional warnings for missing optional fields.

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

	validateCmd.Flags().BoolVar(&validateStrict, "strict", false, "enable strict validation with additional warnings")
}

func runValidate(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	// Determine scan directory
	scanDir := "."
	if len(args) > 0 {
		scanDir = args[0]
	}

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
	switch flags.Format {
	case "json":
		if err := outputValidationJSON(validationResult); err != nil {
			return err
		}
	case "text", "table":
		outputValidationText(validationResult, flags.Quiet)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json)", flags.Format)
	}

	// Determine exit code
	if !validationResult.IsValid() {
		os.Exit(1)
	} else if validateStrict && validationResult.HasWarnings() {
		os.Exit(2)
	}

	return nil
}

// outputValidationText outputs validation results in human-readable text format
func outputValidationText(result *validator.ValidationResult, quiet bool) {
	if len(result.Issues) == 0 {
		if !quiet {
			fmt.Printf("✓ All %d task(s) are valid\n", result.TaskCount)
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
		fmt.Printf("\n❌ Found %d error(s):\n\n", len(errors))
		for _, issue := range errors {
			printIssue(issue)
		}
	}

	// Print warnings
	if len(warnings) > 0 {
		fmt.Printf("\n⚠️  Found %d warning(s):\n\n", len(warnings))
		for _, issue := range warnings {
			printIssue(issue)
		}
	}

	// Print summary
	fmt.Println()
	if result.Errors > 0 {
		fmt.Printf("Validated %d task(s): %d error(s)", result.TaskCount, result.Errors)
		if result.Warnings > 0 {
			fmt.Printf(", %d warning(s)", result.Warnings)
		}
		fmt.Println()
	} else if result.Warnings > 0 {
		fmt.Printf("Validated %d task(s) with %d warning(s)\n", result.TaskCount, result.Warnings)
	}
}

// printIssue prints a single validation issue
func printIssue(issue validator.ValidationIssue) {
	if issue.TaskID != "" {
		fmt.Printf("  [%s] %s\n", issue.TaskID, issue.Message)
	} else {
		fmt.Printf("  %s\n", issue.Message)
	}

	if issue.FilePath != "" {
		fmt.Printf("    File: %s\n", issue.FilePath)
	}
}

// outputValidationJSON outputs validation results as JSON
func outputValidationJSON(result *validator.ValidationResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
