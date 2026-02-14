package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func sampleTasks() []*model.Task {
	return []*model.Task{
		{ID: "001", Title: "Task A", Status: model.StatusPending},
		{ID: "002", Title: "Task B", Status: model.StatusInProgress},
		{ID: "003", Title: "Task C", Status: model.StatusCompleted},
		{ID: "004", Title: "Task D", Status: model.StatusBlocked},
	}
}

func TestNew(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp/tasks", tasks)

	if app.scanDir != "/tmp/tasks" {
		t.Errorf("expected scanDir /tmp/tasks, got %s", app.scanDir)
	}
	if len(app.tasks) != 4 {
		t.Errorf("expected 4 tasks, got %d", len(app.tasks))
	}
	if app.ready {
		t.Error("expected ready to be false initially")
	}
}

func TestUpdate_WindowResize(t *testing.T) {
	app := New("/tmp", nil)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, cmd := app.Update(msg)
	m := updated.(App)

	if cmd != nil {
		t.Error("expected no command from resize")
	}
	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
	if !m.ready {
		t.Error("expected ready to be true after resize")
	}
}

func TestUpdate_QuitKeys(t *testing.T) {
	app := New("/tmp", nil)

	tests := []struct {
		key  string
		quit bool
	}{
		{"q", true},
		{"ctrl+c", true},
		{"a", false},
		{"?", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			_, cmd := app.Update(msg)
			isQuit := cmd != nil
			if isQuit != tt.quit {
				t.Errorf("key %q: expected quit=%v, got %v", tt.key, tt.quit, isQuit)
			}
		})
	}
}

func TestUpdate_HelpToggle(t *testing.T) {
	app := New("/tmp", nil)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updated, _ := app.Update(msg)
	m := updated.(App)
	if !m.showHelp {
		t.Error("expected showHelp to be true after first ?")
	}

	updated, _ = m.Update(msg)
	m = updated.(App)
	if m.showHelp {
		t.Error("expected showHelp to be false after second ?")
	}
}

func TestView_BeforeReady(t *testing.T) {
	app := New("/tmp", nil)

	view := app.View()
	if view != "Initializing..." {
		t.Errorf("expected 'Initializing...', got %q", view)
	}
}

func TestView_AfterReady(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp/tasks", tasks)

	// Simulate resize to set ready
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := updated.(App)

	view := m.View()

	// Header should contain app name and dir
	if !strings.Contains(view, "taskmd") {
		t.Error("expected view to contain 'taskmd'")
	}
	if !strings.Contains(view, "/tmp/tasks") {
		t.Error("expected view to contain scan dir")
	}

	// Content should show task list with summary
	if !strings.Contains(view, "4 tasks:") {
		t.Error("expected view to contain '4 tasks:'")
	}
	if !strings.Contains(view, "1 pending") {
		t.Error("expected view to contain pending count")
	}
	if !strings.Contains(view, "1 in-progress") {
		t.Error("expected view to contain in-progress count")
	}

	// Should show task IDs
	if !strings.Contains(view, "001") {
		t.Error("expected view to contain task ID 001")
	}

	// Footer should show key bindings
	if !strings.Contains(view, "q: quit") {
		t.Error("expected view to contain quit hint")
	}
	if !strings.Contains(view, "navigate") {
		t.Error("expected view to contain navigate hint")
	}
}

func TestView_WithHelp(t *testing.T) {
	app := New("/tmp", sampleTasks())

	// Resize then toggle help
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := updated.(App)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = updated.(App)

	view := m.View()

	if !strings.Contains(view, "Key Bindings:") {
		t.Error("expected help section in view")
	}
	if !strings.Contains(view, "Toggle help") {
		t.Error("expected help content about toggle")
	}
}

func TestView_EmptyTasks(t *testing.T) {
	app := New("/tmp", nil)

	updated, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := updated.(App)

	view := m.View()
	if !strings.Contains(view, "No tasks found") {
		t.Error("expected 'No tasks found' for empty task list")
	}
}

