package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/verify"
)

// ErrVerifyFailed is returned when one or more verification checks fail.
var ErrVerifyFailed = errors.New("verification failed")

var (
	verifyTaskID  string
	verifyFormat  string
	verifyDryRun  bool
	verifyTimeout int
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Run verification checks for a task",
	Long: `Verify runs the acceptance checks defined in a task's verify field.

Each verify step has a type:
  bash   — runs a shell command, reports pass/fail based on exit code
  assert — displays a check for the agent to evaluate (not executed)

Exit codes:
  0 - All executable checks passed
  1 - One or more executable checks failed

Examples:
  taskmd verify --task-id 042
  taskmd verify --task-id 042 --format json
  taskmd verify --task-id 042 --dry-run
  taskmd verify --task-id 042 --timeout 120`,
	Args: cobra.NoArgs,
	RunE: runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().StringVar(&verifyTaskID, "task-id", "", "task ID to verify (required)")
	verifyCmd.Flags().StringVar(&verifyFormat, "format", "table", "output format (table, json)")
	verifyCmd.Flags().BoolVar(&verifyDryRun, "dry-run", false, "list checks without executing")
	verifyCmd.Flags().IntVar(&verifyTimeout, "timeout", 60, "per-command timeout in seconds")

	_ = verifyCmd.MarkFlagRequired("task-id")
}

func runVerify(cmd *cobra.Command, _ []string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	task := findExactMatch(verifyTaskID, result.Tasks)
	if task == nil {
		return fmt.Errorf("task not found: %s", verifyTaskID)
	}

	if len(task.Verify) == 0 {
		fmt.Println("No verification checks defined for this task.")
		return nil
	}

	if errs := model.ValidateVerifySteps(task.Verify); len(errs) > 0 {
		return fmt.Errorf("invalid verify steps:\n  %s", strings.Join(errs, "\n  "))
	}

	projectRoot := resolveProjectRoot()
	opts := verify.Options{
		ProjectRoot: projectRoot,
		DryRun:      verifyDryRun,
		Timeout:     time.Duration(verifyTimeout) * time.Second,
		Verbose:     flags.Verbose,
		LogFunc: func(format string, args ...any) {
			if !flags.Quiet {
				fmt.Fprintf(os.Stderr, format+"\n", args...)
			}
		},
	}

	vResult := verify.Run(task.Verify, opts)

	switch verifyFormat {
	case "json":
		return WriteJSON(os.Stdout, vResult)
	case "table":
		printVerifyTable(vResult)
	default:
		return ValidateFormat(verifyFormat, []string{"table", "json"})
	}

	if vResult.HasFailures() {
		return ErrVerifyFailed
	}
	return nil
}

func printVerifyTable(result *verify.Result) {
	r := getRenderer()

	for _, step := range result.Steps {
		printVerifyStep(step, r)
	}

	fmt.Println()
	fmt.Printf("  %s\n", verifySummaryLine(result, r))
}

func printVerifyStep(step verify.StepResult, r *lipgloss.Renderer) {
	statusStr := formatVerifyStatus(step.Status, r)
	switch step.Type {
	case "bash":
		fmt.Printf("  %s  %s\n", statusStr, step.Command)
		if step.Dir != "" {
			fmt.Printf("       dir: %s\n", formatDim(step.Dir, r))
		}
		if step.Status == verify.StatusFail && step.Stderr != "" {
			for _, line := range strings.Split(strings.TrimRight(step.Stderr, "\n"), "\n") {
				fmt.Printf("       %s\n", formatDim(line, r))
			}
		}
	case "assert":
		fmt.Printf("  %s  %s\n", statusStr, step.Check)
	default:
		label := step.Type
		if step.Warning != "" {
			label = step.Warning
		}
		fmt.Printf("  %s  %s\n", statusStr, label)
	}
}

func verifySummaryLine(result *verify.Result, r *lipgloss.Renderer) string {
	var parts []string
	if result.Passed > 0 {
		parts = append(parts, formatSuccess(fmt.Sprintf("%d passed", result.Passed), r))
	}
	if result.Failed > 0 {
		parts = append(parts, formatError(fmt.Sprintf("%d failed", result.Failed), r))
	}
	if result.Pending > 0 {
		parts = append(parts, formatWarning(fmt.Sprintf("%d pending", result.Pending), r))
	}
	if result.Skipped > 0 {
		parts = append(parts, formatDim(fmt.Sprintf("%d skipped", result.Skipped), r))
	}
	return strings.Join(parts, ", ")
}

func formatVerifyStatus(status verify.StepStatus, r *lipgloss.Renderer) string {
	switch status {
	case verify.StatusPass:
		return formatSuccess("PASS", r)
	case verify.StatusFail:
		return formatError("FAIL", r)
	case verify.StatusPending:
		return formatWarning("PEND", r)
	case verify.StatusSkip:
		return formatDim("SKIP", r)
	default:
		return string(status)
	}
}

// resolveProjectRoot returns the directory containing .taskmd.yaml, or cwd as fallback.
func resolveProjectRoot() string {
	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		return filepath.Dir(configFile)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
