package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}

	if err := WriteJSON(&buf, data); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", parsed["key"])
	}

	// Verify indentation
	if !strings.Contains(buf.String(), "  ") {
		t.Error("expected indented JSON output")
	}
}

func TestWriteYAML(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"key": "value"}

	if err := WriteYAML(&buf, data); err != nil {
		t.Fatalf("WriteYAML failed: %v", err)
	}

	// Verify it's valid YAML
	var parsed map[string]string
	if err := yaml.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid YAML: %v", err)
	}
	if parsed["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", parsed["key"])
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		supported []string
		wantErr   bool
	}{
		{"valid format", "json", []string{"table", "json", "yaml"}, false},
		{"invalid format", "csv", []string{"table", "json", "yaml"}, true},
		{"first option valid", "table", []string{"table", "json"}, false},
		{"last option valid", "yaml", []string{"table", "json", "yaml"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.format, tt.supported)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "unsupported format") {
					t.Errorf("expected 'unsupported format' in error, got: %v", err)
				}
				if !strings.Contains(err.Error(), tt.format) {
					t.Errorf("expected format name %q in error, got: %v", tt.format, err)
				}
			}
		})
	}
}
