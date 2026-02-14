package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mockSource implements the Source interface for testing.
type mockSource struct {
	name      string
	tasks     []ExternalTask
	fetchErr  error
	configErr error
}

func (m *mockSource) Name() string                        { return m.name }
func (m *mockSource) ValidateConfig(_ SourceConfig) error { return m.configErr }
func (m *mockSource) FetchTasks(_ SourceConfig) ([]ExternalTask, error) {
	return m.tasks, m.fetchErr
}

func setupMockSource(name string, tasks []ExternalTask) {
	Register(&mockSource{name: name, tasks: tasks})
}

func cleanupRegistry(name string) {
	Unregister(name)
}

func TestEngine_CreateNewTasks(t *testing.T) {
	sourceName := "test-create"
	defer cleanupRegistry(sourceName)

	setupMockSource(sourceName, []ExternalTask{
		{
			ExternalID:  "EXT-1",
			Title:       "First task",
			Description: "Task one description",
			Status:      "open",
			Priority:    "p1",
			Labels:      []string{"bug"},
		},
		{
			ExternalID:  "EXT-2",
			Title:       "Second task",
			Description: "Task two description",
			Status:      "open",
		},
	})

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "tasks")

	engine := &Engine{ConfigDir: dir, Verbose: false}
	srcCfg := SourceConfig{
		Name:      sourceName,
		OutputDir: outputDir,
		FieldMap: FieldMap{
			Status:       map[string]string{"open": "pending"},
			Priority:     map[string]string{"p1": "high"},
			LabelsToTags: true,
		},
	}

	result, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Created) != 2 {
		t.Fatalf("expected 2 created, got %d", len(result.Created))
	}
	if len(result.Updated) != 0 {
		t.Errorf("expected 0 updated, got %d", len(result.Updated))
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(result.Conflicts))
	}

	// Verify files were created
	for _, a := range result.Created {
		if _, err := os.Stat(a.FilePath); os.IsNotExist(err) {
			t.Errorf("expected file to exist: %s", a.FilePath)
		}
	}

	// Verify state was saved
	state, err := LoadState(dir, sourceName)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if len(state.Tasks) != 2 {
		t.Errorf("expected 2 tasks in state, got %d", len(state.Tasks))
	}
}

func TestEngine_Idempotent(t *testing.T) {
	sourceName := "test-idempotent"
	defer cleanupRegistry(sourceName)

	tasks := []ExternalTask{
		{ExternalID: "EXT-1", Title: "A task", Status: "open"},
	}
	setupMockSource(sourceName, tasks)

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "tasks")

	engine := &Engine{ConfigDir: dir}
	srcCfg := SourceConfig{
		Name:      sourceName,
		OutputDir: outputDir,
		FieldMap:  FieldMap{Status: map[string]string{"open": "pending"}},
	}

	// First sync
	result1, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("first sync error: %v", err)
	}
	if len(result1.Created) != 1 {
		t.Fatalf("expected 1 created, got %d", len(result1.Created))
	}

	// Second sync — same data, should skip
	result2, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("second sync error: %v", err)
	}
	if len(result2.Created) != 0 {
		t.Errorf("expected 0 created on second sync, got %d", len(result2.Created))
	}
	if len(result2.Skipped) != 1 {
		t.Errorf("expected 1 skipped on second sync, got %d", len(result2.Skipped))
	}
}

func TestEngine_UpdateOnExternalChange(t *testing.T) {
	sourceName := "test-update"
	defer cleanupRegistry(sourceName)

	mock := &mockSource{
		name: sourceName,
		tasks: []ExternalTask{
			{ExternalID: "EXT-1", Title: "Original title", Status: "open"},
		},
	}
	Register(mock)

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "tasks")

	engine := &Engine{ConfigDir: dir}
	srcCfg := SourceConfig{
		Name:      sourceName,
		OutputDir: outputDir,
		FieldMap:  FieldMap{Status: map[string]string{"open": "pending", "closed": "completed"}},
	}

	// First sync
	_, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("first sync error: %v", err)
	}

	// Update mock data
	mock.tasks = []ExternalTask{
		{ExternalID: "EXT-1", Title: "Updated title", Status: "closed"},
	}

	// Second sync — should update
	result, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("second sync error: %v", err)
	}
	if len(result.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d", len(result.Updated))
	}
}

