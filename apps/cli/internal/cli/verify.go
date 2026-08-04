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

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/verify"
)

// ErrVerifyFailed is returned when one or more verification checks fail.
var ErrVerifyFailed = errors.New("verification failed")

var (
	verifyTaskID  string
	verifyFormat  string
	verifyDryRun  bool
	verifyTimeout int
	verifyAll     bool
)

var verifyCmd = &cobra.Command{
	Use:        "verify [task-id]",
	SuggestFor: []string{"test", "check"},
	Short:      "Run verification checks for a task",
	Long: `Verify runs the acceptance checks defined in a task's verify field.

The task is identified by a positional argument or --task-id (exact match only).

Each verify step has a type:
  bash   — runs a shell command, reports pass/fail based on exit code
  assert — displays a check for the agent to evaluate (not executed)

By default, verification stops at the first failure (fail-fast). Use --all
to run every check regardless of failures.

Exit codes:
  0 - All executable checks passed
  1 - One or more executable checks failed

Examples:
  taskmd verify 042
  taskmd verify 042 --all
  taskmd verify 042 --format json
  taskmd verify 042 --dry-run
  taskmd verify --task-id 042 --timeout 120`,
	Args: cobra.MaximumNArgs(1),
	RunE: runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().StringVar(&verifyTaskID, "task-id", "", "task ID to verify")
	verifyCmd.Flags().StringVar(&verifyFormat, "format", "table", "output format (table, json)")
	verifyCmd.Flags().BoolVar(&verifyDryRun, "dry-run", false, "list checks without executing")
	verifyCmd.Flags().IntVar(&verifyTimeout, "timeout", 60, "per-command timeout in seconds")
	verifyCmd.Flags().BoolVar(&verifyAll, "all", false, "run all checks even if one fails")
}

func resolveVerifyTaskID(args []string) (string, error) {
	positional := ""
	if len(args) > 0 {
		positional = args[0]
	}
	flagVal := verifyTaskID

	if positional != "" && flagVal != "" && positional != flagVal {
		return "", fmt.Errorf("conflicting task ID: positional argument %q and --task-id %q differ", positional, flagVal)
	}

	if positional != "" {
		return positional, nil
	}
	if flagVal != "" {
		return flagVal, nil
	}
	return "", fmt.Errorf("task ID required: provide as positional argument or --task-id flag")
}

func runVerify(cmd *cobra.Command, args []string) error {
	taskID, err := resolveVerifyTaskID(args)
	if err != nil {
		return err
	}

	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}

	task := findExactMatch(taskID, tasks)
	if task == nil {
		return fmt.Errorf("task not found: %s", taskID)
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
		FailFast:    !verifyAll,
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
		if step.Status == verify.StatusFail {
			printFailureOutput(step, r)
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

func printFailureOutput(step verify.StepResult, r *lipgloss.Renderer) {
	printDimLines := func(s string) {
		for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
			fmt.Printf("       %s\n", formatDim(line, r))
		}
	}

	if step.Stdout != "" {
		printDimLines(step.Stdout)
	}
	if step.Stderr != "" {
		printDimLines(step.Stderr)
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

// resolveProjectRoot returns the absolute path of the directory containing .taskmd.yaml, or cwd as fallback.
// Since initConfig() already walks up the directory tree to find the config, we just use the config location.
func resolveProjectRoot() string {
	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		abs, err := filepath.Abs(filepath.Dir(configFile))
		if err == nil {
			return abs
		}
		return filepath.Dir(configFile)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
