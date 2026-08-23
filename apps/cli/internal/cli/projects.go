package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/driangle/taskmd/sdk/go/model"
)

var projectsFormat string

var projectsCmd = &cobra.Command{
	Use:        "projects",
	SuggestFor: []string{"proj", "project"},
	Short:      "List and manage registered projects",
	Long: `Projects lists all globally registered projects with summary task stats.

When run without a subcommand, lists all registered projects.
Use subcommands to register and unregister projects.

Projects are registered in ~/.taskmd.yaml under the "projects" key.

Examples:
  taskmd projects
  taskmd projects --format json
  taskmd projects register
  taskmd projects unregister --id my-project`,
	RunE: runProjects,
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	projectsCmd.Flags().StringVar(&projectsFormat, "format", "table", "output format (table, json, yaml)")
}

// ProjectSummary holds computed stats for a registered project.
type ProjectSummary struct {
	ID         string `json:"id" yaml:"id"`
	Name       string `json:"name" yaml:"name"`
	Path       string `json:"path" yaml:"path"`
	Tasks      int    `json:"tasks" yaml:"tasks"`
	Pending    int    `json:"pending" yaml:"pending"`
	InProgress int    `json:"in_progress" yaml:"in_progress"`
	Completed  int    `json:"completed" yaml:"completed"`
}

func runProjects(_ *cobra.Command, _ []string) error {
	projects, err := LoadGlobalRegistry()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	if len(projects) == 0 {
		fmt.Fprintln(os.Stderr, "No projects registered. Run 'taskmd projects register' in a project, or add projects to ~/.taskmd.yaml:")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  projects:")
		fmt.Fprintln(os.Stderr, "    - id: my-project")
		fmt.Fprintln(os.Stderr, "      name: \"My Project\"")
		fmt.Fprintln(os.Stderr, "      path: /path/to/project")
		return nil
	}

	summaries := collectProjectSummaries(projects)

	switch projectsFormat {
	case "json":
		return WriteJSON(os.Stdout, summaries)
	case "yaml":
		return WriteYAML(os.Stdout, summaries)
	case "table":
		return outputProjectsTable(summaries)
	default:
		return ValidateFormat(projectsFormat, []string{"table", "json", "yaml"})
	}
}

// projectDirConfig is a minimal struct to read the task directory from .taskmd.yaml.
type projectDirConfig struct {
	Dir     string `yaml:"dir"`
	TaskDir string `yaml:"task-dir"`
}

// resolveProjectScanDir reads a project's .taskmd.yaml to find the task directory for scanning.
// Unlike resolveProjectTaskDir in root.go, this doesn't use viper (avoids polluting global state).
func resolveProjectScanDir(projectPath string) string {
	configPath := filepath.Join(projectPath, ".taskmd.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return filepath.Join(projectPath, "tasks")
	}

	var cfg projectDirConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return filepath.Join(projectPath, "tasks")
	}

	dir := cfg.TaskDir
	if dir == "" {
		dir = cfg.Dir
	}
	if dir == "" {
		dir = "tasks"
	}

	if filepath.IsAbs(dir) {
		return dir
	}
	return filepath.Join(projectPath, dir)
}

func collectProjectSummaries(projects []GlobalProjectEntry) []ProjectSummary {
	summaries := make([]ProjectSummary, 0, len(projects))
	for _, p := range projects {
		summary, err := scanProjectSummary(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping project %q: %v\n", p.Name, err)
			continue
		}
		summaries = append(summaries, summary)
	}
	return summaries
}

func scanProjectSummary(p GlobalProjectEntry) (ProjectSummary, error) {
	info, err := os.Stat(p.Path)
	if err != nil || !info.IsDir() {
		return ProjectSummary{}, fmt.Errorf("path %q is not accessible", p.Path)
	}

	scanDir := resolveProjectScanDir(p.Path)
	taskScanner := newTaskScanner(scanDir, GlobalFlags{})
	result, err := taskScanner.Scan()
	if err != nil {
		return ProjectSummary{}, fmt.Errorf("scan failed: %w", err)
	}

	return buildProjectSummary(p, result.Tasks), nil
}

func buildProjectSummary(p GlobalProjectEntry, tasks []*model.Task) ProjectSummary {
	summary := ProjectSummary{
		ID:   p.ID,
		Name: p.Name,
		Path: p.Path,
	}
	for _, task := range tasks {
		summary.Tasks++
		switch task.Status {
		case model.StatusPending:
			summary.Pending++
		case model.StatusInProgress:
			summary.InProgress++
		case model.StatusCompleted:
			summary.Completed++
		}
	}
	return summary
}

func outputProjectsTable(summaries []ProjectSummary) error {
	tw := NewTableWriter()
	tw.AddHeader([]string{"PROJECT", "PATH", "TASKS", "PENDING", "IN-PROGRESS", "COMPLETED"})
	tw.AddSeparator()
	for _, s := range summaries {
		row := []string{
			s.Name,
			s.Path,
			fmt.Sprintf("%d", s.Tasks),
			fmt.Sprintf("%d", s.Pending),
			fmt.Sprintf("%d", s.InProgress),
			fmt.Sprintf("%d", s.Completed),
		}
		tw.AddRow(row, row)
	}
	tw.Flush(os.Stdout)
	return nil
}
