package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/taskcontext"
)

var (
	ctxTaskID         string
	ctxFormat         string
	ctxResolve        bool
	ctxIncludeContent bool
	ctxIncludeDeps    bool
	ctxMaxFiles       int
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Show file context for a task",
	Long: `Context resolves all relevant files for a task into a structured output.

Files come from two sources:
  1. Scope files — resolved from the task's "touches" field via scope definitions in .taskmd.yaml
  2. Explicit files — listed directly in the task's "context" field

Each file entry includes:
  - path: relative file path
  - source: where it came from (scope:<name> or explicit)
  - exists: whether the file exists on disk

Flags:
  --resolve          expand directory paths to individual files
  --include-content  inline file contents and task body
  --include-deps     include files from direct dependency tasks
  --max-files        cap number of files returned

Examples:
  taskmd context --task-id 042
  taskmd context --task-id 042 --format json
  taskmd context --task-id 042 --include-content --resolve
  taskmd context --task-id 042 --include-deps --max-files 20`,
	Args: cobra.NoArgs,
	RunE: runContext,
}

func init() {
	rootCmd.AddCommand(contextCmd)

	contextCmd.Flags().StringVar(&ctxTaskID, "task-id", "", "task ID to build context for (required)")
	contextCmd.Flags().StringVar(&ctxFormat, "format", "text", "output format (text, json, yaml)")
	contextCmd.Flags().BoolVar(&ctxResolve, "resolve", false, "expand directory paths to individual files")
	contextCmd.Flags().BoolVar(&ctxIncludeContent, "include-content", false, "inline file contents and task body")
	contextCmd.Flags().BoolVar(&ctxIncludeDeps, "include-deps", false, "include files from direct dependency tasks")
	contextCmd.Flags().IntVar(&ctxMaxFiles, "max-files", 0, "cap number of files (0 = unlimited)")

	_ = contextCmd.MarkFlagRequired("task-id")
}

func runContext(cmd *cobra.Command, _ []string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	task := findExactMatch(ctxTaskID, result.Tasks)
	if task == nil {
		return fmt.Errorf("task not found: %s", ctxTaskID)
	}

	scopes := loadScopePathsConfig()
	projectRoot := resolveProjectRoot()

	opts := taskcontext.Options{
		Scopes:         scopes,
		ProjectRoot:    projectRoot,
		Resolve:        ctxResolve,
		IncludeContent: ctxIncludeContent,
		MaxFiles:       ctxMaxFiles,
	}

	ctxResult, err := taskcontext.Resolve(task, opts)
	if err != nil {
		return fmt.Errorf("context resolution failed: %w", err)
	}

	if ctxIncludeDeps {
		mergeDependencyFiles(task, result.Tasks, opts, ctxResult)
	}

	addDependencyEntries(task, result.Tasks, ctxResult)

	return outputContext(ctxResult, ctxFormat)
}

// mergeDependencyFiles resolves context for each direct dependency and merges files.
func mergeDependencyFiles(task *model.Task, allTasks []*model.Task, opts taskcontext.Options, target *taskcontext.Result) {
	taskMap := buildTaskMap(allTasks)

	for _, depID := range task.Dependencies {
		dep, ok := taskMap[depID]
		if !ok {
			continue
		}
		depResult, err := taskcontext.Resolve(dep, opts)
		if err != nil {
			continue
		}
		target.Files = mergeFiles(target.Files, depResult.Files)
	}
}

// mergeFiles appends src files into dst, deduplicating by path.
func mergeFiles(dst, src []taskcontext.FileEntry) []taskcontext.FileEntry {
	seen := make(map[string]bool, len(dst))
	for _, f := range dst {
		seen[f.Path] = true
	}
	for _, f := range src {
		if seen[f.Path] {
			continue
		}
		seen[f.Path] = true
		dst = append(dst, f)
	}
	return dst
}

