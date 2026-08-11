package filter

import (
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/model"
)

func TestApply_OwnerFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Owner: "alice"},
		{ID: "002", Title: "Task B", Owner: "bob"},
		{ID: "003", Title: "Task C", Owner: ""},
	}

	filtered, err := Apply(tasks, []string{"owner=alice"}, effort.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(filtered))
	}
	if filtered[0].ID != "001" {
		t.Errorf("expected task 001, got %s", filtered[0].ID)
	}
}

func TestApply_MultipleFilters(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Status: model.StatusPending, Owner: "alice"},
		{ID: "002", Title: "Task B", Status: model.StatusPending, Owner: "bob"},
		{ID: "003", Title: "Task C", Status: model.StatusCompleted, Owner: "alice"},
	}

	filtered, err := Apply(tasks, []string{"status=pending", "owner=alice"}, effort.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(filtered))
	}
	if filtered[0].ID != "001" {
		t.Errorf("expected task 001, got %s", filtered[0].ID)
	}
}

func TestApply_InvalidFilterFormat(t *testing.T) {
	tasks := []*model.Task{{ID: "001"}}

	_, err := Apply(tasks, []string{"badfilter"}, effort.Default())
	if err == nil {
		t.Fatal("expected error for invalid filter format")
	}
}

func TestApply_TypeFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Feature A", Type: model.TypeFeature},
		{ID: "002", Title: "Bug fix B", Type: model.TypeBug},
		{ID: "003", Title: "Chore C", Type: model.TypeChore},
		{ID: "004", Title: "No type"},
	}

	filtered, err := Apply(tasks, []string{"type=bug"}, effort.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(filtered))
	}
	if filtered[0].ID != "002" {
		t.Errorf("expected task 002, got %s", filtered[0].ID)
	}
}

func TestApply_GroupWildcardFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Group: "cli/graph"},
		{ID: "002", Title: "Task B", Group: "cli/next"},
		{ID: "003", Title: "Task C", Group: "web/board"},
	}

	filtered, err := Apply(tasks, []string{"group=cli/*"}, effort.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(filtered))
	}
	if filtered[0].ID != "001" || filtered[1].ID != "002" {
		t.Errorf("expected tasks 001 and 002, got %s and %s", filtered[0].ID, filtered[1].ID)
	}
}

func TestApply_TouchesWildcardFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Touches: []string{"cli/graph", "web/board"}},
		{ID: "002", Title: "Task B", Touches: []string{"web/api"}},
		{ID: "003", Title: "Task C", Touches: []string{"docs"}},
	}

	filtered, err := Apply(tasks, []string{"touches=web/*"}, effort.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(filtered))
	}
	if filtered[0].ID != "001" || filtered[1].ID != "002" {
		t.Errorf("expected tasks 001 and 002, got %s and %s", filtered[0].ID, filtered[1].ID)
	}
}

func TestApply_GroupExactStillWorks(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Group: "cli"},
		{ID: "002", Title: "Task B", Group: "web"},
	}

	filtered, err := Apply(tasks, []string{"group=cli"}, effort.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 1 || filtered[0].ID != "001" {
		t.Fatalf("expected task 001, got %v", filtered)
	}
}

func TestApply_ParentFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Parent"},
		{ID: "002", Title: "Child A", Parent: "001"},
		{ID: "003", Title: "Child B", Parent: "001"},
		{ID: "004", Title: "Orphan"},
	}

	t.Run("filter by parent ID", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"parent=001"}, effort.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(filtered))
		}
	})

	t.Run("filter parent=true", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"parent=true"}, effort.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks with parent, got %d", len(filtered))
		}
	})

	t.Run("filter parent=false", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"parent=false"}, effort.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks without parent, got %d", len(filtered))
		}
	})
}

func TestApply_PriorityComparison(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Low", Priority: model.PriorityLow},
		{ID: "002", Title: "Medium", Priority: model.PriorityMedium},
		{ID: "003", Title: "High", Priority: model.PriorityHigh},
		{ID: "004", Title: "Critical", Priority: model.PriorityCritical},
		{ID: "005", Title: "Unset"},
	}

	tests := []struct {
		name        string
		expr        string
		expectedIDs []string
	}{
		{">=medium", "priority>=medium", []string{"002", "003", "004"}},
		{">medium", "priority>medium", []string{"003", "004"}},
		{"<=medium", "priority<=medium", []string{"001", "002"}},
		{"<medium", "priority<medium", []string{"001"}},
		{">=critical", "priority>=critical", []string{"004"}},
		{">critical", "priority>critical", nil},
		{"<=low", "priority<=low", []string{"001"}},
		{"<low", "priority<low", nil},
		{">=low (all set)", "priority>=low", []string{"001", "002", "003", "004"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := Apply(tasks, []string{tt.expr}, effort.Default())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(filtered) != len(tt.expectedIDs) {
				t.Fatalf("expected %d tasks, got %d", len(tt.expectedIDs), len(filtered))
			}
			for i, id := range tt.expectedIDs {
				if filtered[i].ID != id {
					t.Errorf("result[%d]: expected %s, got %s", i, id, filtered[i].ID)
				}
			}
		})
	}
}