func TestNavigation_Down(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp", tasks)

	if app.selectedIndex != 0 {
		t.Errorf("expected initial selectedIndex to be 0, got %d", app.selectedIndex)
	}

	// Press 'j' to move down
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m := updated.(App)
	if m.selectedIndex != 1 {
		t.Errorf("expected selectedIndex to be 1 after j, got %d", m.selectedIndex)
	}

	// Press down arrow
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(App)
	if m.selectedIndex != 2 {
		t.Errorf("expected selectedIndex to be 2 after down, got %d", m.selectedIndex)
	}

	// Try to move past end
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(App)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(App)
	if m.selectedIndex != 3 {
		t.Errorf("expected selectedIndex to stay at 3 (max), got %d", m.selectedIndex)
	}
}

func TestNavigation_Up(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp", tasks)
	app.selectedIndex = 2

	// Press 'k' to move up
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m := updated.(App)
	if m.selectedIndex != 1 {
		t.Errorf("expected selectedIndex to be 1 after k, got %d", m.selectedIndex)
	}

	// Press up arrow
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(App)
	if m.selectedIndex != 0 {
		t.Errorf("expected selectedIndex to be 0 after up, got %d", m.selectedIndex)
	}

	// Try to move before start
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(App)
	if m.selectedIndex != 0 {
		t.Errorf("expected selectedIndex to stay at 0 (min), got %d", m.selectedIndex)
	}
}

func TestNavigation_GotoTop(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp", tasks)
	app.selectedIndex = 3

	// Press 'g' to go to top
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m := updated.(App)
	if m.selectedIndex != 0 {
		t.Errorf("expected selectedIndex to be 0 after g, got %d", m.selectedIndex)
	}
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to be 0 after g, got %d", m.scrollOffset)
	}
}

func TestNavigation_GotoBottom(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp", tasks)

	// Press 'G' to go to bottom
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	m := updated.(App)
	expectedIndex := len(tasks) - 1
	if m.selectedIndex != expectedIndex {
		t.Errorf("expected selectedIndex to be %d after G, got %d", expectedIndex, m.selectedIndex)
	}
}

func TestSearch_EnterSearchMode(t *testing.T) {
	app := New("/tmp", sampleTasks())

	// Press '/' to enter search mode
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m := updated.(App)
	if !m.searchMode {
		t.Error("expected searchMode to be true after pressing /")
	}
	if m.searchQuery != "" {
		t.Errorf("expected searchQuery to be empty, got %s", m.searchQuery)
	}
}

func TestSearch_TypeQuery(t *testing.T) {
	app := New("/tmp", sampleTasks())
	app.searchMode = true

	// Type "Task A"
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("T")})
	m := updated.(App)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updated.(App)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = updated.(App)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(App)

	if m.searchQuery != "Task" {
		t.Errorf("expected searchQuery to be 'Task', got %s", m.searchQuery)
	}
}

func TestSearch_Backspace(t *testing.T) {
	app := New("/tmp", sampleTasks())
	app.searchMode = true
	app.searchQuery = "Test"

	// Press backspace
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m := updated.(App)

	if m.searchQuery != "Tes" {
		t.Errorf("expected searchQuery to be 'Tes', got %s", m.searchQuery)
	}
}

func TestSearch_CancelWithEscape(t *testing.T) {
	app := New("/tmp", sampleTasks())
	app.searchMode = true
	app.searchQuery = "Test"

	// Press Escape to cancel
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m := updated.(App)

	if m.searchMode {
		t.Error("expected searchMode to be false after Escape")
	}
	if m.searchQuery != "" {
		t.Errorf("expected searchQuery to be cleared, got %s", m.searchQuery)
	}
}

func TestSearch_ApplyWithEnter(t *testing.T) {
	app := New("/tmp", sampleTasks())
	app.searchMode = true
	app.searchQuery = "Test"

	// Press Enter to apply
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(App)

	if m.searchMode {
		t.Error("expected searchMode to be false after Enter")
	}
	if m.searchQuery != "Test" {
		t.Errorf("expected searchQuery to be preserved, got %s", m.searchQuery)
	}
}

