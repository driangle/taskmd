package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
	"github.com/driangle/taskmd/sdk/go/validator"
)

func TestParseScopeEntries_WithDescription(t *testing.T) {
	scopeMap := map[string]any{
		"cli/graph": map[string]any{
			"description": "Graph visualization",
			"paths":       []any{"apps/cli/internal/graph/"},
		},
	}

	scopes := parseScopeEntries(scopeMap)

	sc, ok := scopes["cli/graph"]
	if !ok {
		t.Fatal("expected scope cli/graph to exist")
	}
	if sc.Description != "Graph visualization" {
		t.Errorf("Description = %q, want %q", sc.Description, "Graph visualization")
	}
	if len(sc.Paths) != 1 || sc.Paths[0] != "apps/cli/internal/graph/" {
		t.Errorf("Paths = %v, want [apps/cli/internal/graph/]", sc.Paths)
	}
}

func TestParseScopeEntries_WithoutDescription(t *testing.T) {
	scopeMap := map[string]any{
		"cli/output": map[string]any{
			"paths": []any{"apps/cli/internal/cli/format.go"},
		},
	}

	scopes := parseScopeEntries(scopeMap)

	sc, ok := scopes["cli/output"]
	if !ok {
		t.Fatal("expected scope cli/output to exist")
	}
	if sc.Description != "" {
		t.Errorf("Description = %q, want empty string", sc.Description)
	}
	if len(sc.Paths) != 1 {
		t.Errorf("Paths = %v, want 1 element", sc.Paths)
	}
}

// --- Helpers ---

// validateFiles is the recurring valid-task fixture for validate command tests:
// two valid tasks where 002 depends on 001. Command-specific, so it stays inline.
func validateFiles() map[string]string {
	return map[string]string{
		"001-valid.md": `---
id: "001"
title: "Valid Task"
status: pending
priority: high
effort: small
dependencies: []
tags: ["test"]
created: 2026-02-08
---

A valid task.
`,
		"002-valid.md": `---
id: "002"
title: "Another Valid Task"
status: completed
priority: medium
effort: medium
dependencies: ["001"]
tags: ["test"]
created: 2026-02-08
---

Another valid task.
`,
	}
}

// --- Unit tests (no file I/O) ---

func TestMergeValidationResults(t *testing.T) {
	target := &validator.ValidationResult{
		Issues:   []validator.ValidationIssue{{Level: validator.LevelError, Message: "err1"}},
		Errors:   1,
		Warnings: 0,
	}
	source := &validator.ValidationResult{
		Issues:   []validator.ValidationIssue{{Level: validator.LevelWarning, Message: "warn1"}},
		Errors:   0,
		Warnings: 1,
	}

	mergeValidationResults(target, source)

	if len(target.Issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(target.Issues))
	}
	if target.Errors != 1 {
		t.Errorf("expected 1 error, got %d", target.Errors)
	}
	if target.Warnings != 1 {
		t.Errorf("expected 1 warning, got %d", target.Warnings)
	}
}

func TestOutputValidationText_NoIssues(t *testing.T) {
	result := &validator.ValidationResult{
		Issues:    []validator.ValidationIssue{},
		TaskCount: 5,
	}

	output, _ := captureOutput(t, func() {
		outputValidationText(result, false)
	})

	if !strings.Contains(output, "5 task(s) are valid") {
		t.Errorf("expected success message, got:\n%s", output)
	}
}

func TestOutputValidationText_WithErrors(t *testing.T) {
	result := &validator.ValidationResult{
		Issues: []validator.ValidationIssue{
			{Level: validator.LevelError, TaskID: "001", Message: "missing title"},
			{Level: validator.LevelError, TaskID: "002", Message: "bad status"},
		},
		Errors:    2,
		TaskCount: 3,
	}

	output, _ := captureOutput(t, func() {
		outputValidationText(result, false)
	})

	if !strings.Contains(output, "2 error(s)") {
		t.Errorf("expected error count, got:\n%s", output)
	}
}

func TestOutputValidationText_WithWarnings(t *testing.T) {
	result := &validator.ValidationResult{
		Issues: []validator.ValidationIssue{
			{Level: validator.LevelWarning, TaskID: "001", Message: "no priority"},
		},
		Warnings:  1,
		TaskCount: 2,
	}

	output, _ := captureOutput(t, func() {
		outputValidationText(result, false)
	})

	if !strings.Contains(output, "1 warning(s)") {
		t.Errorf("expected warning count, got:\n%s", output)
	}
}

