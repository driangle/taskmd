package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/template"
	"github.com/driangle/taskmd/sdk/go/nextid"
	"github.com/driangle/taskmd/sdk/go/slug"
	"github.com/driangle/taskmd/sdk/go/taskfile"
	"github.com/driangle/taskmd/sdk/go/validator"
)

var (
	addPriority  string
	addEffort    string
	addTags      string
	addStatus    string
	addOwner     string
	addDependsOn string
	addParent    string
	addGroup     string
	addFormat    string
	addEdit      bool
	addTemplate  string
	addSlug      string
	addPhase     string
)

var addCmd = &cobra.Command{
	Use:        "add <title>",
	SuggestFor: []string{"create", "new"},
	Short:      "Create a new task",
	Long: `Add creates a new task markdown file with proper frontmatter.

The title is used to generate both the task title and the filename slug.
Use --slug to provide a custom slug instead of the auto-generated one.
A sequential ID is automatically assigned based on existing tasks.

Use --template to start from a reusable template (bug, feature, chore, or custom).
CLI flags override template values when explicitly provided.

Examples:
  taskmd add "Fix the login bug"
  taskmd add "Implement OAuth" --priority high --tags backend,auth
  taskmd add "Design mockups" --group design --effort large
  taskmd add "Quick fix" --edit
  taskmd add "Login fails on Safari" --template bug
  taskmd add "Dark mode support" --template feature --priority high
  taskmd add "Fix the login bug" --slug fix-login`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVar(&addPriority, "priority", "medium", "task priority (low, medium, high, critical)")
	addCmd.Flags().StringVar(&addEffort, "effort", "", "task effort (small, medium, large)")
	addCmd.Flags().StringVar(&addTags, "tags", "", "comma-separated tags")
	addCmd.Flags().StringVar(&addStatus, "status", "pending", "task status (pending, in-progress, completed, blocked, cancelled)")
	addCmd.Flags().StringVar(&addOwner, "owner", "", "task owner/assignee")
	addCmd.Flags().StringVar(&addDependsOn, "depends-on", "", "comma-separated dependency task IDs")
	addCmd.Flags().StringVar(&addParent, "parent", "", "parent task ID")
	addCmd.Flags().StringVar(&addGroup, "group", "", "subdirectory to create the task in")
	addCmd.Flags().StringVar(&addFormat, "format", "plain", "output format (plain, json)")
	addCmd.Flags().BoolVar(&addEdit, "edit", false, "open the new task in $EDITOR")
	addCmd.Flags().StringVar(&addTemplate, "template", "", "use a task template (e.g. bug, feature, chore)")
	addCmd.Flags().StringVar(&addSlug, "slug", "", "custom filename slug (default: auto-generated from title)")
	addCmd.Flags().StringVar(&addPhase, "phase", "", "phase name")

	_ = addCmd.RegisterFlagCompletionFunc("priority", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return validPriorityValues, cobra.ShellCompDirectiveNoFileComp
	})
	_ = addCmd.RegisterFlagCompletionFunc("effort", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return validEffortValues, cobra.ShellCompDirectiveNoFileComp
	})
	_ = addCmd.RegisterFlagCompletionFunc("status", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return validStatusValues, cobra.ShellCompDirectiveNoFileComp
	})
}

