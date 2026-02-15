package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var (
	statusFormat    string
	statusExact     bool
	statusThreshold float64
)

// statusStdinReader is the reader used for interactive selection prompts.
// Override in tests to simulate user input.
var statusStdinReader io.Reader = os.Stdin

var statusCmd = &cobra.Command{
	Use:   "status <query>",
	Short: "Get lightweight metadata for a task (no body, no resolved deps)",
	Long: `Status displays only the frontmatter metadata of a task, without body content,
resolved dependency info, context files, or worklog data. Use this when you just
need to quickly check a task's status, priority, or other metadata.

Matching uses the same logic as 'get' (ID, title, file path, fuzzy).

Examples:
  taskmd status 042
  taskmd status "Setup project"
  taskmd status 042 --format json
  taskmd status 042 --format yaml
  taskmd status sho --exact`,
	Args: cobra.ExactArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVar(&statusFormat, "format", "text", "output format (text, json, yaml)")
	statusCmd.Flags().BoolVar(&statusExact, "exact", false, "disable fuzzy matching, exact only")
	statusCmd.Flags().Float64Var(&statusThreshold, "threshold", 0.6, "fuzzy match sensitivity (0.0-1.0)")
}

// statusOutput is the lightweight metadata struct for JSON/YAML output.
type statusOutput struct {
	ID           string   `json:"id" yaml:"id"`
	Title        string   `json:"title" yaml:"title"`
	Status       string   `json:"status" yaml:"status"`
	Priority     string   `json:"priority,omitempty" yaml:"priority,omitempty"`
	Effort       string   `json:"effort,omitempty" yaml:"effort,omitempty"`
	Tags         []string `json:"tags" yaml:"tags"`
	Owner        string   `json:"owner,omitempty" yaml:"owner,omitempty"`
	Parent       string   `json:"parent,omitempty" yaml:"parent,omitempty"`
	Created      string   `json:"created,omitempty" yaml:"created,omitempty"`
	Dependencies []string `json:"dependencies" yaml:"dependencies"`
	Group        string   `json:"group,omitempty" yaml:"group,omitempty"`
	FilePath     string   `json:"file_path" yaml:"file_path"`
}

func runStatus(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	query := args[0]

	scanDir := ResolveScanDir(nil)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	tasks := result.Tasks
	makeFilePathsRelative(tasks, scanDir)

	// Swap stdin reader for fuzzy selection prompts
	origReader := getStdinReader
	getStdinReader = statusStdinReader
	defer func() { getStdinReader = origReader }()

	task, err := resolveTask(query, tasks, statusExact, statusThreshold)
	if err != nil {
		return err
	}

	out := buildStatusOutputFromTask(task)

	switch statusFormat {
	case "text":
		return outputStatusText(out, os.Stdout)
	case "json":
		return WriteJSON(os.Stdout, out)
	case "yaml":
		return WriteYAML(os.Stdout, out)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", statusFormat)
	}
}

func buildStatusOutputFromTask(task *model.Task) statusOutput {
	created := ""
	if !task.Created.IsZero() {
		created = task.Created.Format("2006-01-02")
	}
	return statusOutput{
		ID:           task.ID,
		Title:        task.Title,
		Status:       string(task.Status),
		Priority:     string(task.Priority),
		Effort:       string(task.Effort),
		Tags:         task.Tags,
		Owner:        task.Owner,
		Parent:       task.Parent,
		Created:      created,
		Dependencies: task.Dependencies,
		Group:        task.Group,
		FilePath:     task.FilePath,
	}
}

func outputStatusText(out statusOutput, w io.Writer) error {
	r := getRenderer()

	fmt.Fprintf(w, "%s %s\n", formatLabel("Task:", r), formatTaskID(out.ID, r))
	fmt.Fprintf(w, "%s %s\n", formatLabel("Title:", r), out.Title)
	fmt.Fprintf(w, "%s %s\n", formatLabel("Status:", r), formatStatus(out.Status, r))
	printStatusOptionalField(w, "Priority", out.Priority, r)
	printStatusOptionalField(w, "Effort", out.Effort, r)
	if len(out.Tags) > 0 {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Tags:", r), strings.Join(out.Tags, ", "))
	}
	if out.Owner != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Owner:", r), out.Owner)
	}
	if out.Parent != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Parent:", r), out.Parent)
	}
	if out.Created != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Created:", r), out.Created)
	}
	if len(out.Dependencies) > 0 {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Dependencies:", r), strings.Join(out.Dependencies, ", "))
	}
	if out.Group != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Group:", r), out.Group)
	}
	fmt.Fprintf(w, "%s %s\n", formatLabel("File:", r), formatDim(out.FilePath, r))
	return nil
}

func printStatusOptionalField(w io.Writer, label, value string, r *lipgloss.Renderer) {
	if value == "" {
		return
	}
	var colored string
	switch label {
	case "Priority":
		colored = formatPriority(value, r)
	case "Effort":
		colored = formatEffort(value, r)
	default:
		colored = value
	}
	fmt.Fprintf(w, "%s %s\n", formatLabel(label+":", r), colored)
}
