package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// The JSON eval grades output *shape* as well as content, which the text evals
// cannot. It is the one place where the exact rendering is the thing under
// test, so parsing — not token matching — is the right tool.
//
// `taskmd get <id> --format json` emits a single **object**, where
// `taskmd list --format json` emits an array. Both are accepted: an agent that
// wraps the object in a one-element array has answered the question.

// assertJSONTask asserts the agent's output carries a parseable JSON task
// object for exactly `id`, whose title contains `titleToken` and whose status
// is `status`.
//
// "Exactly" is the two-sided half: an agent that dumps every task as JSON
// answers a different question, and fails here.
func assertJSONTask(output, id, titleToken, status string) error {
	objs, err := taskObjects(output)
	if err != nil {
		return err
	}

	ids := map[string]bool{}
	var match map[string]any
	for _, obj := range objs {
		got := field(obj, "id")
		ids[got] = true
		if got == id {
			match = obj
		}
	}

	if match == nil {
		return fmt.Errorf("JSON carries %v, want task %s", sortedSet(ids), id)
	}
	if len(ids) > 1 {
		return fmt.Errorf("JSON carries %v, want only task %s", sortedSet(ids), id)
	}

	if title := strings.ToLower(field(match, "title")); !strings.Contains(title, titleToken) {
		return fmt.Errorf("task %s has title %q, want one containing %q", id, field(match, "title"), titleToken)
	}
	if got := strings.ToLower(field(match, "status")); got != status {
		return fmt.Errorf("task %s has status %q, want %q", id, field(match, "status"), status)
	}
	return nil
}

// taskObjects returns the JSON objects in the output that look like tasks —
// that is, that carry a non-empty "id".
//
// It scans for both `{` and `[` rather than assuming a fenced block, so it
// works whether the agent printed the object bare, fenced it in ```json, or
// wrapped it in an array. Once a value parses, the scan jumps past it: without
// that, the nested `dependencies.depends_on[0]` object — which carries the
// *dependency's* id — would be counted as a second task and every dependent
// task would fail the "only one task" assertion.
func taskObjects(output string) ([]map[string]any, error) {
	var (
		found    []map[string]any
		sawJSON  bool
		nextFrom int
	)

	for i := 0; i < len(output); i++ {
		if i < nextFrom {
			continue
		}
		c := output[i]
		if c != '{' && c != '[' {
			continue
		}
		end, ok := matchBracket(output, i)
		if !ok {
			continue
		}
		raw := output[i : end+1]

		if c == '{' {
			var obj map[string]any
			if err := json.Unmarshal([]byte(raw), &obj); err != nil {
				continue
			}
			sawJSON = true
			nextFrom = end + 1
			if field(obj, "id") != "" {
				found = append(found, obj)
			}
			continue
		}

		var arr []map[string]any
		if err := json.Unmarshal([]byte(raw), &arr); err != nil {
			continue
		}
		sawJSON = true
		nextFrom = end + 1
		for _, obj := range arr {
			if field(obj, "id") != "" {
				found = append(found, obj)
			}
		}
	}

	if len(found) > 0 {
		return found, nil
	}
	if sawJSON {
		return nil, fmt.Errorf("output contains JSON, but no object carrying a non-empty \"id\"")
	}
	return nil, fmt.Errorf("output contains no parseable JSON task object")
}

// matchBracket returns the index of the delimiter closing the `{` or `[` at
// start, honouring string literals and escapes so a bracket inside a title does
// not confuse it.
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
		case c == '[', c == '{':
			depth++
		case c == ']', c == '}':
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