// addDependencyEntries populates the Dependencies field on the result.
func addDependencyEntries(task *model.Task, allTasks []*model.Task, result *taskcontext.Result) {
	if len(task.Dependencies) == 0 {
		return
	}
	taskMap := buildTaskMap(allTasks)
	for _, depID := range task.Dependencies {
		entry := taskcontext.DepEntry{ID: depID}
		if dep, ok := taskMap[depID]; ok {
			entry.Title = dep.Title
			entry.Status = string(dep.Status)
		}
		result.Dependencies = append(result.Dependencies, entry)
	}
}

func outputContext(result *taskcontext.Result, format string) error {
	switch format {
	case "text":
		return outputContextText(result, os.Stdout)
	case "json":
		return WriteJSON(os.Stdout, result)
	case "yaml":
		return WriteYAML(os.Stdout, result)
	default:
		return ValidateFormat(format, []string{"text", "json", "yaml"})
	}
}

func outputContextText(result *taskcontext.Result, w *os.File) error {
	r := getRenderer()

	fmt.Fprintf(w, "Context for task %s (%s)\n", formatTaskID(result.TaskID, r), result.Title)

	if result.TaskBody != "" {
		separator := strings.Repeat("\u2500", 49)
		fmt.Fprintf(w, "\n%s\n%s\n%s\n%s\n", formatLabel("Task:", r), separator, result.TaskBody, separator)
	}

	printContextFilesBySource(w, result.Files, r)
	printContextDeps(w, result.Dependencies, r)

	return nil
}

// printContextFilesBySource groups files by source type and prints them.
func printContextFilesBySource(w *os.File, files []taskcontext.FileEntry, r *lipgloss.Renderer) {
	// Group scope files by scope name
	scopeGroups := make(map[string][]taskcontext.FileEntry)
	var scopeOrder []string
	var explicit []taskcontext.FileEntry

	for _, f := range files {
		if strings.HasPrefix(f.Source, "scope:") {
			name := strings.TrimPrefix(f.Source, "scope:")
			if _, exists := scopeGroups[name]; !exists {
				scopeOrder = append(scopeOrder, name)
			}
			scopeGroups[name] = append(scopeGroups[name], f)
		} else {
			explicit = append(explicit, f)
		}
	}

	for _, name := range scopeOrder {
		fmt.Fprintf(w, "\n%s\n", formatLabel(fmt.Sprintf("Scope files (%s):", name), r))
		for _, f := range scopeGroups[name] {
			printContextFile(w, f, r)
		}
	}

	if len(explicit) > 0 {
		fmt.Fprintf(w, "\n%s\n", formatLabel("Explicit files:", r))
		for _, f := range explicit {
			printContextFile(w, f, r)
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(w, "\n%s\n", formatDim("No context files found.", r))
	}
}

func printContextFile(w *os.File, f taskcontext.FileEntry, r *lipgloss.Renderer) {
	path := f.Path
	if !f.Exists {
		path += " " + formatWarning("(missing)", r)
	}
	if f.Content != "" {
		fmt.Fprintf(w, "  %s (%d lines)\n", path, f.Lines)
		separator := strings.Repeat("\u2500", 49)
		fmt.Fprintf(w, "  %s\n%s\n  %s\n", separator, f.Content, separator)
	} else {
		fmt.Fprintf(w, "  %s\n", path)
	}
}

func printContextDeps(w *os.File, deps []taskcontext.DepEntry, r *lipgloss.Renderer) {
	if len(deps) == 0 {
		return
	}
	fmt.Fprintf(w, "\n%s\n", formatLabel("Dependencies:", r))
	for _, d := range deps {
		status := ""
		if d.Status != "" {
			status = fmt.Sprintf("(%s)", formatStatus(d.Status, r))
		}
		title := d.Title
		if title == "" {
			title = d.ID
		}
		fmt.Fprintf(w, "  %s %s %s\n", formatTaskID(d.ID, r), title, status)
	}
}