type addResult struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	FilePath string `json:"file_path"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}

func runAdd(cmd *cobra.Command, args []string) error {
	title := args[0]

	if err := validateAddEnums(); err != nil {
		return err
	}
	if err := ValidateFormat(addFormat, []string{"plain", "json"}); err != nil {
		return err
	}

	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	id, err := resolveNextID(scanDir, flags)
	if err != nil {
		return err
	}

	outputDir := scanDir
	if addGroup != "" {
		outputDir = filepath.Join(scanDir, addGroup)
	}

	suffix := addSlug
	if suffix == "" {
		suffix = slug.Slugify(title)
	}
	filePath := filepath.Join(outputDir, fmt.Sprintf("%s-%s.md", id, suffix))

	content, err := resolveTaskContent(cmd, id, title)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	if err := outputAddResult(id, title, filePath); err != nil {
		return err
	}

	if addEdit {
		return openInEditor(filePath)
	}

	return nil
}

func resolveNextID(scanDir string, flags GlobalFlags) (string, error) {
	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return "", err
	}
	ids := make([]string, len(tasks))
	for i, task := range tasks {
		ids[i] = task.ID
	}

	cfg := resolveIDConfig()
	return generateID(ids, cfg)
}

func generateID(ids []string, cfg validator.IDConfig) (string, error) {
	switch cfg.Strategy {
	case "prefixed":
		return nextid.GeneratePrefixed(ids, cfg.Prefix, cfg.Padding), nil
	case "random":
		return nextid.GenerateRandom(ids, cfg.Length)
	case "ulid":
		return nextid.GenerateULID(ids, cfg.Length)
	default:
		return nextid.Calculate(ids).NextID, nil
	}
}

func resolveTaskContent(cmd *cobra.Command, id, title string) (string, error) {
	if addTemplate != "" {
		return buildFromTemplate(cmd, id, title)
	}
	return buildTaskFileContent(id, title), nil
}

// buildFromTemplate resolves a template, renders it with variables, and applies CLI flag overrides.
func buildFromTemplate(cmd *cobra.Command, id, title string) (string, error) {
	projectRoot := resolveProjectRoot()
	userHome, _ := os.UserHomeDir()

	tmpl, ok := template.Resolve(addTemplate, projectRoot, userHome)
	if !ok {
		available := template.Discover(projectRoot, userHome)
		names := make([]string, len(available))
		for i, t := range available {
			names[i] = t.Name
		}
		return "", fmt.Errorf("template %q not found (available: %s)", addTemplate, strings.Join(names, ", "))
	}

	vars := map[string]string{
		"id":    id,
		"title": title,
		"date":  time.Now().Format("2006-01-02"),
	}

	content := template.RenderTask(tmpl, vars)

	// Apply CLI flag overrides (only for explicitly-set flags)
	overrides := buildTemplateOverrides(cmd)
	content = template.ApplyOverrides(content, overrides)

	return content, nil
}

// buildTemplateOverrides collects frontmatter overrides from explicitly-set CLI flags.
func buildTemplateOverrides(cmd *cobra.Command) map[string]string {
	overrides := make(map[string]string)

	if cmd.Flags().Changed("status") {
		overrides["status"] = addStatus
	}
	if cmd.Flags().Changed("priority") {
		overrides["priority"] = addPriority
	}
	if cmd.Flags().Changed("effort") {
		overrides["effort"] = addEffort
	}
	if cmd.Flags().Changed("owner") {
		overrides["owner"] = addOwner
	}
	if cmd.Flags().Changed("parent") {
		overrides["parent"] = fmt.Sprintf("%q", addParent)
	}
	if cmd.Flags().Changed("tags") {
		tags := parseTags()
		overrides["tags"] = taskfile.FormatInlineTags(tags)[len("tags: "):]
	}
	if cmd.Flags().Changed("phase") {
		overrides["phase"] = addPhase
	}
	if cmd.Flags().Changed("depends-on") {
		deps := parseDependsOn()
		overrides["dependencies"] = formatDependencies(deps)[len("dependencies: "):]
	}

	return overrides
}

func validateAddEnums() error {
	if !contains(validPriorityValues, addPriority) {
		return invalidValueError("priority", addPriority, validPriorityValues)
	}
	if !contains(validStatusValues, addStatus) {
		return invalidValueError("status", addStatus, validStatusValues)
	}
	if addEffort != "" && !contains(validEffortValues, addEffort) {
		return invalidValueError("effort", addEffort, validEffortValues)
	}
	return nil
}

func buildTaskFileContent(id, title string) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %q\n", id)
	fmt.Fprintf(&b, "title: %q\n", title)
	fmt.Fprintf(&b, "status: %s\n", addStatus)
	fmt.Fprintf(&b, "priority: %s\n", addPriority)

	if addEffort != "" {
		fmt.Fprintf(&b, "effort: %s\n", addEffort)
	}
	if addOwner != "" {
		fmt.Fprintf(&b, "owner: %s\n", addOwner)
	}
	if addParent != "" {
		fmt.Fprintf(&b, "parent: %q\n", addParent)
	}
	if addPhase != "" {
		fmt.Fprintf(&b, "phase: %s\n", addPhase)
	}

	deps := parseDependsOn()
	b.WriteString(formatDependencies(deps) + "\n")

	tags := parseTags()
	b.WriteString(taskfile.FormatInlineTags(tags) + "\n")

	fmt.Fprintf(&b, "created_at: %s\n", time.Now().Format("2006-01-02"))
	b.WriteString("---\n")

	b.WriteString("\n")
	fmt.Fprintf(&b, "# %s\n", title)
	b.WriteString("\n")
	b.WriteString("## Objective\n")
	b.WriteString("\n")
	b.WriteString("<!-- Describe the goal of this task -->\n")
	b.WriteString("\n")
	b.WriteString("## Tasks\n")
	b.WriteString("\n")
	b.WriteString("- [ ] TODO\n")
	b.WriteString("\n")
	b.WriteString("## Acceptance Criteria\n")
	b.WriteString("\n")
	b.WriteString("- TODO\n")

	return b.String()
}

func parseTags() []string {
	if addTags == "" {
		return nil
	}
	parts := strings.Split(addTags, ",")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

func parseDependsOn() []string {
	if addDependsOn == "" {
		return nil
	}
	parts := strings.Split(addDependsOn, ",")
	var deps []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			deps = append(deps, p)
		}
	}
	return deps
}

func formatDependencies(deps []string) string {
	if len(deps) == 0 {
		return "dependencies: []"
	}
	quoted := make([]string, len(deps))
	for i, d := range deps {
		quoted[i] = `"` + d + `"`
	}
	return "dependencies: [" + strings.Join(quoted, ", ") + "]"
}

func outputAddResult(id, title, filePath string) error {
	switch addFormat {
	case "json":
		return WriteJSON(os.Stdout, addResult{
			ID:       id,
			Title:    title,
			FilePath: filePath,
			Status:   addStatus,
			Priority: addPriority,
		})
	default:
		fmt.Printf("Created task %s: %s\n", id, filePath)
		return nil
	}
}

func openInEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("$EDITOR is not set; cannot open file for editing")
	}
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