func TestOutputValidationText_Quiet(t *testing.T) {
	result := &validator.ValidationResult{
		Issues:    []validator.ValidationIssue{},
		TaskCount: 3,
	}

	output, _ := captureOutput(t, func() {
		outputValidationText(result, true)
	})

	if output != "" {
		t.Errorf("expected no output in quiet mode, got:\n%s", output)
	}
}

func TestPrintIssue_WithTaskID(t *testing.T) {
	issue := validator.ValidationIssue{
		Level:   validator.LevelError,
		TaskID:  "042",
		Message: "something wrong",
	}

	output, _ := captureOutput(t, func() {
		printIssue(issue, getRenderer())
	})

	if !strings.Contains(output, "042") {
		t.Errorf("expected task ID in output, got:\n%s", output)
	}
	if !strings.Contains(output, "something wrong") {
		t.Errorf("expected message in output, got:\n%s", output)
	}
}

func TestPrintIssue_WithoutTaskID(t *testing.T) {
	issue := validator.ValidationIssue{
		Level:   validator.LevelError,
		Message: "global error",
	}

	output, _ := captureOutput(t, func() {
		printIssue(issue, getRenderer())
	})

	if !strings.Contains(output, "global error") {
		t.Errorf("expected message in output, got:\n%s", output)
	}
	// Should not contain bracket prefix for task ID
	if strings.Contains(output, "[") && strings.Contains(output, "]") {
		t.Errorf("expected no [ID] prefix, got:\n%s", output)
	}
}

func TestPrintIssue_WithFilePath(t *testing.T) {
	issue := validator.ValidationIssue{
		Level:    validator.LevelError,
		TaskID:   "001",
		FilePath: "tasks/001-test.md",
		Message:  "some issue",
	}

	output, _ := captureOutput(t, func() {
		printIssue(issue, getRenderer())
	})

	if !strings.Contains(output, "tasks/001-test.md") {
		t.Errorf("expected file path in output, got:\n%s", output)
	}
}

func TestOutputValidationJSON(t *testing.T) {
	result := &validator.ValidationResult{
		Issues: []validator.ValidationIssue{
			{Level: validator.LevelError, TaskID: "001", Message: "missing title"},
		},
		Errors:    1,
		TaskCount: 2,
	}

	var jsonErr error
	output, _ := captureOutput(t, func() {
		jsonErr = outputValidationJSON(result)
	})

	if jsonErr != nil {
		t.Fatalf("outputValidationJSON failed: %v", jsonErr)
	}

	var parsed validator.ValidationResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, output)
	}

	if parsed.Errors != 1 {
		t.Errorf("errors = %d, want 1", parsed.Errors)
	}
	if parsed.TaskCount != 2 {
		t.Errorf("task_count = %d, want 2", parsed.TaskCount)
	}
	if len(parsed.Issues) != 1 {
		t.Errorf("issues count = %d, want 1", len(parsed.Issues))
	}
}

// --- Command-level tests ---

