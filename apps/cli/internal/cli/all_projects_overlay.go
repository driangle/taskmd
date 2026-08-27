package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/model"
)

// projectScopeConfig reads just the worktree scope from a project's
// .taskmd.yaml.
type projectScopeConfig struct {
	WorktreeScope string `yaml:"worktree_scope"`
}

// projectScan is one registered project's scan: its tasks with effective
// statuses substituted, plus the worktree overlay when that project's overlay
// is active (nil otherwise).
type projectScan struct {
	tasks   []*model.Task
	overlay *worktreeOverlay
}

// provenance returns the overlay task for id, or nil when this project's
// overlay is inactive or the id is unknown.
func (s projectScan) provenance(id string) *OverlayTask {
	if s.overlay == nil {
		return nil
	}
	return s.overlay.Get(id)
}

// recommendationInputs returns the tasks to feed the next recommender and the
// worktree exclusions to apply, which are empty when the overlay is inactive.
func (s projectScan) recommendationInputs() ([]*model.Task, map[string]string) {
	if s.overlay == nil {
		return s.tasks, nil
	}
	return s.overlay.RecommendationInputs()
}

// scanProject scans one project and merges its sibling worktrees when that
// project's overlay is active (spec §7: the overlay applies per-repo). Task
// file paths are relative to the project's scan dir, as elsewhere in
// --all-projects output.
func scanProject(entry GlobalProjectEntry) (projectScan, error) {
	info, err := os.Stat(entry.Path)
	if err != nil || !info.IsDir() {
		return projectScan{}, fmt.Errorf("path %q is not accessible", entry.Path)
	}

	scanDir := resolveProjectScanDir(entry.Path)
	result, err := newTaskScanner(scanDir, GlobalFlags{}).Scan()
	if err != nil {
		return projectScan{}, fmt.Errorf("scan failed: %w", err)
	}

	overlay, err := buildProjectOverlay(entry.Path, scanDir, result.Tasks)
	if err != nil {
		return projectScan{}, err
	}

	if overlay == nil {
		makeFilePathsRelative(result.Tasks, scanDir)
		return projectScan{tasks: result.Tasks}, nil
	}

	// Relativize before EffectiveTasks: it hands back shallow copies for
	// tasks whose status changed, which would otherwise keep absolute paths.
	overlay.RelativizeSiblingPaths()
	makeFilePathsRelative(overlay.Local(), scanDir)
	printProjectOverlayWarnings(entry.ID, overlay)
	return projectScan{tasks: overlay.EffectiveTasks(), overlay: overlay}, nil
}

// printProjectOverlayWarnings reports a project's cross-worktree consistency
// warnings, naming the project since several are merged into one view.
func printProjectOverlayWarnings(projectID string, overlay *worktreeOverlay) {
	if GetGlobalFlags().Quiet {
		return
	}
	for _, warning := range overlay.Warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s: %s\n", projectID, warning.Message)
	}
}

// buildProjectOverlay builds the worktree overlay for one registered project,
// or returns nil when it is inactive for that project.
func buildProjectOverlay(projectPath, scanDir string, localTasks []*model.Task) (*worktreeOverlay, error) {
	enabled, err := projectWorktreeScopeEnabled(projectPath)
	if err != nil {
		return nil, err
	}
	builder := worktree.Builder{Enabled: enabled, Discover: discoverSiblingWorktrees}
	overlay, err := builder.Build(scanDir, localTasks)
	if err != nil || overlay == nil {
		return nil, err
	}
	return overlay, nil
}

// projectWorktreeScopeEnabled decides whether the overlay is enabled for one
// registered project. A --worktree-scope flag or TASKMD_WORKTREE_SCOPE env set
// for this invocation is a deliberate per-invocation instruction and overrides
// every project; otherwise each project's own .taskmd.yaml decides, so the
// invoking project's config never leaks into the others.
func projectWorktreeScopeEnabled(projectPath string) (bool, error) {
	if enabled, ok, err := worktreeScopeOverride(); ok || err != nil {
		return enabled, err
	}
	return worktreeScopeEnabled(readProjectWorktreeScope(projectPath))
}

// worktreeScopeOverride reports an invocation-wide scope override from the
// --worktree-scope flag or the TASKMD_WORKTREE_SCOPE env var. ok is false when
// neither is set, leaving the decision to each project's config.
func worktreeScopeOverride() (enabled, ok bool, err error) {
	value := ""
	if f := rootCmd.PersistentFlags().Lookup("worktree-scope"); f != nil && f.Changed {
		value = worktreeScopeFlag
	} else if env := os.Getenv("TASKMD_WORKTREE_SCOPE"); env != "" {
		value = env
	}
	if value == "" {
		return false, false, nil
	}
	enabled, err = worktreeScopeEnabled(value)
	return enabled, true, err
}

// readProjectWorktreeScope reads a project's worktree_scope directly from its
// .taskmd.yaml, bypassing viper so scanning one project never mutates global
// config state (the same pattern as resolveProjectScanDir and the sibling
// tasks-dir reads in gitmeta). An unreadable or malformed file falls back to
// the default scope.
func readProjectWorktreeScope(projectPath string) string {
	data, err := os.ReadFile(filepath.Join(projectPath, ".taskmd.yaml"))
	if err != nil {
		return ""
	}
	var cfg projectScopeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return cfg.WorktreeScope
}

// worktreeScopeEnabled maps a worktree scope value to whether the overlay is
// enabled. The empty value is the unified default.
func worktreeScopeEnabled(value string) (bool, error) {
	switch value {
	case "", worktreeScopeUnified:
		return true, nil
	case worktreeScopeIsolated:
		return false, nil
	default:
		return false, invalidValueError("worktree_scope", value,
			[]string{worktreeScopeUnified, worktreeScopeIsolated})
	}
}