func TestEngine_ConflictOnLocalChange(t *testing.T) {
	sourceName := "test-conflict"
	defer cleanupRegistry(sourceName)

	setupMockSource(sourceName, []ExternalTask{
		{ExternalID: "EXT-1", Title: "A task", Status: "open"},
	})

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "tasks")

	engine := &Engine{ConfigDir: dir}
	srcCfg := SourceConfig{
		Name:      sourceName,
		OutputDir: outputDir,
		FieldMap:  FieldMap{Status: map[string]string{"open": "pending"}},
	}

	// First sync creates the file
	result1, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("first sync error: %v", err)
	}
	if len(result1.Created) != 1 {
		t.Fatalf("expected 1 created, got %d", len(result1.Created))
	}

	// Modify the local file
	filePath := result1.Created[0].FilePath
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	modified := strings.Replace(string(data), "pending", "in-progress", 1)
	if err := os.WriteFile(filePath, []byte(modified), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Second sync — should detect conflict
	result2, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("second sync error: %v", err)
	}
	if len(result2.Conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d (created=%d, updated=%d, skipped=%d)",
			len(result2.Conflicts), len(result2.Created), len(result2.Updated), len(result2.Skipped))
	}
}

func TestEngine_ConflictRemote(t *testing.T) {
	sourceName := "test-conflict-remote"
	defer cleanupRegistry(sourceName)

	setupMockSource(sourceName, []ExternalTask{
		{ExternalID: "EXT-1", Title: "A task", Status: "open"},
	})

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "tasks")

	engine := &Engine{ConfigDir: dir, ConflictStrategy: ConflictRemote}
	srcCfg := SourceConfig{
		Name:      sourceName,
		OutputDir: outputDir,
		FieldMap:  FieldMap{Status: map[string]string{"open": "pending"}},
	}

	// First sync creates the file
	result1, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("first sync error: %v", err)
	}
	filePath := result1.Created[0].FilePath

	// Modify the local file
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	modified := strings.Replace(string(data), "pending", "in-progress", 1)
	if err := os.WriteFile(filePath, []byte(modified), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Second sync with ConflictRemote — should overwrite local
	result2, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("second sync error: %v", err)
	}
	if len(result2.Conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(result2.Conflicts))
	}
	if len(result2.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d", len(result2.Updated))
	}

	// Verify file was overwritten with remote content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if strings.Contains(string(content), "in-progress") {
		t.Error("expected local changes to be overwritten")
	}
}

func TestEngine_ConflictLocal(t *testing.T) {
	sourceName := "test-conflict-local"
	defer cleanupRegistry(sourceName)

	setupMockSource(sourceName, []ExternalTask{
		{ExternalID: "EXT-1", Title: "A task", Status: "open"},
	})

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "tasks")

	engine := &Engine{ConfigDir: dir, ConflictStrategy: ConflictLocal}
	srcCfg := SourceConfig{
		Name:      sourceName,
		OutputDir: outputDir,
		FieldMap:  FieldMap{Status: map[string]string{"open": "pending"}},
	}

	// First sync creates the file
	result1, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("first sync error: %v", err)
	}
	filePath := result1.Created[0].FilePath

	// Modify the local file
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	modified := strings.Replace(string(data), "pending", "in-progress", 1)
	if err := os.WriteFile(filePath, []byte(modified), 0644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Second sync with ConflictLocal — should keep local, update state
	result2, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("second sync error: %v", err)
	}
	if len(result2.Conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(result2.Conflicts))
	}
	if len(result2.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d", len(result2.Updated))
	}

	// Verify local changes are preserved
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(content), "in-progress") {
		t.Error("expected local changes to be preserved")
	}

	// Third sync — should skip (state is now up to date)
	result3, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("third sync error: %v", err)
	}
	if len(result3.Skipped) != 1 {
		t.Errorf("expected 1 skipped on third sync, got %d (conflicts=%d, updated=%d, created=%d)",
			len(result3.Skipped), len(result3.Conflicts), len(result3.Updated), len(result3.Created))
	}
}

func TestEngine_DryRun(t *testing.T) {
	sourceName := "test-dryrun"
	defer cleanupRegistry(sourceName)

	setupMockSource(sourceName, []ExternalTask{
		{ExternalID: "EXT-1", Title: "A task", Status: "open"},
	})

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "tasks")

	engine := &Engine{ConfigDir: dir, DryRun: true}
	srcCfg := SourceConfig{
		Name:      sourceName,
		OutputDir: outputDir,
		FieldMap:  FieldMap{Status: map[string]string{"open": "pending"}},
	}

	result, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Created) != 1 {
		t.Errorf("expected 1 created, got %d", len(result.Created))
	}

	// Verify no files were created
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	for _, e := range entries {
		if e.Name() == "tasks" {
			t.Error("expected no tasks directory in dry-run mode")
		}
	}

	// Verify no state was saved
	state, err := LoadState(dir, sourceName)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}
	if len(state.Tasks) != 0 {
		t.Errorf("expected empty state in dry-run mode, got %d tasks", len(state.Tasks))
	}
}

func TestEngine_UnknownSource(t *testing.T) {
	dir := t.TempDir()
	engine := &Engine{ConfigDir: dir}
	srcCfg := SourceConfig{
		Name:      "nonexistent",
		OutputDir: filepath.Join(dir, "tasks"),
	}

	_, err := engine.RunSync(srcCfg)
	if err == nil {
		t.Fatal("expected error for unknown source")
	}
}

