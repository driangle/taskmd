//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runWithHome executes the taskmd binary with a specific HOME directory.
// This allows tests to place a .taskmd.yaml in the home dir and verify
// that global config is picked up.
func runWithHome(t *testing.T, home, dir string, args ...string) runResult {
	t.Helper()

	cmd := buildCmd(dir, args...)

	cmd.Env = []string{
		"HOME=" + home,
		"NO_COLOR=1",
		"PATH=" + os.Getenv("PATH"),
	}

	return execCmd(t, cmd, args)
}

// --- Project-level config tests ---

func TestConfig_ProjectTaskDir(t *testing.T) {
	// A project-level .taskmd.yaml with task-dir should make commands
	// scan that subdirectory instead of ".".
	root := t.TempDir()

	// Create a subdirectory with a task file.
	tasksDir := filepath.Join(root, "my-tasks")
	writeTask(t, tasksDir, "001-alpha.md", "001", "Alpha Task", "pending", nil)

	// Write project config pointing task-dir at the subdirectory.
	writeConfig(t, root, "task-dir: my-tasks\n")

	result := mustRun(t, root, "list")

	if !strings.Contains(result.Stdout, "Alpha Task") {
		t.Errorf("expected project config task-dir to find task, got:\n%s", result.Stdout)
	}
}

func TestConfig_ProjectDirLegacy(t *testing.T) {
	// The legacy "dir" key in config should also work.
	root := t.TempDir()

	tasksDir := filepath.Join(root, "legacy-tasks")
	writeTask(t, tasksDir, "001-beta.md", "001", "Beta Task", "pending", nil)

	writeConfig(t, root, "dir: legacy-tasks\n")

	result := mustRun(t, root, "list")

	if !strings.Contains(result.Stdout, "Beta Task") {
		t.Errorf("expected legacy dir config to find task, got:\n%s", result.Stdout)
	}
}

func TestConfig_ProjectVerbose(t *testing.T) {
	// Setting verbose: true in project config should enable verbose output
	// (scanner logs printed to stderr).
	root := t.TempDir()

	writeTask(t, root, "001-test.md", "001", "Test Task", "pending", nil)
	writeConfig(t, root, "dir: .\nverbose: true\n")

	result := mustRun(t, root, "list")

	// Verbose mode causes scanner to log details to stderr.
	if !strings.Contains(result.Stderr, "Scanning directory:") {
		t.Errorf("expected verbose config to produce scanner logs on stderr, got stderr:\n%s", result.Stderr)
	}

	// Without verbose, stderr should be empty.
	rootQuiet := t.TempDir()
	writeTask(t, rootQuiet, "001-test.md", "001", "Test Task", "pending", nil)
	writeConfig(t, rootQuiet, "dir: .\n")

	quietResult := mustRun(t, rootQuiet, "list")

	if strings.Contains(quietResult.Stderr, "Scanning directory:") {
		t.Errorf("expected no verbose output without verbose config, got stderr:\n%s", quietResult.Stderr)
	}
}

// --- Home-level config tests ---

func TestConfig_HomeFallback(t *testing.T) {
	// When no project-level config exists, the home-level config should
	// be used as a fallback. task-dir resolves relative to the config file.
	root := t.TempDir()
	homeDir := t.TempDir()

	// Create tasks relative to the home directory (where the config lives).
	tasksDir := filepath.Join(homeDir, "home-tasks")
	writeTask(t, tasksDir, "001-gamma.md", "001", "Gamma Task", "pending", nil)

	// Put the config in the home directory, not the project directory.
	writeConfig(t, homeDir, "task-dir: home-tasks\nverbose: true\n")

	// No .taskmd.yaml in root — should fall back to $HOME/.taskmd.yaml.
	result := runWithHome(t, homeDir, root, "list")

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s",
			result.ExitCode, result.Stdout, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "Gamma Task") {
		t.Errorf("expected home config task-dir to find task, got:\n%s", result.Stdout)
	}
	// Verbose from home config should produce scanner logs.
	if !strings.Contains(result.Stderr, "Scanning directory:") {
		t.Errorf("expected home config verbose to produce scanner logs, got stderr:\n%s", result.Stderr)
	}
}