func TestGetFilteredTasks_NoQuery(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp", tasks)

	filtered := app.getFilteredTasks()
	if len(filtered) != len(tasks) {
		t.Errorf("expected %d tasks, got %d", len(tasks), len(filtered))
	}
}

func TestGetFilteredTasks_ByTitle(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp", tasks)
	app.searchQuery = "Task A"

	filtered := app.getFilteredTasks()
	if len(filtered) != 1 {
		t.Errorf("expected 1 task, got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].ID != "001" {
		t.Errorf("expected task 001, got %s", filtered[0].ID)
	}
}

func TestGetFilteredTasks_ByID(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp", tasks)
	app.searchQuery = "002"

	filtered := app.getFilteredTasks()
	if len(filtered) != 1 {
		t.Errorf("expected 1 task, got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].ID != "002" {
		t.Errorf("expected task 002, got %s", filtered[0].ID)
	}
}

func TestGetFilteredTasks_CaseInsensitive(t *testing.T) {
	tasks := sampleTasks()
	app := New("/tmp", tasks)
	app.searchQuery = "task a"

	filtered := app.getFilteredTasks()
	if len(filtered) != 1 {
		t.Errorf("expected 1 task (case insensitive), got %d", len(filtered))
	}
}

func TestNewWithOptions_FocusTask(t *testing.T) {
	tasks := sampleTasks()
	opts := Options{FocusTaskID: "003"}
	app := NewWithOptions("/tmp", tasks, opts)

	if app.selectedIndex != 2 {
		t.Errorf("expected selectedIndex to be 2 for task 003, got %d", app.selectedIndex)
	}
}

func TestNewWithOptions_Filter(t *testing.T) {
	tasks := sampleTasks()
	opts := Options{Filter: "status=pending"}
	app := NewWithOptions("/tmp", tasks, opts)

	if app.searchQuery != "pending" {
		t.Errorf("expected searchQuery to be 'pending', got %s", app.searchQuery)
	}
}

func TestNewWithOptions_ReadOnly(t *testing.T) {
	tasks := sampleTasks()
	opts := Options{ReadOnly: true}
	app := NewWithOptions("/tmp", tasks, opts)

	if !app.readonly {
		t.Error("expected readonly to be true")
	}
}

func TestNewWithOptions_GroupBy(t *testing.T) {
	tasks := sampleTasks()
	opts := Options{GroupBy: "priority"}
	app := NewWithOptions("/tmp", tasks, opts)

	if app.groupBy != "priority" {
		t.Errorf("expected groupBy to be 'priority', got %s", app.groupBy)
	}
}

func TestParseFilter_Simple(t *testing.T) {
	result := parseFilter("status=pending")
	if result != "pending" {
		t.Errorf("expected 'pending', got %s", result)
	}
}

func TestParseFilter_NoEquals(t *testing.T) {
	result := parseFilter("pending")
	if result != "pending" {
		t.Errorf("expected 'pending', got %s", result)
	}
}

func sampleTasksWithBody() []*model.Task {
	return []*model.Task{
		{
			ID:           "001",
			Title:        "Task A",
			Status:       model.StatusPending,
			Priority:     model.PriorityHigh,
			Tags:         []string{"cli", "feature"},
			Dependencies: []string{"000"},
			Body:         "This is the **body** of task A.",
			FilePath:     "/tmp/tasks/001.md",
		},
		{
			ID:       "002",
			Title:    "Task B",
			Status:   model.StatusInProgress,
			Body:     "Body of task B.",
			FilePath: "/tmp/tasks/002.md",
		},
	}
}

func initApp(tasks []*model.Task) App {
	app := New("/tmp/tasks", tasks)
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return updated.(App)
}

func pressKey(m App, key string) App {
	var msg tea.KeyMsg
	switch key {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "escape":
		msg = tea.KeyMsg{Type: tea.KeyEscape}
	case "backspace":
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	case "ctrl+c":
		msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
	updated, cmd := m.Update(msg)
	app := updated.(App)
	// If the update returned a command (e.g. async detail rendering), execute it
	// and feed the resulting message back into Update to complete the cycle.
	if cmd != nil {
		if result := cmd(); result != nil {
			updated, _ = app.Update(result)
			app = updated.(App)
		}
	}
	return app
}

func TestDetailView_EnterOpensDetail(t *testing.T) {
	m := initApp(sampleTasksWithBody())

	if m.viewMode != viewList {
		t.Errorf("expected viewList, got %d", m.viewMode)
	}

	m = pressKey(m, "enter")

	if m.viewMode != viewDetail {
		t.Errorf("expected viewDetail after Enter, got %d", m.viewMode)
	}
	if m.detailScrollOffset != 0 {
		t.Errorf("expected detailScrollOffset 0, got %d", m.detailScrollOffset)
	}
	if m.renderedDetail == "" {
		t.Error("expected renderedDetail to be populated")
	}
}

func TestDetailView_EscReturnsToList(t *testing.T) {
	m := initApp(sampleTasksWithBody())
	m = pressKey(m, "enter")

	if m.viewMode != viewDetail {
		t.Fatal("expected to be in detail view")
	}

	m = pressKey(m, "escape")

	if m.viewMode != viewList {
		t.Errorf("expected viewList after Esc, got %d", m.viewMode)
	}
}

func TestDetailView_BackspaceReturnsToList(t *testing.T) {
	m := initApp(sampleTasksWithBody())
	m = pressKey(m, "enter")

	if m.viewMode != viewDetail {
		t.Fatal("expected to be in detail view")
	}

	m = pressKey(m, "backspace")

	if m.viewMode != viewList {
		t.Errorf("expected viewList after Backspace, got %d", m.viewMode)
	}
}

func TestDetailView_ScrollDown(t *testing.T) {
	m := initApp(sampleTasksWithBody())
	m = pressKey(m, "enter")

	initial := m.detailScrollOffset
	m = pressKey(m, "j")

	if m.detailScrollOffset != initial+1 {
		t.Errorf("expected detailScrollOffset %d, got %d", initial+1, m.detailScrollOffset)
	}
}

func TestDetailView_ScrollUp(t *testing.T) {
	m := initApp(sampleTasksWithBody())
	m = pressKey(m, "enter")

	// Scroll down first
	m = pressKey(m, "j")
	m = pressKey(m, "j")

	m = pressKey(m, "k")
	if m.detailScrollOffset != 1 {
		t.Errorf("expected detailScrollOffset 1, got %d", m.detailScrollOffset)
	}

	// Scroll up past 0 should clamp
	m = pressKey(m, "k")
	m = pressKey(m, "k")
	if m.detailScrollOffset != 0 {
		t.Errorf("expected detailScrollOffset 0 (clamped), got %d", m.detailScrollOffset)
	}
}

func TestDetailView_QuitFromDetail(t *testing.T) {
	m := initApp(sampleTasksWithBody())
	m = pressKey(m, "enter")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("expected quit command from detail view")
	}
}

func TestDetailView_RenderContainsMetadata(t *testing.T) {
	m := initApp(sampleTasksWithBody())
	m = pressKey(m, "enter")

	view := m.View()

	checks := []struct {
		label string
		want  string
	}{
		{"task ID", "001"},
		{"task title", "Task A"},
		{"status", "pending"},
		{"file path", "/tmp/tasks/001.md"},
	}

	for _, c := range checks {
		if !strings.Contains(view, c.want) {
			t.Errorf("expected detail view to contain %s (%q)", c.label, c.want)
		}
	}
}

func TestDetailView_RenderContainsBody(t *testing.T) {
	m := initApp(sampleTasksWithBody())
	m = pressKey(m, "enter")

	view := m.View()

	// The body contains "body" as a word — it should appear in the rendered output
	if !strings.Contains(view, "body") {
		t.Error("expected detail view to contain body content")
	}
}

func TestDetailView_FooterShowsDetailHints(t *testing.T) {
	m := initApp(sampleTasksWithBody())
	m = pressKey(m, "enter")

	view := m.View()

	if !strings.Contains(view, "Esc: back") {
		t.Error("expected footer to contain 'Esc: back'")
	}
	if !strings.Contains(view, "scroll") {
		t.Error("expected footer to contain 'scroll'")
	}
}

func TestDetailView_PreservesListSelection(t *testing.T) {
	m := initApp(sampleTasksWithBody())

	// Move to second task
	m = pressKey(m, "j")
	if m.selectedIndex != 1 {
		t.Fatalf("expected selectedIndex 1, got %d", m.selectedIndex)
	}

	// Enter detail view and return
	m = pressKey(m, "enter")
	m = pressKey(m, "escape")

	if m.selectedIndex != 1 {
		t.Errorf("expected selectedIndex preserved at 1, got %d", m.selectedIndex)
	}
}

func TestDetailView_EnterWithNoTasks(t *testing.T) {
	m := initApp(nil)
	m = pressKey(m, "enter")

	if m.viewMode != viewList {
		t.Error("expected viewMode to stay viewList when no tasks")
	}
}

// Live Reload Tests

func TestFileChangeMsg_TriggersRefresh(t *testing.T) {
	m := initApp(sampleTasks())

	// Send file change message
	msg := fileChangeMsg{}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("expected refresh command after file change")
	}
}

func TestTasksRefreshedMsg_UpdatesTasks(t *testing.T) {
	m := initApp(sampleTasks())

	// Original tasks
	if len(m.tasks) != 4 {
		t.Fatalf("expected 4 initial tasks, got %d", len(m.tasks))
	}

	// Send refresh with new tasks
	newTasks := []*model.Task{
		{ID: "001", Title: "Updated Task A", Status: model.StatusPending},
		{ID: "005", Title: "New Task E", Status: model.StatusPending},
	}
	msg := tasksRefreshedMsg{tasks: newTasks}
	updated, cmd := m.Update(msg)
	m = updated.(App)

	if len(m.tasks) != 2 {
		t.Errorf("expected 2 tasks after refresh, got %d", len(m.tasks))
	}
	if m.tasks[0].Title != "Updated Task A" {
		t.Errorf("expected updated title, got %s", m.tasks[0].Title)
	}
	if cmd == nil {
		t.Error("expected command to continue listening for file changes")
	}
}

func TestTasksRefreshedMsg_ShowsIndicator(t *testing.T) {
	m := initApp(sampleTasks())

	msg := tasksRefreshedMsg{tasks: sampleTasks()}
	updated, _ := m.Update(msg)
	m = updated.(App)

	if !m.showRefreshIndicator {
		t.Error("expected showRefreshIndicator to be true after refresh")
	}
}

func TestHideRefreshIndicatorMsg_HidesIndicator(t *testing.T) {
	m := initApp(sampleTasks())
	m.showRefreshIndicator = true

	msg := hideRefreshIndicatorMsg{}
	updated, _ := m.Update(msg)
	m = updated.(App)

	if m.showRefreshIndicator {
		t.Error("expected showRefreshIndicator to be false after hide message")
	}
}

func TestUpdateTasksPreservingState_PreservesSelection(t *testing.T) {
	m := initApp(sampleTasks())
	m.selectedIndex = 2 // Select task 003

	// Refresh with same tasks (simulating file edit)
	newTasks := []*model.Task{
		{ID: "001", Title: "Task A Updated", Status: model.StatusPending},
		{ID: "002", Title: "Task B Updated", Status: model.StatusInProgress},
		{ID: "003", Title: "Task C Updated", Status: model.StatusCompleted},
		{ID: "004", Title: "Task D Updated", Status: model.StatusBlocked},
	}

	m.updateTasksPreservingState(newTasks)

	if m.selectedIndex != 2 {
		t.Errorf("expected selectedIndex to be preserved at 2, got %d", m.selectedIndex)
	}
	if m.tasks[2].ID != "003" {
		t.Errorf("expected task 003 at index 2, got %s", m.tasks[2].ID)
	}
	if m.tasks[2].Title != "Task C Updated" {
		t.Errorf("expected updated title, got %s", m.tasks[2].Title)
	}
}

func TestUpdateTasksPreservingState_TaskRemoved(t *testing.T) {
	m := initApp(sampleTasks())
	m.selectedIndex = 2 // Select task 003

	// Refresh with task 003 removed
	newTasks := []*model.Task{
		{ID: "001", Title: "Task A", Status: model.StatusPending},
		{ID: "002", Title: "Task B", Status: model.StatusInProgress},
		{ID: "004", Title: "Task D", Status: model.StatusBlocked},
	}

	m.updateTasksPreservingState(newTasks)

	// Selection should be clamped to valid range
	if m.selectedIndex >= len(newTasks) {
		t.Errorf("expected selectedIndex to be clamped to <%d, got %d", len(newTasks), m.selectedIndex)
	}
	if m.selectedIndex < 0 {
		t.Errorf("expected selectedIndex to be non-negative, got %d", m.selectedIndex)
	}
}

func TestUpdateTasksPreservingState_TaskAddedBefore(t *testing.T) {
	m := initApp(sampleTasks())
	m.selectedIndex = 1 // Select task 002

	// Add a new task at the beginning
	newTasks := []*model.Task{
		{ID: "000", Title: "New Task", Status: model.StatusPending},
		{ID: "001", Title: "Task A", Status: model.StatusPending},
		{ID: "002", Title: "Task B", Status: model.StatusInProgress},
		{ID: "003", Title: "Task C", Status: model.StatusCompleted},
		{ID: "004", Title: "Task D", Status: model.StatusBlocked},
	}

	m.updateTasksPreservingState(newTasks)

	// Should still select task 002 (now at index 2)
	if m.tasks[m.selectedIndex].ID != "002" {
		t.Errorf("expected task 002 to still be selected, got %s", m.tasks[m.selectedIndex].ID)
	}
}

func TestUpdateTasksPreservingState_WithFilter(t *testing.T) {
	m := initApp(sampleTasks())
	m.searchQuery = "Task C"
	filteredTasks := m.getFilteredTasks()
	if len(filteredTasks) != 1 {
		t.Fatalf("expected 1 filtered task, got %d", len(filteredTasks))
	}
	m.selectedIndex = 0 // Select the only filtered task (003)

	// Refresh with task 003 updated
	newTasks := []*model.Task{
		{ID: "001", Title: "Task A", Status: model.StatusPending},
		{ID: "002", Title: "Task B", Status: model.StatusInProgress},
		{ID: "003", Title: "Task C Modified", Status: model.StatusCompleted},
		{ID: "004", Title: "Task D", Status: model.StatusBlocked},
	}

	m.updateTasksPreservingState(newTasks)

	// Should still select task 003 in filtered view
	filteredTasks = m.getFilteredTasks()
	if len(filteredTasks) == 0 {
		t.Fatal("expected at least one filtered task")
	}
	// Check if selection is valid and points to the right task
	if m.selectedIndex >= len(filteredTasks) {
		t.Fatalf("selectedIndex %d out of range for %d filtered tasks", m.selectedIndex, len(filteredTasks))
	}
	if filteredTasks[m.selectedIndex].ID != "003" {
		t.Errorf("expected task 003 to be selected, got %s", filteredTasks[m.selectedIndex].ID)
	}
}

func TestRenderHeader_ShowsRefreshIndicator(t *testing.T) {
	m := initApp(sampleTasks())
	m.showRefreshIndicator = true

	header := m.renderHeader()

	// Check for refresh indicator (⟳)
	if !strings.Contains(header, "⟳") && !strings.Contains(header, "refresh") {
		t.Error("expected header to contain refresh indicator when showRefreshIndicator is true")
	}
}

func TestRenderHeader_HidesRefreshIndicator(t *testing.T) {
	m := initApp(sampleTasks())
	m.showRefreshIndicator = false

	header := m.renderHeader()

	// Should not contain refresh indicator
	if strings.Contains(header, "⟳") {
		t.Error("expected header to not contain refresh indicator when showRefreshIndicator is false")
	}
}
