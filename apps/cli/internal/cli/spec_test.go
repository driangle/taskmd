package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSpecCommand_WritesToDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("spec")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	content, err := os.ReadFile(repo.Path(specFilename))
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !bytes.Equal(content, specTemplate) {
		t.Error("written content does not match embedded spec")
	}
}

func TestSpecCommand_RefusesOverwriteWithoutForce(t *testing.T) {
	repo := newTaskRepo(t, nil)
	existingPath := repo.Write(specFilename, "existing content")

	res := repo.Run("spec")
	if res.Err == nil {
		t.Fatal("expected error when file already exists")
	}

	if !bytes.Contains([]byte(res.Err.Error()), []byte("already exists")) {
		t.Errorf("error message %q should contain 'already exists'", res.Err.Error())
	}

	// Verify original content was not overwritten
	content, _ := os.ReadFile(existingPath)
	if string(content) != "existing content" {
		t.Error("existing file should not have been overwritten")
	}
}

func TestSpecCommand_OverwritesWithForce(t *testing.T) {
	repo := newTaskRepo(t, nil)
	existingPath := repo.Write(specFilename, "old content")

	res := repo.Run("spec", "--force")
	if res.Err != nil {
		t.Fatalf("unexpected error with --force: %v", res.Err)
	}

	content, err := os.ReadFile(existingPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !bytes.Equal(content, specTemplate) {
		t.Error("file should have been overwritten with spec content")
	}
}

func TestSpecCommand_StdoutPrintsWithoutCreatingFile(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("spec", "--stdout")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if res.Stdout != string(specTemplate) {
		t.Error("stdout output does not match embedded spec")
	}

	// Verify no file was created
	if _, err := os.Stat(repo.Path(specFilename)); err == nil {
		t.Errorf("%s should not have been created with --stdout", specFilename)
	}
}

func TestSpecCommand_DirWritesToSpecifiedDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)
	subDir := repo.Path("docs")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	res := repo.Run("spec", "--dir", subDir)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	content, err := os.ReadFile(filepath.Join(subDir, specFilename))
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !bytes.Equal(content, specTemplate) {
		t.Error("written content does not match embedded spec")
	}
}

func TestSpecCommand_NonExistentDirectoryReturnsError(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("spec", "--dir", "/nonexistent/path/that/does/not/exist")
	if res.Err == nil {
		t.Fatal("expected error for non-existent directory")
	}

	if !bytes.Contains([]byte(res.Err.Error()), []byte("directory does not exist")) {
		t.Errorf("error message %q should contain 'directory does not exist'", res.Err.Error())
	}
}

func TestSpecCommand_ContentMatchesTemplate(t *testing.T) {
	if len(specTemplate) == 0 {
		t.Fatal("embedded spec should not be empty")
	}

	if !bytes.HasPrefix(specTemplate, []byte("# taskmd Specification")) {
		t.Error("spec should start with expected header")
	}
}

func TestSpecTemplate_MatchesCanonicalSpec(t *testing.T) {
	// docs/taskmd_specification.md is the single source of truth.
	// Two copies exist for technical reasons:
	//   - apps/cli/internal/cli/templates/TASKMD_SPEC.md (go:embed requires module-local file)
	//   - apps/docs/reference/specification.md (VitePress requires file in docs tree)
	// Run `make sync-spec` from apps/cli/ to fix drift.
	repoRoot := filepath.Join("..", "..", "..", "..")
	canonicalPath := filepath.Join(repoRoot, "docs", "taskmd_specification.md")
	canonical, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Skipf("skipping: canonical spec not found at %s", canonicalPath)
	}

	if !bytes.Equal(specTemplate, canonical) {
		t.Error("embedded spec template has drifted from docs/taskmd_specification.md.\n" +
			"Run `make sync-spec` from apps/cli/ to fix.")
	}

	docsPath := filepath.Join(repoRoot, "apps", "docs", "reference", "specification.md")
	docsSite, err := os.ReadFile(docsPath)
	if err != nil {
		t.Skipf("skipping: docs site spec not found at %s", docsPath)
	}

	if !bytes.Equal(docsSite, canonical) {
		t.Error("apps/docs/reference/specification.md has drifted from docs/taskmd_specification.md.\n" +
			"Run `make sync-spec` from apps/cli/ to fix.")
	}

	// Also verify the operations spec is synced (linked from specification.md).
	// The sync rewrites ./taskmd_specification.md → ./specification.md for VitePress.
	canonicalOps := filepath.Join(repoRoot, "docs", "taskmd_operations.md")
	docsOps := filepath.Join(repoRoot, "apps", "docs", "reference", "taskmd_operations.md")
	canonicalOpsContent, err := os.ReadFile(canonicalOps)
	if err != nil {
		t.Skipf("skipping: canonical operations spec not found at %s", canonicalOps)
	}
	expectedOps := bytes.ReplaceAll(canonicalOpsContent,
		[]byte("./taskmd_specification.md"), []byte("./specification.md"))
	docsOpsContent, err := os.ReadFile(docsOps)
	if err != nil {
		t.Errorf("apps/docs/reference/taskmd_operations.md is missing.\n"+
			"Run `make sync-spec` from apps/cli/ to fix. (source: %s)", canonicalOps)
	} else if !bytes.Equal(docsOpsContent, expectedOps) {
		t.Error("apps/docs/reference/taskmd_operations.md has drifted from docs/taskmd_operations.md.\n" +
			"Run `make sync-spec` from apps/cli/ to fix.")
	}
}
