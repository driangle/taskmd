package cli

import (
	"testing"
)

// TestResolveTaskDir_EnvVarOverridesConfig verifies that TASKMD_TASK_DIR is
// honored as the shared-tracker switch, overriding per-worktree committed
// config (the taskDir fallback here stands in for a config value).
func TestResolveTaskDir_EnvVarOverridesConfig(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper()
	defer resetFlags()

	taskDir = "/some/local/tasks"
	t.Setenv("TASKMD_TASK_DIR", "/shared/tasks")

	got := resolveTaskDir()
	if want := "/shared/tasks"; got != want {
		t.Errorf("resolveTaskDir() = %q, want %q (env should override config)", got, want)
	}
}

// TestResolveTaskDir_ExplicitFlagBeatsEnvVar verifies that an explicit
// --task-dir CLI flag (a deliberate per-invocation override) wins over the
// ambient TASKMD_TASK_DIR env var.
func TestResolveTaskDir_ExplicitFlagBeatsEnvVar(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper()
	defer resetFlags()

	t.Setenv("TASKMD_TASK_DIR", "/shared/tasks")

	tf := rootCmd.PersistentFlags().Lookup("task-dir")
	if err := tf.Value.Set("/explicit/tasks"); err != nil {
		t.Fatalf("set task-dir flag: %v", err)
	}
	tf.Changed = true
	defer func() {
		_ = tf.Value.Set(".")
		tf.Changed = false
	}()

	got := resolveTaskDir()
	if want := "/explicit/tasks"; got != want {
		t.Errorf("resolveTaskDir() = %q, want %q (explicit flag should beat env)", got, want)
	}
}

// TestResolveTaskDir_EnvVarUnsetFallsThrough verifies that when the env var is
// absent, resolution falls through to the existing behavior.
func TestResolveTaskDir_EnvVarUnsetFallsThrough(t *testing.T) {
	resetViper()
	resetFlags()
	defer resetViper()
	defer resetFlags()

	taskDir = "/local/tasks"

	got := resolveTaskDir()
	if want := "/local/tasks"; got != want {
		t.Errorf("resolveTaskDir() = %q, want %q", got, want)
	}
}
