package cli

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/verify"
)

// verifyFiles is the verify-specific task set covering verify-block permutations
// (passing, failing, assert, mixed, no-verify, unknown-type, custom-dir,
// fail-fast). Kept inline because these one-off verify shapes are the subject of
// the tests below.
func verifyFiles() map[string]string {
	return map[string]string{
		"001-pass.md": `---
id: "001"
title: "Task with passing checks"
status: pending
created: 2026-02-14
verify:
  - type: bash
    run: "echo hello"
  - type: bash
    run: "echo world"
---

# Task with passing checks
`,
		"002-fail.md": `---
id: "002"
title: "Task with failing check"
status: pending
created: 2026-02-14
verify:
  - type: bash
    run: "exit 1"
---

# Task with failing check
`,
		"003-assert.md": `---
id: "003"
title: "Task with assert check"
status: pending
created: 2026-02-14
verify:
  - type: assert
    check: "The output contains expected data"
---

# Task with assert check
`,
		"004-mixed.md": `---
id: "004"
title: "Task with mixed checks"
status: pending
created: 2026-02-14
verify:
  - type: bash
    run: "echo pass"
  - type: bash
    run: "exit 1"
  - type: assert
    check: "Something should be true"
---

# Task with mixed checks
`,
		"005-no-verify.md": `---
id: "005"
title: "Task without verify"
status: pending
created: 2026-02-14
---

# Task without verify
`,
		"006-unknown.md": `---
id: "006"
title: "Task with unknown type"
status: pending
created: 2026-02-14
verify:
  - type: http
    run: "https://example.com"
---

# Task with unknown type
`,
		"007-dir.md": `---
id: "007"
title: "Task with custom dir"
status: pending
created: 2026-02-14
verify:
  - type: bash
    run: "pwd"
    dir: "."
---

# Task with custom dir
`,
		"008-failfast.md": `---
id: "008"
title: "Task for fail-fast testing"
status: pending
created: 2026-02-14
verify:
  - type: bash
    run: "exit 1"
  - type: bash
    run: "echo should-not-run"
  - type: bash
    run: "echo also-skipped"
---

# Task for fail-fast testing
`,
	}
}

// verifyCombined runs `verify <args...>` against repo and returns combined
// stdout+stderr plus the command error, mirroring the pre-migration capture
// helper (verify prints results to both streams).
func verifyCombined(t *testing.T, repo *taskRepo, args ...string) (string, error) {
	t.Helper()
	res := repo.Run(append([]string{"verify"}, args...)...)
	return res.Stdout + res.Stderr, res.Err
}

func TestVerify_AllPass(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	output, err := verifyCombined(t, repo, "001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "PASS") {
		t.Errorf("expected PASS in output, got: %s", output)
	}
	if !strings.Contains(output, "2 passed") {
		t.Errorf("expected '2 passed' in output, got: %s", output)
	}
}

func TestVerify_Fail(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	output, err := verifyCombined(t, repo, "002")
	if err == nil {
		t.Fatal("expected error for failing check")
	}
	if !errors.Is(err, ErrVerifyFailed) {
		t.Fatalf("expected ErrVerifyFailed, got: %v", err)
	}

	if !strings.Contains(output, "FAIL") {
		t.Errorf("expected FAIL in output, got: %s", output)
	}
	if !strings.Contains(output, "1 failed") {
		t.Errorf("expected '1 failed' in output, got: %s", output)
	}
}

func TestVerify_Assert(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	output, err := verifyCombined(t, repo, "003")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "PEND") {
		t.Errorf("expected PEND in output, got: %s", output)
	}
	if !strings.Contains(output, "1 pending") {
		t.Errorf("expected '1 pending' in output, got: %s", output)
	}
}

func TestVerify_Mixed(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	// --all runs all steps despite failures
	output, err := verifyCombined(t, repo, "004", "--all")
	if err == nil {
		t.Fatal("expected error for mixed checks with failures")
	}
	if !errors.Is(err, ErrVerifyFailed) {
		t.Fatalf("expected ErrVerifyFailed, got: %v", err)
	}

	if !strings.Contains(output, "PASS") {
		t.Errorf("expected PASS in output, got: %s", output)
	}
	if !strings.Contains(output, "FAIL") {
		t.Errorf("expected FAIL in output, got: %s", output)
	}
	if !strings.Contains(output, "PEND") {
		t.Errorf("expected PEND in output, got: %s", output)
	}
}

func TestVerify_NoField(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	output, err := verifyCombined(t, repo, "005")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "No verification checks defined") {
		t.Errorf("expected no-checks message, got: %s", output)
	}
}

func TestVerify_TaskNotFound(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	_, err := verifyCombined(t, repo, "999")
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("expected 'task not found' error, got: %v", err)
	}
}