func TestConfig_ProjectOverridesHome(t *testing.T) {
	// When both project and home configs exist, the project config wins.
	root := t.TempDir()
	homeDir := t.TempDir()

	// Create two task directories.
	projectTasks := filepath.Join(root, "project-tasks")
	homeTasks := filepath.Join(root, "home-tasks")
	writeTask(t, projectTasks, "001-project.md", "001", "Project Task", "pending", nil)
	writeTask(t, homeTasks, "001-home.md", "002", "Home Task", "pending", nil)

	// Project config points to project-tasks.
	writeConfig(t, root, "task-dir: project-tasks\n")

	// Home config points to home-tasks.
	writeConfig(t, homeDir, "task-dir: home-tasks\n")

	result := runWithHome(t, homeDir, root, "list")

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s",
			result.ExitCode, result.Stdout, result.Stderr)
	}
	// Project config should win: we should see "Project Task", not "Home Task".
	if !strings.Contains(result.Stdout, "Project Task") {
		t.Errorf("expected project config to override home config, got:\n%s", result.Stdout)
	}
	if strings.Contains(result.Stdout, "Home Task") {
		t.Errorf("expected home config to NOT be used when project config exists, got:\n%s", result.Stdout)
	}
}

// --- CLI flag override tests ---

func TestConfig_CLIFlagOverridesProjectConfig(t *testing.T) {
	// A --task-dir CLI flag should override the project config value.
	root := t.TempDir()

	// Project config points to "config-tasks".
	configTasks := filepath.Join(root, "config-tasks")
	writeTask(t, configTasks, "001-config.md", "001", "Config Task", "pending", nil)
	writeConfig(t, root, "task-dir: config-tasks\n")

	// CLI flag points to "flag-tasks".
	flagTasks := filepath.Join(root, "flag-tasks")
	writeTask(t, flagTasks, "001-flag.md", "002", "Flag Task", "pending", nil)

	result := mustRun(t, root, "list", "--task-dir", "flag-tasks")

	if !strings.Contains(result.Stdout, "Flag Task") {
		t.Errorf("expected --task-dir flag to override config, got:\n%s", result.Stdout)
	}
	if strings.Contains(result.Stdout, "Config Task") {
		t.Errorf("expected config value to NOT be used when flag is set, got:\n%s", result.Stdout)
	}
}

func TestConfig_CLIFlagOverridesHomeConfig(t *testing.T) {
	// A --task-dir CLI flag should override the home config value.
	root := t.TempDir()
	homeDir := t.TempDir()

	// Home config points to "home-tasks".
	homeTasks := filepath.Join(root, "home-tasks")
	writeTask(t, homeTasks, "001-home.md", "001", "Home Task", "pending", nil)
	writeConfig(t, homeDir, "task-dir: home-tasks\n")

	// CLI flag points to "flag-tasks".
	flagTasks := filepath.Join(root, "flag-tasks")
	writeTask(t, flagTasks, "001-flag.md", "002", "Flag Task", "pending", nil)

	// No project config in root.
	result := runWithHome(t, homeDir, root, "list", "--task-dir", "flag-tasks")

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout: %s\nstderr: %s",
			result.ExitCode, result.Stdout, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "Flag Task") {
		t.Errorf("expected --task-dir flag to override home config, got:\n%s", result.Stdout)
	}
	if strings.Contains(result.Stdout, "Home Task") {
		t.Errorf("expected home config to NOT be used when flag is set, got:\n%s", result.Stdout)
	}
}

func TestConfig_VerboseFlagOverridesConfig(t *testing.T) {
	// --verbose flag should work even when config says verbose: false.
	root := t.TempDir()

	writeTask(t, root, "001-test.md", "001", "Test Task", "pending", nil)
	writeConfig(t, root, "dir: .\nverbose: false\n")

	result := mustRun(t, root, "list", "--verbose")

	if !strings.Contains(result.Stderr, "Using config file:") {
		t.Errorf("expected --verbose flag to enable verbose output, got stderr:\n%s", result.Stderr)
	}
}

// --- Default behavior (no config) ---

func TestConfig_NoConfigFile(t *testing.T) {
	// When no .taskmd.yaml exists anywhere, commands should still work
	// with defaults (task-dir = ".").
	root := t.TempDir()

	// Put a task directly in the root (default task-dir = ".").
	writeTask(t, root, "001-default.md", "001", "Default Task", "pending", nil)

	// Don't create any .taskmd.yaml — should use defaults.
	result := mustRun(t, root, "list")

	if !strings.Contains(result.Stdout, "Default Task") {
		t.Errorf("expected default task-dir to scan '.', got:\n%s", result.Stdout)
	}
}

func TestConfig_NoConfigFileNoError(t *testing.T) {
	// Missing config should not produce any error output.
	root := t.TempDir()
	writeTask(t, root, "001-test.md", "001", "Test Task", "pending", nil)

	result := mustRun(t, root, "list")

	// No error or warning about missing config.
	if strings.Contains(result.Stderr, "config") || strings.Contains(result.Stderr, "error") {
		t.Errorf("expected no config-related errors, got stderr:\n%s", result.Stderr)
	}
}

// --- Config options that affect output ---

