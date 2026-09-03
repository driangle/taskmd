package main

// checks maps a check name (referenced from suite.yaml) to its implementation.
//
// The `get-*` checks are wired to `check_output` steps and grade the agent's
// reported output. `no-mutation` is wired to a `check` step and grades the
// filesystem. Every eval runs both.
var checks = map[string]func() error{
	"get-by-id":         gradeOutput(byID),
	"get-by-keyword":    gradeOutput(byKeyword),
	"get-missing":       gradeOutput(missing),
	"get-blocked-state": gradeOutput(blockedState),
	"get-json-format":   gradeOutput(jsonFormat),
	"no-mutation":       fixtureUnchanged,
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

// byID expects the details of task 003 and only those: its status, priority and
// type as recorded in the fixture, and no other fixture task presented as the
// answer.
//
// 002 is absent from the forbidden list on purpose — it is 003's dependency, so
// naming it is correct behavior rather than a leak.
func byID(out string) error {
	if err := assertFocused(out, "003", "001", "004", "005", "006"); err != nil {
		return err
	}
	return assertFields(out, "pending", "critical", "bug")
}

// byKeyword expects the keyword "SSO login bug" resolved to 001 — one behavior,
// resolution, so no field assertions ride along. Nothing else in the fixture is
// a valid answer, so every other task is forbidden, including 005: it shares
// 001's auth subject matter but is not an SSO bug, so offering it as a
// candidate is a failure to disambiguate rather than a richer answer.
func byKeyword(out string) error {
	return assertFocused(out, "001", "002", "003", "004", "005", "006")
}

// missing expects a plain statement that task 042 does not exist, and no
// fixture task dressed up as it.
func missing(out string) error {
	return assertNotFound(out)
}

// blockedState expects the unmet dependency on 002 *and* a verdict that 003 is
// not startable yet.
func blockedState(out string) error {
	return assertBlocked(out)
}

// jsonFormat expects task 004 as a parseable JSON object — output shape as well
// as content.
func jsonFormat(out string) error {
	return assertJSONTask(out, "004", "readme", "pending")
}
