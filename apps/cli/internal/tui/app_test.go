package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
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

	// Content should show task counts
	if !strings.Contains(view, "4 total") {
		t.Error("expected view to contain '4 total'")
	}
	if !strings.Contains(view, "Pending:     1") {
		t.Error("expected view to contain pending count")
	}
	if !strings.Contains(view, "In Progress: 1") {
		t.Error("expected view to contain in-progress count")
	}

	// Footer should show key bindings
	if !strings.Contains(view, "q: quit") {
		t.Error("expected view to contain quit hint")
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
	if !strings.Contains(view, "0 total") {
		t.Error("expected '0 total' for empty task list")
	}
}