func TestConfig_WorkflowSetting(t *testing.T) {
	// The workflow setting should be respected from config.
	// We test this indirectly by verifying the config loads without error
	// and the command succeeds.
	root := t.TempDir()

	writeTask(t, root, "001-test.md", "001", "Test Task", "pending", nil)
	writeConfig(t, root, "dir: .\nworkflow: pr-review\n")

	result := mustRun(t, root, "list")

	if !strings.Contains(result.Stdout, "Test Task") {
		t.Errorf("expected list to work with workflow config, got:\n%s", result.Stdout)
	}
}

func TestConfig_IgnoreDirs(t *testing.T) {
	// The "ignore" config option should skip specified directories.
	root := t.TempDir()

	// Create tasks in two directories.
	writeTask(t, root, "001-visible.md", "001", "Visible Task", "pending", nil)
	writeTask(t, filepath.Join(root, "ignored-dir"), "002-hidden.md", "002", "Hidden Task", "pending", nil)

	writeConfig(t, root, "dir: .\nignore:\n  - ignored-dir\n")

	result := mustRun(t, root, "list")

	if !strings.Contains(result.Stdout, "Visible Task") {
		t.Errorf("expected visible task to appear, got:\n%s", result.Stdout)
	}
	if strings.Contains(result.Stdout, "Hidden Task") {
		t.Errorf("expected ignored-dir tasks to be hidden, got:\n%s", result.Stdout)
	}
}

func TestConfig_ExplicitConfigFlag(t *testing.T) {
	// The --config flag should load a specific config file, overriding
	// both project and home configs.
	root := t.TempDir()

	// Create two task directories.
	configTasks := filepath.Join(root, "explicit-tasks")
	projectTasks := filepath.Join(root, "project-tasks")
	writeTask(t, configTasks, "001-explicit.md", "001", "Explicit Task", "pending", nil)
	writeTask(t, projectTasks, "001-project.md", "002", "Project Task", "pending", nil)

	// Project config points to project-tasks.
	writeConfig(t, root, "task-dir: project-tasks\n")

	// Write a separate config file that points to explicit-tasks.
	explicitConfig := filepath.Join(root, "custom-config.yaml")
	if err := os.WriteFile(explicitConfig, []byte("task-dir: explicit-tasks\n"), 0o644); err != nil {
		t.Fatalf("failed to write explicit config: %v", err)
	}

	result := mustRun(t, root, "list", "--config", explicitConfig)

	if !strings.Contains(result.Stdout, "Explicit Task") {
		t.Errorf("expected --config to use explicit config file, got:\n%s", result.Stdout)
	}
	if strings.Contains(result.Stdout, "Project Task") {
		t.Errorf("expected project config to NOT be used with --config, got:\n%s", result.Stdout)
	}
}

// --- Subdirectory discovery tests ---

func TestConfig_SubdirectoryDiscovery(t *testing.T) {
	// Running taskmd from a subdirectory should walk up and find .taskmd.yaml
	// at the project root, then resolve task-dir relative to that config.
	root := t.TempDir()

	// Create config at project root pointing to "my-tasks".
	tasksDir := filepath.Join(root, "my-tasks")
	writeTask(t, tasksDir, "001-alpha.md", "001", "Alpha Task", "pending", nil)
	writeConfig(t, root, "task-dir: my-tasks\n")

	// Create a subdirectory to run from.
	subDir := filepath.Join(root, "src", "pkg")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Run from the subdirectory — should walk up and find config + tasks.
	result := mustRun(t, subDir, "list")

	if !strings.Contains(result.Stdout, "Alpha Task") {
		t.Errorf("expected subdirectory discovery to find tasks, got:\n%s\nstderr: %s", result.Stdout, result.Stderr)
	}
}

func TestConfig_SubdirectorySetCommand(t *testing.T) {
	// The set command should also work from a subdirectory.
	root := t.TempDir()

	tasksDir := filepath.Join(root, "tasks")
	writeTask(t, tasksDir, "001-todo.md", "001", "Todo Task", "pending", nil)
	writeConfig(t, root, "task-dir: tasks\n")

	subDir := filepath.Join(root, "apps", "cli")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Run set from subdirectory.
	result := run(t, subDir, "set", "001", "--status", "in-progress")
	if result.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %s\nstderr: %s",
			result.ExitCode, result.Stdout, result.Stderr)
	}

	// Verify the task was updated by listing from subdirectory.
	listResult := mustRun(t, subDir, "list")
	if !strings.Contains(listResult.Stdout, "in-progress") {
		t.Errorf("expected task to be in-progress after set, got:\n%s", listResult.Stdout)
	}
}

