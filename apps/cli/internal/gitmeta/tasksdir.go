package gitmeta

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configFilename = ".taskmd.yaml"
	defaultTaskDir = "tasks"
)

// worktreeDirConfig is the minimal slice of .taskmd.yaml needed to locate a
// worktree's tasks directory. Read with raw yaml, never viper, so sibling
// worktree configs cannot pollute the running command's global config state
// (the same pattern as resolveProjectScanDir in internal/cli).
type worktreeDirConfig struct {
	Dir     string `yaml:"dir"`
	TaskDir string `yaml:"task-dir"`
}

// resolveTasksDir reads root's .taskmd.yaml and returns the absolute tasks
// directory, resolving a relative dir/task-dir against the worktree root and
// falling back to <root>/tasks when the config is unreadable or silent.
func resolveTasksDir(root string) string {
	fallback := filepath.Join(root, defaultTaskDir)

	data, err := os.ReadFile(filepath.Join(root, configFilename))
	if err != nil {
		return fallback
	}

	var cfg worktreeDirConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		Warnf("gitmeta: unparseable %s in %s: %v", configFilename, root, err)
		return fallback
	}

	dir := cfg.TaskDir
	if dir == "" {
		dir = cfg.Dir
	}
	if dir == "" {
		return fallback
	}
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Join(root, dir)
}