func TestApply_EffortComparison(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Small", Effort: model.EffortSmall},
		{ID: "002", Title: "Medium", Effort: model.EffortMedium},
		{ID: "003", Title: "Large", Effort: model.EffortLarge},
		{ID: "004", Title: "Unset"},
	}

	tests := []struct {
		name        string
		expr        string
		expectedIDs []string
	}{
		{">=medium", "effort>=medium", []string{"002", "003"}},
		{">small", "effort>small", []string{"002", "003"}},
		{"<=medium", "effort<=medium", []string{"001", "002"}},
		{"<large", "effort<large", []string{"001", "002"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := Apply(tasks, []string{tt.expr}, effort.Default())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(filtered) != len(tt.expectedIDs) {
				t.Fatalf("expected %d tasks, got %d", len(tt.expectedIDs), len(filtered))
			}
			for i, id := range tt.expectedIDs {
				if filtered[i].ID != id {
					t.Errorf("result[%d]: expected %s, got %s", i, id, filtered[i].ID)
				}
			}
		})
	}
}

func TestApply_ComparisonErrors(t *testing.T) {
	tasks := []*model.Task{{ID: "001"}}

	tests := []struct {
		name string
		expr string
	}{
		{"unsupported field", "status>=pending"},
		{"invalid priority value", "priority>=unknown"},
		{"invalid effort value", "effort>huge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Apply(tasks, []string{tt.expr}, effort.Default())
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestApply_ComparisonCombinedWithEquality(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Status: model.StatusPending, Priority: model.PriorityLow},
		{ID: "002", Status: model.StatusPending, Priority: model.PriorityHigh},
		{ID: "003", Status: model.StatusCompleted, Priority: model.PriorityHigh},
	}

	filtered, err := Apply(tasks, []string{"status=pending", "priority>=high"}, effort.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != "002" {
		t.Fatalf("expected task 002, got %v", filtered)
	}
}

func TestParseExpr(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		field   string
		op      string
		value   string
		wantErr bool
	}{
		{"equality", "status=pending", "status", "=", "pending", false},
		{"gte", "priority>=high", "priority", ">=", "high", false},
		{"gt", "priority>low", "priority", ">", "low", false},
		{"lte", "effort<=medium", "effort", "<=", "medium", false},
		{"lt", "effort<large", "effort", "<", "large", false},
		{"equality with spaces", " status = pending ", "status", "=", "pending", false},
		{"missing value", "badfilter", "", "", "", true},
		{"unsupported field for op", "status>=pending", "", "", "", true},
		{"invalid value for op", "priority>=nope", "", "", "", true},
		{"missing value after op", "priority>=", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := parseExpr(tt.expr, effort.Default())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.Field != tt.field || c.Op != tt.op || c.Value != tt.value {
				t.Errorf("got {%s %s %s}, want {%s %s %s}", c.Field, c.Op, c.Value, tt.field, tt.op, tt.value)
			}
		})
	}
}

func TestApply_PhaseFilter(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", Phase: "v0.2"},
		{ID: "002", Title: "Task B", Phase: "v0.3"},
		{ID: "003", Title: "Task C", Phase: "v0.2"},
		{ID: "004", Title: "Task D"},
	}

	t.Run("exact match", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"phase=v0.2"}, effort.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(filtered))
		}
		if filtered[0].ID != "001" || filtered[1].ID != "003" {
			t.Errorf("expected tasks 001 and 003, got %s and %s", filtered[0].ID, filtered[1].ID)
		}
	})

	t.Run("no match", func(t *testing.T) {
		filtered, err := Apply(tasks, []string{"phase=v1.0"}, effort.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(filtered) != 0 {
			t.Fatalf("expected 0 tasks, got %d", len(filtered))
		}
	})
}

// sentinelTasks provides a fixed set with a mix of set/unset optional fields.
func sentinelTasks() []*model.Task {
	return []*model.Task{
		{ID: "001", Title: "A", Phase: "v0.2", Owner: "alice", Group: "cli", Priority: model.PriorityHigh},
		{ID: "002", Title: "B", Phase: "v0.3", Owner: "", Group: "web", Priority: ""},
		{ID: "003", Title: "C", Phase: "", Owner: "bob", Group: "", Priority: model.PriorityLow},
	}
}

func TestApply_SentinelNoneAny(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantIDs []string
	}{
		{"phase none", "phase=none", []string{"003"}},
		{"phase any", "phase=any", []string{"001", "002"}},
		{"owner none", "owner=none", []string{"002"}},
		{"owner any", "owner=any", []string{"001", "003"}},
		{"group none", "group=none", []string{"003"}},
		{"group any", "group=any", []string{"001", "002"}},
		{"priority none", "priority=none", []string{"002"}},
		{"priority any", "priority=any", []string{"001", "003"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFilterIDs(t, sentinelTasks(), tt.expr, tt.wantIDs)
		})
	}
}