func TestConfig_SubdirectoryStopsAtGit(t *testing.T) {
	// Walk-up should stop at .git boundaries.
	// Create an outer project with config, a nested project with .git and
	// no config, and verify the nested project does NOT pick up the outer config.
	outer := t.TempDir()

	// Outer project has config + tasks.
	outerTasks := filepath.Join(outer, "outer-tasks")
	writeTask(t, outerTasks, "001-outer.md", "001", "Outer Task", "pending", nil)
	writeConfig(t, outer, "task-dir: outer-tasks\n")

	// Inner project has a .git dir (boundary) but no config.
	inner := filepath.Join(outer, "nested", "project")
	if err := os.MkdirAll(filepath.Join(inner, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create inner .git: %v", err)
	}

	// Running from inside the inner project should NOT find the outer config.
	// With no config and default task-dir ".", it should scan the inner dir
	// which has no tasks — resulting in an empty list (no "Outer Task").
	result := run(t, inner, "list")

	if strings.Contains(result.Stdout, "Outer Task") {
		t.Errorf("expected .git boundary to prevent walk-up, but found outer task:\n%s", result.Stdout)
	}
}

func TestConfig_SubdirectoryDefaultTaskDir(t *testing.T) {
	// When config has no task-dir (defaults to "."), running from a subdirectory
	// should still find the config and scan relative to the config location.
	root := t.TempDir()

	// Config with no task-dir setting — tasks are at the root.
	writeTask(t, root, "001-root.md", "001", "Root Task", "pending", nil)
	writeConfig(t, root, "# empty config\n")

	subDir := filepath.Join(root, "deep", "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// From subdir, the config is found but task-dir defaults to "." which
	// is relative to cwd. Since there's no task-dir in config, the default
	// "." means cwd, which has no tasks.
	result := run(t, subDir, "list")

	// This should NOT find tasks (default "." is relative to cwd, not config).
	// This test documents the expected behavior: only explicit task-dir
	// in config gets resolved relative to config.
	if strings.Contains(result.Stdout, "Root Task") {
		t.Errorf("expected default task-dir to use cwd, not config dir, got:\n%s", result.Stdout)
	}
}

// --- worklogs ---

// The worklogs key gates `taskmd worklog --add`. Worklogs are opt-in: only
// `worklogs: true` permits a write, and an absent key means disabled. Both
// values must also validate cleanly — the key was documented long before it
// was recognized.

func TestConfig_WorklogsDisabled(t *testing.T) {
	for _, tc := range []struct{ name, config string }{
		{"explicitly disabled", "dir: .\nworklogs: false\n"},
		{"absent key", "dir: .\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeTask(t, root, "001-alpha.md", "001", "Alpha Task", "in-progress", nil)
			writeConfig(t, root, tc.config)

			result := run(t, root, "worklog", "001", "--add", "should not be written")

			if result.ExitCode == 0 {
				t.Errorf("expected a non-zero exit when worklogs are disabled, got:\n%s", result.Stdout)
			}
			if !strings.Contains(result.Stderr, "worklogs are disabled") {
				t.Errorf("expected the error to explain why, got stderr:\n%s", result.Stderr)
			}
			if !strings.Contains(result.Stderr, "worklogs: true") {
				t.Errorf("expected the error to say how to enable them, got stderr:\n%s", result.Stderr)
			}
			if _, err := os.Stat(filepath.Join(root, ".worklogs", "001.md")); !os.IsNotExist(err) {
				t.Error("expected no worklog file to be written")
			}
		})
	}
}

func TestConfig_WorklogsEnabled(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "001-alpha.md", "001", "Alpha Task", "in-progress", nil)
	writeConfig(t, root, "dir: .\nworklogs: true\n")

	mustRun(t, root, "worklog", "001", "--add", "an entry")

	data, err := os.ReadFile(filepath.Join(root, ".worklogs", "001.md"))
	if err != nil {
		t.Fatalf("expected a worklog file to be written: %v", err)
	}
	if !strings.Contains(string(data), "an entry") {
		t.Errorf("expected the entry text in the worklog, got:\n%s", data)
	}
}

func TestConfig_WorklogsValidatesCleanly(t *testing.T) {
	for _, value := range []string{"true", "false"} {
		t.Run(value, func(t *testing.T) {
			root := t.TempDir()
			writeTask(t, root, "001-alpha.md", "001", "Alpha Task", "pending", nil)
			writeConfig(t, root, "dir: .\nworklogs: "+value+"\n")

			result := mustRun(t, root, "validate")

			if strings.Contains(result.Stdout+result.Stderr, "unknown config key") {
				t.Errorf("worklogs: %s should be a known config key, got:\n%s", value, result.Stdout)
			}
		})
	}
}

// --- Helper ---

// writeConfig creates a .taskmd.yaml file in the given directory.
func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create config dir %s: %v", dir, err)
	}
	path := filepath.Join(dir, ".taskmd.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config %s: %v", path, err)
	}
}
