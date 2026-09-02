package main

// checks maps a check name (referenced from suite.yaml) to its implementation.
//
// The `list-*` checks are wired to `check_output` steps and grade the agent's
// reported output. `no-mutation` is wired to a `check` step and grades the
// filesystem. Every eval runs both.
var checks = map[string]func() error{
	"list-all":           gradeOutput(listAll),
	"list-status-filter": gradeOutput(statusFilter),
	"list-scope-filter":  gradeOutput(scopeFilter),
	"list-json-format":   gradeOutput(jsonFormat),
	"list-phase-filter":  gradeOutput(phaseFilter),
	"no-mutation":        fixtureUnchanged,
}

// gradeOutput turns an assertion over the agent's answer into a runnable check.
// Splitting the stdin read from the assertion is what lets output_test.go
// exercise every grader in both directions without a running agent.
func gradeOutput(assert func(string) error) func() error {
	return func() error {
		out, err := readOutput()
		if err != nil {
			return err
		}
		return assert(out)
	}
}

// listAll expects every task, including the completed one and the two nested
// under tasks/cli/ and tasks/web/. This is the assertion that catches an agent
// which only globs the root task directory.
func listAll(out string) error {
	return assertReported(out, "001", "002", "003", "004", "005", "006")
}

// statusFilter expects the four pending tasks. 001 is in-progress and 005 is
// completed, so both must be absent — the negative half is the point, since an
// agent that dumps all six would pass a presence-only check.
func statusFilter(out string) error {
	return assertReported(out, "002", "003", "004", "006")
}

// scopeFilter expects the two tasks in the cli group, and only those.
func scopeFilter(out string) error {
	return assertReported(out, "001", "005")
}

// jsonFormat expects the high-priority tasks as a parseable JSON array carrying
// id and title — output shape as well as content.
func jsonFormat(out string) error {
	return assertJSONReported(out, "001", "005", "006")
}

// phaseFilter expects the four mvp tasks. 004 and 006 are in `polish`.
func phaseFilter(out string) error {
	return assertReported(out, "001", "002", "003", "005")
}
