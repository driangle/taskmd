package verify

import (
	"testing"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestRun_BashPass(t *testing.T) {
	steps := []model.VerifyStep{
		{Type: "bash", Run: "echo hello"},
	}
	result := Run(steps, Options{})

	if len(result.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(result.Steps))
	}
	if result.Steps[0].Status != StatusPass {
		t.Errorf("expected pass, got %s", result.Steps[0].Status)
	}
	if result.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", result.Passed)
	}
	if result.HasFailures() {
		t.Error("expected no failures")
	}
}

func TestRun_BashFail(t *testing.T) {
	steps := []model.VerifyStep{
		{Type: "bash", Run: "exit 1"},
	}
	result := Run(steps, Options{})

	if result.Steps[0].Status != StatusFail {
		t.Errorf("expected fail, got %s", result.Steps[0].Status)
	}
	if result.Steps[0].ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.Steps[0].ExitCode)
	}
	if result.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", result.Failed)
	}
	if !result.HasFailures() {
		t.Error("expected failures")
	}
}

func TestRun_BashTimeout(t *testing.T) {
	steps := []model.VerifyStep{
		{Type: "bash", Run: "sleep 10"},
	}
	result := Run(steps, Options{Timeout: 100 * time.Millisecond})

	if result.Steps[0].Status != StatusFail {
		t.Errorf("expected fail, got %s", result.Steps[0].Status)
	}
}

func TestRun_BashCustomDir(t *testing.T) {
	dir := t.TempDir()
	steps := []model.VerifyStep{
		{Type: "bash", Run: "pwd"},
	}
	result := Run(steps, Options{ProjectRoot: dir})

	if result.Steps[0].Status != StatusPass {
		t.Errorf("expected pass, got %s", result.Steps[0].Status)
	}
	// stdout should contain the temp dir path
	if result.Steps[0].Stdout == "" {
		t.Error("expected stdout to contain working directory")
	}
}

func TestRun_BashStepDir(t *testing.T) {
	dir := t.TempDir()
	steps := []model.VerifyStep{
		{Type: "bash", Run: "pwd", Dir: "."},
	}
	result := Run(steps, Options{ProjectRoot: dir})

	if result.Steps[0].Status != StatusPass {
		t.Errorf("expected pass, got %s", result.Steps[0].Status)
	}
}

func TestRun_DryRun(t *testing.T) {
	steps := []model.VerifyStep{
		{Type: "bash", Run: "echo hello"},
		{Type: "assert", Check: "something is true"},
	}
	result := Run(steps, Options{DryRun: true})

	if result.Steps[0].Status != StatusSkip {
		t.Errorf("expected skip for bash in dry-run, got %s", result.Steps[0].Status)
	}
	// assert steps are always pending regardless of dry-run
	if result.Steps[1].Status != StatusPending {
		t.Errorf("expected pending for assert, got %s", result.Steps[1].Status)
	}
}

func TestRun_AssertStep(t *testing.T) {
	steps := []model.VerifyStep{
		{Type: "assert", Check: "The output contains expected data"},
	}
	result := Run(steps, Options{})

	if result.Steps[0].Status != StatusPending {
		t.Errorf("expected pending, got %s", result.Steps[0].Status)
	}
	if result.Steps[0].Check != "The output contains expected data" {
		t.Errorf("expected check text to be preserved")
	}
	if result.Pending != 1 {
		t.Errorf("expected 1 pending, got %d", result.Pending)
	}
}

func TestRun_UnknownType(t *testing.T) {
	steps := []model.VerifyStep{
		{Type: "http", Run: "https://example.com"},
	}
	result := Run(steps, Options{})

	if result.Steps[0].Status != StatusSkip {
		t.Errorf("expected skip, got %s", result.Steps[0].Status)
	}
	if result.Steps[0].Warning == "" {
		t.Error("expected warning for unknown type")
	}
	if result.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Skipped)
	}
}

func TestRun_Empty(t *testing.T) {
	result := Run(nil, Options{})

	if len(result.Steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(result.Steps))
	}
	if result.HasFailures() {
		t.Error("empty result should not have failures")
	}
}

func TestRun_Mixed(t *testing.T) {
	steps := []model.VerifyStep{
		{Type: "bash", Run: "echo pass"},
		{Type: "bash", Run: "exit 1"},
		{Type: "assert", Check: "something"},
		{Type: "unknown"},
	}
	result := Run(steps, Options{})

	if result.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", result.Passed)
	}
	if result.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", result.Failed)
	}
	if result.Pending != 1 {
		t.Errorf("expected 1 pending, got %d", result.Pending)
	}
	if result.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Skipped)
	}
}

func TestRun_LogFunc(t *testing.T) {
	var logged []string
	steps := []model.VerifyStep{
		{Type: "bash", Run: "echo test"},
	}
	Run(steps, Options{
		LogFunc: func(format string, _ ...any) {
			logged = append(logged, format)
		},
	})

	if len(logged) != 1 {
		t.Errorf("expected 1 log call, got %d", len(logged))
	}
}

func TestRun_StdoutCaptured(t *testing.T) {
	steps := []model.VerifyStep{
		{Type: "bash", Run: "echo hello-world"},
	}
	result := Run(steps, Options{})

	if result.Steps[0].Stdout != "hello-world\n" {
		t.Errorf("expected stdout 'hello-world\\n', got %q", result.Steps[0].Stdout)
	}
}

func TestRun_StderrCaptured(t *testing.T) {
	steps := []model.VerifyStep{
		{Type: "bash", Run: "echo error-msg >&2 && exit 1"},
	}
	result := Run(steps, Options{})

	if result.Steps[0].Stderr != "error-msg\n" {
		t.Errorf("expected stderr 'error-msg\\n', got %q", result.Steps[0].Stderr)
	}
}
