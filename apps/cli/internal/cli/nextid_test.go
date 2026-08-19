package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/nextid"
)

// nextIDStdout runs `next-id <args...>` against repo, fails on error, returns stdout.
func nextIDStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"next-id"}, args...)...)
	if res.Err != nil {
		t.Fatalf("next-id %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestNextID_NumericIDs(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-first.md": `---
id: "001"
title: "First task"
status: pending
priority: medium
created: 2026-02-14
---
`,
		"002-second.md": `---
id: "002"
title: "Second task"
status: pending
priority: medium
created: 2026-02-14
---
`,
		"005-fifth.md": `---
id: "005"
title: "Fifth task (gap)"
status: pending
priority: medium
created: 2026-02-14
---
`,
	})

	output := strings.TrimSpace(nextIDStdout(t, repo))
	if output != "006" {
		t.Errorf("expected 006, got %q", output)
	}
}

func TestNextID_EmptyDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := strings.TrimSpace(nextIDStdout(t, repo))
	if output != "001" {
		t.Errorf("expected 001, got %q", output)
	}
}

func TestNextID_PrefixedIDs(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"WEB-001.md": `---
id: "WEB-001"
title: "Web task 1"
status: pending
priority: medium
created: 2026-02-14
---
`,
		"WEB-002.md": `---
id: "WEB-002"
title: "Web task 2"
status: pending
priority: medium
created: 2026-02-14
---
`,
		"WEB-003.md": `---
id: "WEB-003"
title: "Web task 3"
status: pending
priority: medium
created: 2026-02-14
---
`,
	})

	output := strings.TrimSpace(nextIDStdout(t, repo))
	if output != "WEB-004" {
		t.Errorf("expected WEB-004, got %q", output)
	}
}

func TestNextID_JSONFormat(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-task.md": `---
id: "001"
title: "Task one"
status: pending
priority: medium
created: 2026-02-14
---
`,
		"002-task.md": `---
id: "002"
title: "Task two"
status: completed
priority: high
created: 2026-02-14
---
`,
	})

	output := nextIDStdout(t, repo, "--format", "json")

	var result nextid.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, output)
	}

	if result.NextID != "003" {
		t.Errorf("NextID = %q, want %q", result.NextID, "003")
	}
	if result.MaxID != "002" {
		t.Errorf("MaxID = %q, want %q", result.MaxID, "002")
	}
	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
}

func TestNextID_UnsupportedFormat(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("next-id", "--format", "yaml")
	if res.Err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestNextID_PlainFormat(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"010-task.md": `---
id: "010"
title: "Task ten"
status: pending
priority: medium
created: 2026-02-14
---
`,
	})

	output := strings.TrimSpace(nextIDStdout(t, repo, "--format", "plain"))
	if output != "011" {
		t.Errorf("expected 011, got %q", output)
	}
}

func TestNextID_CountsArchivedIDs(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-first.md": `---
id: "001"
title: "First task"
status: pending
priority: medium
created: 2026-02-14
---
`,
		"archive/002-second.md": `---
id: "002"
title: "Second task"
status: completed
priority: medium
created: 2026-02-14
---
`,
	})

	output := strings.TrimSpace(nextIDStdout(t, repo))
	if output != "003" {
		t.Errorf("expected 003 (002 is archived), got %q", output)
	}
}
