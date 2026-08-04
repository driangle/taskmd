package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestTableWriter_BasicAlignment(t *testing.T) {
	t.Parallel()
	tw := NewTableWriter()
	tw.AddHeader([]string{"ID", "TITLE", "STATUS"})
	tw.AddSeparator()
	tw.AddRow(
		[]string{"001", "Short", "pending"},
		[]string{"001", "Short", "pending"},
	)
	tw.AddRow(
		[]string{"002", "A much longer title", "in-progress"},
		[]string{"002", "A much longer title", "in-progress"},
	)

	var buf bytes.Buffer
	tw.Flush(&buf)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	// All lines should have the same visible width (trimmed trailing space aside)
	// Check that separator dashes match column widths
	sepCols := strings.Split(lines[1], "  ")
	if len(sepCols) != 3 {
		t.Fatalf("expected 3 separator columns, got %d", len(sepCols))
	}

	// Separator for TITLE column should be at least as wide as "A much longer title"
	titleSep := sepCols[1]
	if len(titleSep) < len("A much longer title") {
		t.Errorf("title separator %q (%d) shorter than longest value (%d)",
			titleSep, len(titleSep), len("A much longer title"))
	}
}

func TestTableWriter_ColorAlignment(t *testing.T) {
	t.Parallel()
	tw := NewTableWriter()
	tw.AddHeader([]string{"ID", "STATUS"})
	tw.AddSeparator()

	// Simulate colored values (ANSI codes add invisible bytes)
	tw.AddRow(
		[]string{"001", "pending"},
		[]string{"\x1b[36m001\x1b[0m", "\x1b[33mpending\x1b[0m"},
	)
	tw.AddRow(
		[]string{"002", "in-progress"},
		[]string{"\x1b[36m002\x1b[0m", "\x1b[32min-progress\x1b[0m"},
	)

	var colorBuf bytes.Buffer
	tw.Flush(&colorBuf)

	// Build the same table without color for comparison
	tw2 := NewTableWriter()
	tw2.AddHeader([]string{"ID", "STATUS"})
	tw2.AddSeparator()
	tw2.AddRow([]string{"001", "pending"}, []string{"001", "pending"})
	tw2.AddRow([]string{"002", "in-progress"}, []string{"002", "in-progress"})

	var plainBuf bytes.Buffer
	tw2.Flush(&plainBuf)

	colorLines := strings.Split(strings.TrimRight(colorBuf.String(), "\n"), "\n")
	plainLines := strings.Split(strings.TrimRight(plainBuf.String(), "\n"), "\n")

	if len(colorLines) != len(plainLines) {
		t.Fatalf("line count mismatch: color=%d, plain=%d", len(colorLines), len(plainLines))
	}

	// After stripping ANSI, each line should have the same visible content as the plain version
	for i := range colorLines {
		stripped := StripANSI(colorLines[i])
		if stripped != plainLines[i] {
			t.Errorf("line %d mismatch:\n  stripped: %q\n  plain:    %q", i, stripped, plainLines[i])
		}
	}
}

func TestTableWriter_EmptyTable(t *testing.T) {
	t.Parallel()
	tw := NewTableWriter()

	var buf bytes.Buffer
	tw.Flush(&buf)

	if buf.String() != "" {
		t.Errorf("expected empty output for empty table, got %q", buf.String())
	}
}

func TestTableWriter_SeparatorSizing(t *testing.T) {
	t.Parallel()
	tw := NewTableWriter()
	tw.AddHeader([]string{"A", "BB"})
	tw.AddSeparator()
	tw.AddRow([]string{"XXXX", "Y"}, []string{"XXXX", "Y"})

	var buf bytes.Buffer
	tw.Flush(&buf)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Separator should use max widths: "XXXX" (4) and "BB" (2)
	sepParts := strings.Split(lines[1], "  ")
	if len(sepParts) != 2 {
		t.Fatalf("expected 2 separator parts, got %d", len(sepParts))
	}
	if sepParts[0] != "----" {
		t.Errorf("first sep = %q, want %q", sepParts[0], "----")
	}
	if sepParts[1] != "--" {
		t.Errorf("second sep = %q, want %q", sepParts[1], "--")
	}
}

func TestStripANSI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no ansi", "hello", "hello"},
		{"single code", "\x1b[36mhello\x1b[0m", "hello"},
		{"multiple codes", "\x1b[1m\x1b[36mhello\x1b[0m", "hello"},
		{"mixed", "a\x1b[31mb\x1b[0mc", "abc"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StripANSI(tt.input)
			if got != tt.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTableWriter_HeaderOnly(t *testing.T) {
	t.Parallel()
	tw := NewTableWriter()
	tw.AddHeader([]string{"COL1", "COL2"})
	tw.AddSeparator()

	var buf bytes.Buffer
	tw.Flush(&buf)

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "COL1") {
		t.Error("expected header to contain COL1")
	}
	if lines[1] != "----  ----" {
		t.Errorf("expected separator %q, got %q", "----  ----", lines[1])
	}
}
