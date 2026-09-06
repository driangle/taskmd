package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// checkGroupIsScannable rejects a --group whose destination the scanner would
// never descend into. Writing there succeeds and then the task is invisible to
// list, get, board, stats, next and validate — a silent failure the user only
// discovers when they go looking for the task, so `add` refuses up front.
//
// The comparison delegates to the scanner rather than matching paths here:
// ignore entries are bare directory names matched at any depth, so "a/content"
// is just as unreachable as "content".
func checkGroupIsScannable(scanDir, group string, flags GlobalFlags) error {
	if group == "" {
		return nil
	}

	segment := newTaskScanner(scanDir, flags).SkippedSegment(group)
	if segment == "" {
		return nil
	}

	return fmt.Errorf("--group %q resolves to %s, which is %s. The task would not be "+
		"visible to list, get or validate. %s",
		group,
		displayGroupPath(scanDir, group),
		skipReason(group, segment, flags.IgnoreDirs),
		skipRemedy(segment, flags.IgnoreDirs),
	)
}

// skipReason explains why the scanner skips segment, naming the config key when
// the user put it there so the message points at something they can change.
// The offending segment is called out only when it is not the whole group, so a
// nested group says which of its components is at fault without repeating
// itself for the common top-level case.
func skipReason(group, segment string, ignoreDirs []string) string {
	blamed := ""
	if segment != group {
		blamed = fmt.Sprintf(" (%q)", segment)
	}

	switch {
	case slices.Contains(ignoreDirs, segment):
		return fmt.Sprintf("excluded by the 'ignore' key in .taskmd.yaml%s", blamed)
	case strings.HasPrefix(segment, "."):
		return fmt.Sprintf("hidden%s — names beginning with a dot are always skipped by the scanner", blamed)
	default:
		return fmt.Sprintf("always skipped by the scanner%s", blamed)
	}
}

// skipRemedy is the actionable half of the message: what the user can do about
// the segment that blocked the write.
func skipRemedy(segment string, ignoreDirs []string) string {
	if slices.Contains(ignoreDirs, segment) {
		return "Remove it from 'ignore', or choose another group"
	}
	return "Choose a group the scanner reads"
}

// displayGroupPath renders the destination relative to the working directory
// when it sits underneath it, so the message reads "tasks/content" rather than
// an absolute path.
func displayGroupPath(scanDir, group string) string {
	full := filepath.Join(scanDir, group)
	cwd, err := os.Getwd()
	if err != nil {
		return full
	}
	rel, err := filepath.Rel(cwd, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return full
	}
	return rel
}
