package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/sync"
)

func TestSyncCommand_MissingConfig(t *testing.T) {
	syncDryRun = false
	syncSource = ""

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	err := runSync(syncCmd, nil)
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
	if !strings.Contains(err.Error(), "sync config") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSyncCommand_DryRun(t *testing.T) {
	sourceName := "test-cli-dryrun"
	defer sync.Unregister(sourceName)

	sync.Register(&cliMockSource{
		name: sourceName,
		tasks: []sync.ExternalTask{
			{ExternalID: "CLI-1", Title: "CLI test task", Status: "open"},
		},
	})

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	configContent := "sources:\n  - name: " + sourceName + "\n    output_dir: \"tasks\"\n    field_map:\n      status:\n        open: pending\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".taskmd-sync.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	syncDryRun = true
	syncSource = ""

	err := runSync(syncCmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No files should be created
	_, statErr := os.Stat(filepath.Join(tmpDir, "tasks"))
	if statErr == nil {
		t.Error("expected no tasks directory in dry-run mode")
	}
}

func TestSyncCommand_SourceFilter(t *testing.T) {
	syncDryRun = false
	syncSource = "does-not-exist"

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	configContent := "sources:\n  - name: some-source\n    output_dir: \"tasks\"\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".taskmd-sync.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := runSync(syncCmd, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent source filter")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSyncCommand_FullSync(t *testing.T) {
	sourceName := "test-cli-fullsync"
	defer sync.Unregister(sourceName)

	sync.Register(&cliMockSource{
		name: sourceName,
		tasks: []sync.ExternalTask{
			{ExternalID: "FS-1", Title: "Full sync task", Status: "open"},
		},
	})

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	configContent := "sources:\n  - name: " + sourceName + "\n    output_dir: \"tasks\"\n    field_map:\n      status:\n        open: pending\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".taskmd-sync.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	syncDryRun = false
	syncSource = ""

	err := runSync(syncCmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify task file was created
	entries, err := os.ReadDir(filepath.Join(tmpDir, "tasks"))
	if err != nil {
		t.Fatalf("failed to read tasks dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 task file, got %d", len(entries))
	}
}

func TestSyncCommand_SourceFilterMatch(t *testing.T) {
	sourceName := "test-cli-match"
	defer sync.Unregister(sourceName)

	sync.Register(&cliMockSource{
		name: sourceName,
		tasks: []sync.ExternalTask{
			{ExternalID: "M-1", Title: "Matched task", Status: "open"},
		},
	})

	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	configContent := "sources:\n  - name: " + sourceName + "\n    output_dir: \"tasks\"\n    field_map:\n      status:\n        open: pending\n"
	if err := os.WriteFile(filepath.Join(tmpDir, ".taskmd-sync.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	syncDryRun = false
	syncSource = sourceName

	err := runSync(syncCmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncCommand_FlagDefaults(t *testing.T) {
	syncDryRun = false
	syncSource = ""

	if syncDryRun {
		t.Error("syncDryRun should be false by default")
	}
	if syncSource != "" {
		t.Error("syncSource should be empty by default")
	}
}

// cliMockSource implements sync.Source for CLI integration tests.
type cliMockSource struct {
	name  string
	tasks []sync.ExternalTask
}

func (m *cliMockSource) Name() string                             { return m.name }
func (m *cliMockSource) ValidateConfig(_ sync.SourceConfig) error { return nil }
func (m *cliMockSource) FetchTasks(_ sync.SourceConfig) ([]sync.ExternalTask, error) {
	return m.tasks, nil
}
