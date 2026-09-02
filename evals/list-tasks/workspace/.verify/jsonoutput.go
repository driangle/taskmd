package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// The JSON eval grades output *shape* as well as content, which the
// table-format evals cannot. It is the one place where the exact rendering is
// the thing under test, so parsing — not token matching — is the right tool.

// assertJSONReported asserts the agent's output contains a parseable JSON array
// of task objects whose IDs are exactly `want`, each carrying an id and a title.
func assertJSONReported(output string, want ...string) error {
	arr, err := firstTaskArray(output)
	if err != nil {
		return err
	}

	got := map[string]bool{}
	for i, obj := range arr {
		id := field(obj, "id")
		if id == "" {
			return fmt.Errorf("element %d has no non-empty \"id\" field: %v", i, keysOf(obj))
		}
		if field(obj, "title") == "" {
			return fmt.Errorf("element %d (id %s) has no non-empty \"title\" field: %v", i, id, keysOf(obj))
		}
		got[id] = true
	}

	wanted := map[string]bool{}
	for _, id := range want {
		wanted[id] = true
	}

	var missing, extra []string
	for id := range wanted {
		if !got[id] {
			missing = append(missing, id)
		}
	}
	for id := range got {
		if !wanted[id] {
			extra = append(extra, id)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)

	if len(missing) > 0 || len(extra) > 0 {
		return fmt.Errorf("JSON contains %v, want exactly %v (missing %v, unexpected %v)",
			sortedSet(got), want, missing, extra)
	}
	return nil
}

// firstTaskArray finds the first JSON array in the output that parses as a
// non-empty list of objects. Scanning every `[` rather than assuming a fenced
// block means it works whether the agent wrapped the JSON in ```json, printed
// it bare, or nested it under a key like {"tasks": [...]}.
func firstTaskArray(output string) ([]map[string]any, error) {
	var sawArray bool

	for i, r := range output {
		if r != '[' {
			continue
		}
		end, ok := matchBracket(output, i)
		if !ok {
			continue
		}

		var arr []map[string]any
		if err := json.Unmarshal([]byte(output[i:end+1]), &arr); err != nil {
			continue
		}
		sawArray = true
		if len(arr) > 0 {
			return arr, nil
		}
	}

	if sawArray {
		return nil, fmt.Errorf("output contains a JSON array, but it is empty")
	}
	return nil, fmt.Errorf("output contains no parseable JSON array of task objects")
}

// matchBracket returns the index of the `]` closing the `[` at start, honouring
// string literals and escapes so a bracket inside a title does not confuse it.
func matchBracket(s string, start int) (int, bool) {
	var (
		depth    int
		inString bool
		escaped  bool
	)

	for i := start; i < len(s); i++ {
		c := s[i]
		switch {
		case escaped:
			escaped = false
		case c == '\\' && inString:
			escaped = true
		case c == '"':
			inString = !inString
		case inString:
			// nothing to track inside a string literal
		case c == '[':
			depth++
		case c == ']':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

// field returns the named field case-insensitively, as a trimmed string.
func field(obj map[string]any, name string) string {
	for k, v := range obj {
		if !strings.EqualFold(k, name) {
			continue
		}
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func keysOf(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