func TestApply_EmptyValueAliasesNone(t *testing.T) {
	// field= (empty RHS) must behave identically to field=none.
	assertFilterIDs(t, sentinelTasks(), "phase=", []string{"003"})
	assertFilterIDs(t, sentinelTasks(), "owner=", []string{"002"})
}

func TestApply_SentinelCombined(t *testing.T) {
	// phase=none AND owner=any narrows to task 003 (no phase, has owner).
	assertFilterIDs(t, sentinelTasks(), "phase=none", []string{"003"})
	filtered, err := Apply(sentinelTasks(), []string{"phase=none", "owner=any"}, effort.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != "003" {
		t.Fatalf("expected only task 003, got %v", ids(filtered))
	}
}

func TestApply_BlockedTrueFalseAliases(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "A", Dependencies: []string{"099"}},
		{ID: "002", Title: "B"},
	}
	// true/false retained as aliases for any/none.
	assertFilterIDs(t, tasks, "blocked=true", []string{"001"})
	assertFilterIDs(t, tasks, "blocked=any", []string{"001"})
	assertFilterIDs(t, tasks, "blocked=false", []string{"002"})
	assertFilterIDs(t, tasks, "blocked=none", []string{"002"})
}

func TestApply_ParentTrueFalseAliases(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Parent"},
		{ID: "002", Title: "Child", Parent: "001"},
	}
	assertFilterIDs(t, tasks, "parent=true", []string{"002"})
	assertFilterIDs(t, tasks, "parent=any", []string{"002"})
	assertFilterIDs(t, tasks, "parent=false", []string{"001"})
	assertFilterIDs(t, tasks, "parent=none", []string{"001"})
	// Exact-value matching still works alongside the sentinels.
	assertFilterIDs(t, tasks, "parent=001", []string{"002"})
}

// assertFilterIDs applies a single filter expression and asserts the resulting
// task IDs match wantIDs (order-sensitive; input order is preserved by Apply).
func assertFilterIDs(t *testing.T, tasks []*model.Task, expr string, wantIDs []string) {
	t.Helper()
	filtered, err := Apply(tasks, []string{expr}, effort.Default())
	if err != nil {
		t.Fatalf("Apply(%q, effort.Default()) unexpected error: %v", expr, err)
	}
	got := ids(filtered)
	if len(got) != len(wantIDs) {
		t.Fatalf("Apply(%q, effort.Default()) = %v, want %v", expr, got, wantIDs)
	}
	for i := range wantIDs {
		if got[i] != wantIDs[i] {
			t.Fatalf("Apply(%q, effort.Default()) = %v, want %v", expr, got, wantIDs)
		}
	}
}

func ids(tasks []*model.Task) []string {
	out := make([]string, len(tasks))
	for i, task := range tasks {
		out[i] = task.ID
	}
	return out
}

// --- Configurable effort vocabulary ---

func TestApply_CustomEffortVocabulary(t *testing.T) {
	t.Parallel()

	scale, err := effort.NewScale([]string{"xs", "s", "m", "l", "xl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tasks := []*model.Task{
		{ID: "001", Effort: "xs"},
		{ID: "002", Effort: "m"},
		{ID: "003", Effort: "xl"},
	}

	tests := []struct {
		name    string
		expr    string
		wantIDs []string
	}{
		{"equality on a custom value", "effort=xs", []string{"001"}},
		{"greater than", "effort>s", []string{"002", "003"}},
		{"greater or equal", "effort>=m", []string{"002", "003"}},
		{"less than", "effort<m", []string{"001"}},
		{"less or equal", "effort<=m", []string{"001", "002"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Apply(tasks, []string{tt.expr}, scale)
			if err != nil {
				t.Fatalf("Apply(%q) failed: %v", tt.expr, err)
			}
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("Apply(%q) returned %d tasks, want %d", tt.expr, len(got), len(tt.wantIDs))
			}
			for i, id := range tt.wantIDs {
				if got[i].ID != id {
					t.Errorf("result[%d].ID = %q, want %q", i, got[i].ID, id)
				}
			}
		})
	}
}

// A comparison against a value outside the configured vocabulary must be an
// error naming the configured values, not a silent empty result.
func TestApply_OrdinalErrorListsConfiguredEffortValues(t *testing.T) {
	t.Parallel()

	scale, err := effort.NewScale([]string{"xs", "s", "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = Apply(nil, []string{"effort>small"}, scale)
	if err == nil {
		t.Fatal("expected an error for a value outside the vocabulary")
	}
	if !strings.Contains(err.Error(), "xs, s, m") {
		t.Errorf("error = %q, want it to list the configured values", err.Error())
	}
}

// Priority ordering is not configurable and must be unaffected by the effort scale.
func TestApply_PriorityOrderingUnaffectedByEffortScale(t *testing.T) {
	t.Parallel()

	scale, err := effort.NewScale([]string{"xs", "s"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tasks := []*model.Task{
		{ID: "001", Priority: model.PriorityLow},
		{ID: "002", Priority: model.PriorityCritical},
	}

	got, err := Apply(tasks, []string{"priority>=high"}, scale)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if len(got) != 1 || got[0].ID != "002" {
		t.Errorf("got %v, want only task 002", got)
	}
}
