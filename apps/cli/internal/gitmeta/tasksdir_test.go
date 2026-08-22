package gitmeta

import (
	"path/filepath"
	"testing"
)

func TestResolveTasksDir(t *testing.T) {
	abs := t.TempDir()

	tests := []struct {
		name   string
		config string
		want   func(root string) string
	}{
		{
			name:   "dir key relative",
			config: "dir: ./my-tasks\n",
			want:   func(root string) string { return filepath.Join(root, "my-tasks") },
		},
		{
			name:   "task-dir key wins over dir",
			config: "dir: ./ignored\ntask-dir: preferred\n",
			want:   func(root string) string { return filepath.Join(root, "preferred") },
		},
		{
			name:   "absolute dir kept as-is",
			config: "dir: " + abs + "\n",
			want:   func(string) string { return abs },
		},
		{
			name:   "config without dir keys falls back to tasks",
			config: "worklogs: true\n",
			want:   func(root string) string { return filepath.Join(root, defaultTaskDir) },
		},
		{
			name:   "unparseable config falls back to tasks",
			config: ":\t{not yaml\n",
			want:   func(root string) string { return filepath.Join(root, defaultTaskDir) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := makeWorktreeDir(t, tt.config)
			if got := resolveTasksDir(root); got != tt.want(root) {
				t.Errorf("resolveTasksDir = %q, want %q", got, tt.want(root))
			}
		})
	}
}

func TestResolveTasksDir_MissingConfig(t *testing.T) {
	root := t.TempDir()
	if got := resolveTasksDir(root); got != filepath.Join(root, defaultTaskDir) {
		t.Errorf("resolveTasksDir = %q, want fallback %q", got, filepath.Join(root, defaultTaskDir))
	}
}
