package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// WriteJSON encodes v as indented JSON to w.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteYAML encodes v as YAML to w.
func WriteYAML(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(v)
}

// ValidateFormat checks that format is one of the supported values.
func ValidateFormat(format string, supported []string) error {
	if slices.Contains(supported, format) {
		return nil
	}
	return fmt.Errorf("unsupported format: %s (supported: %s)", format, strings.Join(supported, ", "))
}