func TestRunValidate_ValidTasks(t *testing.T) {
	repo := newTaskRepo(t, validateFiles())

	res := repo.Run("validate")
	if res.Err != nil {
		t.Fatalf("runValidate failed: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "2 task(s) are valid") {
		t.Errorf("expected success message, got:\n%s", res.Stdout)
	}
}

func TestRunValidate_JSONFormat(t *testing.T) {
	repo := newTaskRepo(t, validateFiles())

	res := repo.Run("validate", "--format", "json")
	if res.Err != nil {
		t.Fatalf("runValidate failed: %v", res.Err)
	}

	var parsed validator.ValidationResult
	if err := json.Unmarshal([]byte(res.Stdout), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	if parsed.TaskCount != 2 {
		t.Errorf("task_count = %d, want 2", parsed.TaskCount)
	}
}

func TestRunValidate_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, validateFiles())

	res := repo.Run("validate", "--format", "invalid")
	if res.Err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestRunValidate_StrictMode_NoWarnings(t *testing.T) {
	// Valid tasks with ALL optional fields filled — strict mode should produce no warnings
	repo := newTaskRepo(t, map[string]string{
		"001-complete.md": `---
id: "001"
title: "Complete Task"
status: pending
priority: high
effort: small
group: "backend"
tags: ["test"]
created: 2026-02-08
---

A fully specified task with body content.
`,
	})

	res := repo.Run("validate", "--strict", "--format", "json")
	if res.Err != nil {
		t.Fatalf("runValidate failed: %v", res.Err)
	}

	var parsed validator.ValidationResult
	if err := json.Unmarshal([]byte(res.Stdout), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	if parsed.Warnings != 0 {
		t.Errorf("expected 0 warnings for fully specified task in strict mode, got %d", parsed.Warnings)
		for _, issue := range parsed.Issues {
			if issue.Level == validator.LevelWarning {
				t.Logf("  warning: %s", issue.Message)
			}
		}
	}
}

func TestRunValidate_StrictMode_WithWarnings_Text(t *testing.T) {
	// Test strict validation with text output (warnings are displayed, but os.Exit(2) is called).
	// We test the strict path indirectly by validating through the validator directly.
	v := validator.NewValidator(true)
	tasks := []*model.Task{
		{ID: "001", Title: "Minimal", Status: "pending"},
	}

	result := v.Validate(tasks)

	if result.Warnings == 0 {
		t.Error("expected strict warnings for task missing optional fields")
	}

	expectedWarnings := []string{"no priority", "no effort", "no group", "no tags", "no description"}
	warningMsgs := make([]string, 0)
	for _, issue := range result.Issues {
		if issue.Level == validator.LevelWarning {
			warningMsgs = append(warningMsgs, issue.Message)
		}
	}
	for _, expected := range expectedWarnings {
		found := false
		for _, msg := range warningMsgs {
			if strings.Contains(msg, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected warning containing %q, got warnings: %v", expected, warningMsgs)
		}
	}
}

func TestRunValidate_InvalidTasks(t *testing.T) {
	// Test validation errors through the validator directly (runValidate calls os.Exit on errors).
	v := validator.NewValidator(false)
	tasks := []*model.Task{
		{ID: "001", Title: "Bad Task", Status: "banana", Priority: "mega", Effort: "tiny"},
	}

	result := v.Validate(tasks)

	if result.Errors == 0 {
		t.Error("expected errors for invalid field values")
	}

	errorMessages := make([]string, 0)
	for _, issue := range result.Issues {
		if issue.Level == validator.LevelError {
			errorMessages = append(errorMessages, issue.Message)
		}
	}
	if len(errorMessages) < 3 {
		t.Errorf("expected at least 3 errors (status, priority, effort), got %d: %v", len(errorMessages), errorMessages)
	}
}

func TestRunValidate_WithArchive(t *testing.T) {
	// Active task that depends on an archived task; archived task lives under archive/.
	repo := newTaskRepo(t, map[string]string{
		"010-active.md": `---
id: "010"
title: "Active Task"
status: pending
dependencies: ["050"]
---

Depends on archived task.
`,
		"archive/050-archived.md": `---
id: "050"
title: "Archived Task"
status: completed
---

Done.
`,
	})

	res := repo.Run("validate", "--format", "json")
	if res.Err != nil {
		t.Fatalf("runValidate failed: %v", res.Err)
	}

	var parsed validator.ValidationResult
	if err := json.Unmarshal([]byte(res.Stdout), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	// Should NOT have a missing dependency error since "050" is in archive
	for _, issue := range parsed.Issues {
		if strings.Contains(issue.Message, "non-existent task: '050'") {
			t.Error("archived task ID should not trigger missing dependency error")
		}
	}
}

func TestCollectArchivedIDs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create archive directory with task files
	archiveDir := filepath.Join(tmpDir, "archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatalf("failed to create archive dir: %v", err)
	}

	tasks := map[string]string{
		"050-done.md": `---
id: "050"
title: "Archived 050"
status: completed
---
Done.
`,
		"051-done.md": `---
id: "051"
title: "Archived 051"
status: completed
---
Done.
`,
	}
	for name, content := range tasks {
		if err := os.WriteFile(filepath.Join(archiveDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	s := scanner.NewScanner(tmpDir, false, nil)
	ids := collectArchivedIDs(s)

	if len(ids) != 2 {
		t.Fatalf("expected 2 archived IDs, got %d", len(ids))
	}
	if !ids["050"] {
		t.Error("expected archived ID '050'")
	}
	if !ids["051"] {
		t.Error("expected archived ID '051'")
	}
}

func TestCollectArchivedIDs_NoArchive(t *testing.T) {
	tmpDir := t.TempDir()

	s := scanner.NewScanner(tmpDir, false, nil)
	ids := collectArchivedIDs(s)

	if ids != nil {
		t.Errorf("expected nil for no archive, got %v", ids)
	}
}

func TestRunValidate_WithConfig_Scopes(t *testing.T) {
	// Use validateConfig directly with mock config data to avoid viper global state
	v := validator.NewValidator(false)
	tasks := []*model.Task{
		{ID: "001", Title: "Task with touches", Touches: []string{"cli/graph", "undefined-scope"}},
	}
	validationResult := v.Validate(tasks)

	configData := &validator.ConfigData{
		Scopes: map[string]validator.ScopeConfig{
			"cli/graph": {Description: "Graph visualization", Paths: []string{"apps/cli/internal/graph/"}},
		},
		TopKeys:    []string{"scopes"},
		ConfigPath: ".taskmd.yaml",
	}

	validateConfig(v, validationResult, tasks)
	// Reset and test with actual config data
	validationResult2 := v.Validate(tasks)
	mergeValidationResults(validationResult2, v.ValidateConfig(configData))

	knownScopes := map[string]bool{"cli/graph": true}
	mergeValidationResults(validationResult2, v.ValidateTouchesAgainstScopes(tasks, knownScopes))

	// Should have a warning about "undefined-scope"
	foundUndefinedWarning := false
	for _, issue := range validationResult2.Issues {
		if strings.Contains(issue.Message, "undefined-scope") {
			foundUndefinedWarning = true
			break
		}
	}
	if !foundUndefinedWarning {
		t.Error("expected warning about undefined scope 'undefined-scope'")
	}
}

func TestValidateConfig_WithScopes(t *testing.T) {
	v := validator.NewValidator(false)
	tasks := []*model.Task{
		{ID: "001", Title: "Test Task", Touches: []string{"cli/graph", "bad-scope"}},
	}
	validationResult := &validator.ValidationResult{
		Issues:    make([]validator.ValidationIssue, 0),
		TaskCount: 1,
	}

	configData := &validator.ConfigData{
		Scopes: map[string]validator.ScopeConfig{
			"cli/graph": {Description: "Graph", Paths: []string{"apps/cli/internal/graph/"}},
		},
		TopKeys:    []string{"scopes"},
		ConfigPath: ".taskmd.yaml",
	}

	// Merge config validation
	mergeValidationResults(validationResult, v.ValidateConfig(configData))

	// Merge touches validation
	knownScopes := make(map[string]bool, len(configData.Scopes))
	for name := range configData.Scopes {
		knownScopes[name] = true
	}
	mergeValidationResults(validationResult, v.ValidateTouchesAgainstScopes(tasks, knownScopes))

	// Should warn about "bad-scope"
	found := false
	for _, issue := range validationResult.Issues {
		if strings.Contains(issue.Message, "bad-scope") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about 'bad-scope' touches referencing undefined scope")
	}
}

// --- parseIDConfig tests ---

func TestParseIDConfig_FullConfig(t *testing.T) {
	raw := map[string]any{
		"strategy": "prefixed",
		"prefix":   "dr",
		"length":   6,
		"padding":  3,
	}

	cfg := parseIDConfig(raw)
	if cfg == nil {
		t.Fatal("expected non-nil IDConfig")
		return
	}
	if cfg.Strategy != "prefixed" {
		t.Errorf("Strategy = %q, want %q", cfg.Strategy, "prefixed")
	}
	if cfg.Prefix != "dr" {
		t.Errorf("Prefix = %q, want %q", cfg.Prefix, "dr")
	}
	if cfg.Length != 6 {
		t.Errorf("Length = %d, want 6", cfg.Length)
	}
	if cfg.Padding != 3 {
		t.Errorf("Padding = %d, want 3", cfg.Padding)
	}
}

func TestParseIDConfig_PartialConfig(t *testing.T) {
	raw := map[string]any{
		"strategy": "sequential",
	}

	cfg := parseIDConfig(raw)
	if cfg == nil {
		t.Fatal("expected non-nil IDConfig")
		return
	}
	if cfg.Strategy != "sequential" {
		t.Errorf("Strategy = %q, want %q", cfg.Strategy, "sequential")
	}
	if cfg.Prefix != "" {
		t.Errorf("Prefix = %q, want empty", cfg.Prefix)
	}
	if cfg.Length != 0 {
		t.Errorf("Length = %d, want 0", cfg.Length)
	}
	if cfg.Padding != 0 {
		t.Errorf("Padding = %d, want 0", cfg.Padding)
	}
}

func TestParseIDConfig_Nil(t *testing.T) {
	cfg := parseIDConfig(nil)

	if cfg != nil {
		t.Errorf("expected nil, got %+v", cfg)
	}
}

func TestParseIDConfig_NotAMap(t *testing.T) {
	cfg := parseIDConfig("not-a-map")

	if cfg != nil {
		t.Errorf("expected nil for non-map input, got %+v", cfg)
	}
}

func TestParseIDConfig_ViperIntTypes(t *testing.T) {
	// Viper may return int64 or float64 for numeric values
	raw := map[string]any{
		"strategy": "random",
		"length":   int64(8),
		"padding":  float64(4),
	}

	cfg := parseIDConfig(raw)
	if cfg == nil {
		t.Fatal("expected non-nil IDConfig")
		return
	}
	if cfg.Length != 8 {
		t.Errorf("Length = %d, want 8", cfg.Length)
	}
	if cfg.Padding != 4 {
		t.Errorf("Padding = %d, want 4", cfg.Padding)
	}
}

// --- parsePhasesConfig tests ---

func TestParsePhasesConfig_FullConfig(t *testing.T) {
	raw := []any{
		map[string]any{
			"id":          "core-cli",
			"name":        "v0.2",
			"description": "Core CLI features",
			"due":         "2026-04-01",
		},
		map[string]any{
			"id":          "web-dashboard",
			"name":        "v0.3",
			"description": "Web dashboard",
		},
	}

	phases := parsePhasesConfig(raw)
	if len(phases) != 2 {
		t.Fatalf("expected 2 phases, got %d", len(phases))
	}
	if phases[0].ID != "core-cli" {
		t.Errorf("ID = %q, want %q", phases[0].ID, "core-cli")
	}
	if phases[0].Name != "v0.2" {
		t.Errorf("Name = %q, want %q", phases[0].Name, "v0.2")
	}
	if phases[0].Description != "Core CLI features" {
		t.Errorf("Description = %q, want %q", phases[0].Description, "Core CLI features")
	}
	if phases[0].Due.IsZero() {
		t.Error("expected non-zero due date for v0.2")
	}
	if phases[1].ID != "web-dashboard" {
		t.Errorf("ID = %q, want %q", phases[1].ID, "web-dashboard")
	}
	if phases[1].Name != "v0.3" {
		t.Errorf("Name = %q, want %q", phases[1].Name, "v0.3")
	}
	if !phases[1].Due.IsZero() {
		t.Error("expected zero due date for v0.3 (no due set)")
	}
}

func TestParsePhasesConfig_Nil(t *testing.T) {
	phases := parsePhasesConfig(nil)
	if phases != nil {
		t.Errorf("expected nil, got %+v", phases)
	}
}

func TestParsePhasesConfig_NotASlice(t *testing.T) {
	phases := parsePhasesConfig("not-a-slice")
	if phases != nil {
		t.Errorf("expected nil for non-slice input, got %+v", phases)
	}
}

func TestParsePhasesConfig_SkipsNonMapEntries(t *testing.T) {
	raw := []any{
		"not-a-map",
		map[string]any{"id": "v02", "name": "v0.2"},
	}

	phases := parsePhasesConfig(raw)
	if len(phases) != 1 {
		t.Fatalf("expected 1 phase (skipping non-map), got %d", len(phases))
	}
	if phases[0].Name != "v0.2" {
		t.Errorf("Name = %q, want %q", phases[0].Name, "v0.2")
	}
}

func TestParsePhasesConfig_IDFieldParsed(t *testing.T) {
	raw := []any{
		map[string]any{
			"id":   "my-phase",
			"name": "My Phase",
		},
	}

	phases := parsePhasesConfig(raw)
	if len(phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(phases))
	}
	if phases[0].ID != "my-phase" {
		t.Errorf("ID = %q, want %q", phases[0].ID, "my-phase")
	}
}

func TestParsePhasesConfig_NoID_StillIncluded(t *testing.T) {
	raw := []any{
		map[string]any{
			"name": "Legacy Phase",
		},
	}

	phases := parsePhasesConfig(raw)
	if len(phases) != 1 {
		t.Fatalf("expected 1 phase (no id but still included for validation), got %d", len(phases))
	}
	if phases[0].Name != "Legacy Phase" {
		t.Errorf("Name = %q, want %q", phases[0].Name, "Legacy Phase")
	}
	if phases[0].ID != "" {
		t.Errorf("ID = %q, want empty", phases[0].ID)
	}
}

func TestValidateConfig_DuplicatePhaseID(t *testing.T) {
	v := validator.NewValidator(false)
	config := &validator.ConfigData{
		Phases: []validator.PhaseConfig{
			{ID: "alpha", Name: "Alpha"},
			{ID: "alpha", Name: "Beta"},
		},
		TopKeys:    []string{"phases"},
		ConfigPath: ".taskmd.yaml",
	}

	result := v.ValidateConfig(config)

	found := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "duplicate phase id: 'alpha'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about duplicate phase id 'alpha'")
	}
}

func TestValidateConfig_DuplicatePhaseName(t *testing.T) {
	v := validator.NewValidator(false)
	config := &validator.ConfigData{
		Phases: []validator.PhaseConfig{
			{ID: "a", Name: "Same Name"},
			{ID: "b", Name: "Same Name"},
		},
		TopKeys:    []string{"phases"},
		ConfigPath: ".taskmd.yaml",
	}

	result := v.ValidateConfig(config)

	found := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "duplicate phase name: 'Same Name'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about duplicate phase name 'Same Name'")
	}
}

func TestValidatePhasesAgainstConfig_MatchesByID(t *testing.T) {
	v := validator.NewValidator(false)
	tasks := []*model.Task{
		{ID: "001", Title: "Task", Phase: "core-cli"},
	}

	// knownPhases uses ID, not name
	knownPhases := map[string]bool{"core-cli": true}

	result := v.ValidatePhasesAgainstConfig(tasks, knownPhases)

	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "undefined phase") {
			t.Errorf("unexpected warning: %s", issue.Message)
		}
	}
}

