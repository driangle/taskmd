package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/nextid"
)

func createNextIDTestFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", name, err)
		}
	}
	return tmpDir
}

func resetNextIDFlags() {
	nextIDFormat = "plain"
}

func TestNextID_NumericIDs(t *testing.T) {
	resetNextIDFlags()

	tmpDir := createNextIDTestFiles(t, map[string]string{
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

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runNextID(nextIDCmd, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != "006" {
		t.Errorf("expected 006, got %q", output)
	}
}

func TestNextID_EmptyDirectory(t *testing.T) {
	resetNextIDFlags()

	tmpDir := t.TempDir()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runNextID(nextIDCmd, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != "001" {
		t.Errorf("expected 001, got %q", output)
	}
}

func TestNextID_PrefixedIDs(t *testing.T) {
	resetNextIDFlags()

	tmpDir := createNextIDTestFiles(t, map[string]string{
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

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runNextID(nextIDCmd, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != "WEB-004" {
		t.Errorf("expected WEB-004, got %q", output)
	}
}

func TestNextID_JSONFormat(t *testing.T) {
	resetNextIDFlags()
	nextIDFormat = "json"

	tmpDir := createNextIDTestFiles(t, map[string]string{
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

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runNextID(nextIDCmd, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result nextid.Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, buf.String())
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
	resetNextIDFlags()
	nextIDFormat = "yaml"

	tmpDir := t.TempDir()

	err := runNextID(nextIDCmd, []string{tmpDir})
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", err)
	}
}

func TestNextID_PlainFormat(t *testing.T) {
	resetNextIDFlags()
	nextIDFormat = "plain"

	tmpDir := createNextIDTestFiles(t, map[string]string{
		"010-task.md": `---
id: "010"
title: "Task ten"
status: pending
priority: medium
created: 2026-02-14
---
`,
	})

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runNextID(nextIDCmd, []string{tmpDir})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != "011" {
		t.Errorf("expected 011, got %q", output)
	}
}