func TestEngine_MultipleTasksGetSequentialIDs(t *testing.T) {
	sourceName := "test-sequential"
	defer cleanupRegistry(sourceName)

	setupMockSource(sourceName, []ExternalTask{
		{ExternalID: "EXT-1", Title: "First", Status: "open"},
		{ExternalID: "EXT-2", Title: "Second", Status: "open"},
		{ExternalID: "EXT-3", Title: "Third", Status: "open"},
	})

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "tasks")

	engine := &Engine{ConfigDir: dir}
	srcCfg := SourceConfig{
		Name:      sourceName,
		OutputDir: outputDir,
		FieldMap:  FieldMap{Status: map[string]string{"open": "pending"}},
	}

	result, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Created) != 3 {
		t.Fatalf("expected 3 created, got %d", len(result.Created))
	}

	// IDs should be sequential
	ids := make(map[string]bool)
	for _, a := range result.Created {
		ids[a.LocalID] = true
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 unique IDs, got %d", len(ids))
	}
}

func TestHashExternalTask_Deterministic(t *testing.T) {
	ext := ExternalTask{
		ExternalID:  "1",
		Title:       "Test",
		Description: "Desc",
		Status:      "open",
		Priority:    "high",
		Assignee:    "alice",
		Labels:      []string{"bug"},
		URL:         "https://example.com",
	}

	h1 := HashExternalTask(ext)
	h2 := HashExternalTask(ext)

	if h1 != h2 {
		t.Errorf("hash not deterministic: %s != %s", h1, h2)
	}

	// Different task should have different hash
	ext.Title = "Different"
	h3 := HashExternalTask(ext)
	if h1 == h3 {
		t.Error("expected different hash for different task")
	}
}

func TestHashExternalTask_SensitiveToAllFields(t *testing.T) {
	base := ExternalTask{
		ExternalID: "1", Title: "T", Description: "D",
		Status: "open", Priority: "p1", Assignee: "a",
		Labels: []string{"l"}, URL: "u",
	}
	baseHash := HashExternalTask(base)

	fields := []struct {
		name   string
		modify func(ExternalTask) ExternalTask
	}{
		{"ExternalID", func(e ExternalTask) ExternalTask { e.ExternalID = "2"; return e }},
		{"Title", func(e ExternalTask) ExternalTask { e.Title = "X"; return e }},
		{"Description", func(e ExternalTask) ExternalTask { e.Description = "X"; return e }},
		{"Status", func(e ExternalTask) ExternalTask { e.Status = "closed"; return e }},
		{"Priority", func(e ExternalTask) ExternalTask { e.Priority = "p2"; return e }},
		{"Assignee", func(e ExternalTask) ExternalTask { e.Assignee = "b"; return e }},
		{"Labels", func(e ExternalTask) ExternalTask { e.Labels = []string{"x"}; return e }},
		{"URL", func(e ExternalTask) ExternalTask { e.URL = "x"; return e }},
	}

	for _, f := range fields {
		modified := f.modify(base)
		h := HashExternalTask(modified)
		if h == baseHash {
			t.Errorf("hash unchanged after modifying %s", f.name)
		}
	}
}

func TestHashLocalFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	writeFile(t, path, "hello world")

	h1, err := HashLocalFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	h2, err := HashLocalFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if h1 != h2 {
		t.Error("hash not deterministic")
	}

	// Different content → different hash
	writeFile(t, path, "different content")
	h3, err := HashLocalFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 == h3 {
		t.Error("expected different hash for different content")
	}
}

func TestHashLocalFile_NonexistentFile(t *testing.T) {
	_, err := HashLocalFile("/nonexistent/path.md")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

// Verify timestamps are set in state
func TestEngine_StateTimestamps(t *testing.T) {
	sourceName := "test-timestamps"
	defer cleanupRegistry(sourceName)

	setupMockSource(sourceName, []ExternalTask{
		{ExternalID: "EXT-1", Title: "A task", Status: "open"},
	})

	dir := t.TempDir()
	outputDir := filepath.Join(dir, "tasks")

	before := time.Now().Add(-time.Second)

	engine := &Engine{ConfigDir: dir}
	srcCfg := SourceConfig{
		Name:      sourceName,
		OutputDir: outputDir,
		FieldMap:  FieldMap{Status: map[string]string{"open": "pending"}},
	}

	_, err := engine.RunSync(srcCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, err := LoadState(dir, sourceName)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if state.LastSync.Before(before) {
		t.Error("expected LastSync to be recent")
	}

	ts := state.Tasks["EXT-1"]
	if ts.LastSynced.Before(before) {
		t.Error("expected task LastSynced to be recent")
	}
}