func TestValidatePhasesAgainstConfig_FallsBackToName(t *testing.T) {
	// When phases have no ID, validation uses name as key
	v := validator.NewValidator(false)
	tasks := []*model.Task{
		{ID: "001", Title: "Task", Phase: "Legacy Phase"},
	}

	// Simulate a config with no ID — validateConfig falls back to name
	phases := []validator.PhaseConfig{
		{Name: "Legacy Phase"},
	}
	knownPhases := make(map[string]bool)
	for _, m := range phases {
		key := m.ID
		if key == "" {
			key = m.Name
		}
		knownPhases[key] = true
	}

	result := v.ValidatePhasesAgainstConfig(tasks, knownPhases)

	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "undefined phase") {
			t.Errorf("unexpected warning: %s", issue.Message)
		}
	}
}

func TestParseScopeEntries_NonMapEntry(t *testing.T) {
	// Test with a scope value that is not a map (e.g., a string)
	scopeMap := map[string]any{
		"simple": "not-a-map",
	}

	scopes := parseScopeEntries(scopeMap)

	sc, ok := scopes["simple"]
	if !ok {
		t.Fatal("expected scope 'simple' to exist")
	}
	if sc.Paths != nil {
		t.Errorf("expected nil paths for non-map entry, got %v", sc.Paths)
	}
}

func TestParseScopeEntries_MissingPaths(t *testing.T) {
	// Scope entry with no paths key at all
	scopeMap := map[string]any{
		"no-paths": map[string]any{
			"description": "Has description but no paths",
		},
	}

	scopes := parseScopeEntries(scopeMap)

	sc, ok := scopes["no-paths"]
	if !ok {
		t.Fatal("expected scope 'no-paths' to exist")
	}
	if sc.Description != "Has description but no paths" {
		t.Errorf("Description = %q, want %q", sc.Description, "Has description but no paths")
	}
	if sc.Paths != nil {
		t.Errorf("expected nil paths, got %v", sc.Paths)
	}
}

func TestParseScopeEntries_NonSlicePaths(t *testing.T) {
	// Scope entry with paths that is not a slice
	scopeMap := map[string]any{
		"bad-paths": map[string]any{
			"paths": "not-a-slice",
		},
	}

	scopes := parseScopeEntries(scopeMap)

	sc, ok := scopes["bad-paths"]
	if !ok {
		t.Fatal("expected scope 'bad-paths' to exist")
	}
	if sc.Paths != nil {
		t.Errorf("expected nil paths for non-slice paths, got %v", sc.Paths)
	}
}
