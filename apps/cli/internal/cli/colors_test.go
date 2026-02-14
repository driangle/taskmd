package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestColorsEnabled_NoColorFlag(t *testing.T) {
	noColor = true
	forceColor = false
	defer func() { noColor = false }()
	os.Unsetenv("NO_COLOR")

	if colorsEnabled() {
		t.Error("Expected colors disabled when noColor flag is true")
	}
}

func TestColorsEnabled_NoColorEnvVar(t *testing.T) {
	noColor = false
	forceColor = true
	defer func() { forceColor = false }()

	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	if colorsEnabled() {
		t.Error("Expected colors disabled when NO_COLOR env var is set")
	}
}

func TestColorsEnabled_ForceColor(t *testing.T) {
	noColor = false
	forceColor = true
	defer func() { forceColor = false }()
	os.Unsetenv("NO_COLOR")

	if !colorsEnabled() {
		t.Error("Expected colors enabled when forceColor is true")
	}
}

func TestColorsEnabled_PipeDisablesColors(t *testing.T) {
	noColor = false
	forceColor = false
	os.Unsetenv("NO_COLOR")

	// In tests, stdout is a pipe, so colors should be disabled
	if colorsEnabled() {
		t.Error("Expected colors disabled when stdout is a pipe (not a TTY)")
	}
}

func TestGetRenderer_ColorProfile(t *testing.T) {
	tests := []struct {
		name       string
		noColor    bool
		forceColor bool
		wantASCII  bool
	}{
		{"force color on", false, true, false},
		{"no-color flag", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noColor = tt.noColor
			forceColor = tt.forceColor
			defer func() {
				noColor = false
				forceColor = false
			}()
			os.Unsetenv("NO_COLOR")

			r := getRenderer()
			if tt.wantASCII && r.ColorProfile() != termenv.Ascii {
				t.Error("Expected Ascii color profile")
			}
			if !tt.wantASCII && r.ColorProfile() == termenv.Ascii {
				t.Error("Expected non-Ascii color profile")
			}
		})
	}
}

// ansiRenderer returns a renderer with ANSI256 colors enabled for testing.
func ansiRenderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(termenv.ANSI256)
	return r
}

// asciiRenderer returns a renderer with no colors for testing.
func asciiRenderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(os.Stdout)
	r.SetColorProfile(termenv.Ascii)
	return r
}

func TestFormatHelpers_WithColors(t *testing.T) {
	r := ansiRenderer()

	tests := []struct {
		name string
		fn   func() string
	}{
		{"formatTaskID", func() string { return formatTaskID("001", r) }},
		{"formatStatus", func() string { return formatStatus("completed", r) }},
		{"formatPriority", func() string { return formatPriority("high", r) }},
		{"formatEffort", func() string { return formatEffort("small", r) }},
		{"formatSuccess", func() string { return formatSuccess("ok", r) }},
		{"formatError", func() string { return formatError("fail", r) }},
		{"formatWarning", func() string { return formatWarning("warn", r) }},
		{"formatLabel", func() string { return formatLabel("Key:", r) }},
		{"formatDim", func() string { return formatDim("path", r) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if !strings.Contains(result, "\x1b[") {
				t.Errorf("Expected ANSI codes in %s output, got: %q", tt.name, result)
			}
		})
	}
}

func TestFormatHelpers_WithoutColors(t *testing.T) {
	r := asciiRenderer()

	tests := []struct {
		name     string
		fn       func() string
		contains string
	}{
		{"formatTaskID", func() string { return formatTaskID("001", r) }, "001"},
		{"formatStatus", func() string { return formatStatus("completed", r) }, "completed"},
		{"formatPriority", func() string { return formatPriority("high", r) }, "high"},
		{"formatEffort", func() string { return formatEffort("small", r) }, "small"},
		{"formatSuccess", func() string { return formatSuccess("ok", r) }, "ok"},
		{"formatError", func() string { return formatError("fail", r) }, "fail"},
		{"formatWarning", func() string { return formatWarning("warn", r) }, "warn"},
		{"formatLabel", func() string { return formatLabel("Key:", r) }, "Key:"},
		{"formatDim", func() string { return formatDim("path", r) }, "path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if strings.Contains(result, "\x1b[") {
				t.Errorf("Expected no ANSI codes in %s output, got: %q", tt.name, result)
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected %q in %s output, got: %q", tt.contains, tt.name, result)
			}
		})
	}
}

func TestEffortColors(t *testing.T) {
	r := ansiRenderer()

	small := formatEffort("small", r)
	medium := formatEffort("medium", r)
	large := formatEffort("large", r)

	// Each should contain the text and ANSI codes
	if !strings.Contains(small, "small") || !strings.Contains(small, "\x1b[") {
		t.Errorf("Expected colored 'small' effort, got: %q", small)
	}
	if !strings.Contains(medium, "medium") || !strings.Contains(medium, "\x1b[") {
		t.Errorf("Expected colored 'medium' effort, got: %q", medium)
	}
	if !strings.Contains(large, "large") || !strings.Contains(large, "\x1b[") {
		t.Errorf("Expected colored 'large' effort, got: %q", large)
	}
}

func TestFormatHeading_Effort(t *testing.T) {
	r := ansiRenderer()

	result := formatHeading("small", "effort", r)
	if !strings.Contains(result, "small") {
		t.Errorf("Expected 'small' in heading, got: %q", result)
	}
	if !strings.Contains(result, "\x1b[") {
		t.Error("Expected ANSI codes in effort heading")
	}
}