func TestVerify_UnknownType(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	output, err := verifyCombined(t, repo, "006")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "SKIP") {
		t.Errorf("expected SKIP in output, got: %s", output)
	}
	if !strings.Contains(output, "1 skipped") {
		t.Errorf("expected '1 skipped' in output, got: %s", output)
	}
}

func TestVerify_DryRun(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	output, err := verifyCombined(t, repo, "001", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "SKIP") {
		t.Errorf("expected SKIP in dry-run output, got: %s", output)
	}
	// Should not contain PASS since nothing was executed
	if strings.Contains(output, "PASS") {
		t.Errorf("dry-run should not show PASS, got: %s", output)
	}
}

func TestVerify_JSON(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	res := repo.Run("verify", "001", "--format", "json")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	var result verify.Result
	if err := json.Unmarshal([]byte(res.Stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, res.Stdout)
	}

	if result.Passed != 2 {
		t.Errorf("expected 2 passed in JSON, got %d", result.Passed)
	}
	if len(result.Steps) != 2 {
		t.Errorf("expected 2 steps in JSON, got %d", len(result.Steps))
	}
}

func TestVerify_CustomDir(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	output, err := verifyCombined(t, repo, "007")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "PASS") {
		t.Errorf("expected PASS in output, got: %s", output)
	}
}

func TestVerify_FailFast(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	output, err := verifyCombined(t, repo, "008")
	if err == nil {
		t.Fatal("expected error for failing check")
	}
	if !errors.Is(err, ErrVerifyFailed) {
		t.Fatalf("expected ErrVerifyFailed, got: %v", err)
	}

	if !strings.Contains(output, "FAIL") {
		t.Errorf("expected FAIL in output, got: %s", output)
	}
	if !strings.Contains(output, "SKIP") {
		t.Errorf("expected SKIP for remaining steps, got: %s", output)
	}
	// With fail-fast, only 1 step should fail and 2 should be skipped (not passed)
	if strings.Contains(output, "PASS") {
		t.Errorf("expected no PASS with fail-fast, but got: %s", output)
	}
	if !strings.Contains(output, "1 failed") {
		t.Errorf("expected '1 failed' in output, got: %s", output)
	}
	if !strings.Contains(output, "2 skipped") {
		t.Errorf("expected '2 skipped' in output, got: %s", output)
	}
}

func TestVerify_All(t *testing.T) {
	repo := newTaskRepo(t, verifyFiles())

	output, err := verifyCombined(t, repo, "008", "--all")
	if err == nil {
		t.Fatal("expected error for failing check")
	}
	if !errors.Is(err, ErrVerifyFailed) {
		t.Fatalf("expected ErrVerifyFailed, got: %v", err)
	}

	if !strings.Contains(output, "FAIL") {
		t.Errorf("expected FAIL in output, got: %s", output)
	}
	// With --all, subsequent steps should run (PASS, not SKIP)
	if !strings.Contains(output, "PASS") {
		t.Errorf("expected PASS for subsequent steps with --all, got: %s", output)
	}
	if !strings.Contains(output, "1 failed") {
		t.Errorf("expected '1 failed' in output, got: %s", output)
	}
	if !strings.Contains(output, "2 passed") {
		t.Errorf("expected '2 passed' in output, got: %s", output)
	}
}

func TestResolveProjectRoot_ReturnsAbsolutePath(t *testing.T) {
	// resolveProjectRoot uses viper.ConfigFileUsed() to determine the project root.
	// It should always return an absolute path regardless of how viper found the config.
	root := resolveProjectRoot()

	if !filepath.IsAbs(root) {
		t.Errorf("expected absolute path from resolveProjectRoot(), got relative: %s", root)
	}
}

func TestResolveProjectRoot_MatchesConfigDir(t *testing.T) {
	// When viper has a config file, resolveProjectRoot should return
	// the absolute path to the directory containing that config file.
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		t.Skip("no config file loaded, skipping")
	}

	root := resolveProjectRoot()
	expectedDir, err := filepath.Abs(filepath.Dir(configFile))
	if err != nil {
		t.Fatalf("failed to get abs path: %v", err)
	}

	if root != expectedDir {
		t.Errorf("expected %s, got %s", expectedDir, root)
	}
}

func TestVerify_Timeout(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"010-slow.md": `---
id: "010"
title: "Task with slow check"
status: pending
created: 2026-02-14
verify:
  - type: bash
    run: "sleep 10"
---

# Task with slow check
`,
	})

	output, err := verifyCombined(t, repo, "010", "--timeout", "1")
	if err == nil {
		t.Fatal("expected error for timeout")
	}
	if !errors.Is(err, ErrVerifyFailed) {
		t.Fatalf("expected ErrVerifyFailed, got: %v", err)
	}

	if !strings.Contains(output, "FAIL") {
		t.Errorf("expected FAIL for timeout, got: %s", output)
	}
}
