package cli

import (
	"strings"
	"testing"
)

func TestSuggestValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		valid    []string
		expected string
	}{
		{"exact match", "pending", validStatusValues, "pending"},
		{"close typo", "pening", validStatusValues, "pending"},
		{"complted -> completed", "complted", validStatusValues, "completed"},
		{"hig -> high", "hig", validPriorityValues, "high"},
		{"medim -> medium", "medim", validPriorityValues, "medium"},
		{"smal -> small", "smal", validEffortValues, "small"},
		{"totally wrong", "zzzzzzzzz", validStatusValues, ""},
		{"case insensitive", "PENDING", validStatusValues, "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestValue(tt.input, tt.valid)
			if got != tt.expected {
				t.Errorf("suggestValue(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestInvalidValueError_WithSuggestion(t *testing.T) {
	err := invalidValueError("status", "pening", validStatusValues)
	errMsg := err.Error()

	if !strings.Contains(errMsg, `"pening"`) {
		t.Errorf("expected error to contain the invalid value, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, `did you mean "pending"`) {
		t.Errorf("expected error to suggest 'pending', got: %s", errMsg)
	}
}

func TestInvalidValueError_NoSuggestion(t *testing.T) {
	err := invalidValueError("status", "zzzzzzzzz", validStatusValues)
	errMsg := err.Error()

	if strings.Contains(errMsg, "did you mean") {
		t.Errorf("expected no suggestion for totally wrong input, got: %s", errMsg)
	}
}
