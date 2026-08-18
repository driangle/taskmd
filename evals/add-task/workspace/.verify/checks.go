package main

import "errors"

// checks maps a check name (referenced from suite.yaml) to its implementation.
var checks = map[string]func() error{
	"add-basic":         checkAddBasic,
	"add-bug-template":  checkAddBugTemplate,
	"add-dependency":    checkAddDependency,
	"add-group-routing": checkAddGroupRouting,
}

// checkAddBasic verifies the plain "create a task with priority and tags"
// request: the right metadata, and real content instead of template stubs.
func checkAddBasic() error {
	task, err := newTask()
	if err != nil {
		return err
	}

	body, err := readTaskFile(task)
	if err != nil {
		return err
	}

	return errors.Join(
		titleMentions(task, "notification"),
		equals("priority", task.Priority, "high"),
		hasTags(task, "notifications", "backend"),
		noPlaceholders(body),
		summaryFilled(body, 40),
		hasSubtasks(body, 2),
		sectionFilled(body, "Acceptance Criteria", 40),
		validates(),
	)
}

// checkAddGroupRouting verifies that a task lands in the group directory its
// domain belongs to. The fixture already uses `cli/` and `web/`, so the
// convention is discoverable by reading the project — this does not reward
// hardcoding taskmd's own taxonomy.
func checkAddGroupRouting() error {
	task, err := newTask()
	if err != nil {
		return err
	}

	body, err := readTaskFile(task)
	if err != nil {
		return err
	}

	return errors.Join(
		titleMentions(task, "dark mode", "dark-mode", "theme"),
		equals("group", task.Group, "web"),
		noPlaceholders(body),
		validates(),
	)
}

// checkAddBugTemplate verifies that a bug report is filed with the bug
// template and that the template's own sections are filled in rather than left
// as stubs. Group routing is deliberately not asserted here — it has its own
// eval, so one behavior can't mask the other.
func checkAddBugTemplate() error {
	task, err := newTask()
	if err != nil {
		return err
	}

	body, err := readTaskFile(task)
	if err != nil {
		return err
	}

	return errors.Join(
		titleMentions(task, "search", "crash", "empty query"),
		equals("type", task.Type, "bug"),
		noPlaceholders(body),
		sectionFilled(body, "Steps to Reproduce", 30),
		sectionFilled(body, "Expected Behavior", 20),
		sectionFilled(body, "Actual Behavior", 20),
		validates(),
	)
}

// checkAddDependency verifies that a stated ordering constraint becomes a real
// `dependencies` entry rather than prose in the body.
func checkAddDependency() error {
	task, err := newTask()
	if err != nil {
		return err
	}

	body, err := readTaskFile(task)
	if err != nil {
		return err
	}

	return errors.Join(
		titleMentions(task, "doc", "documentation"),
		equals("priority", task.Priority, "low"),
		hasDependency(task, "002"),
		noPlaceholders(body),
		summaryFilled(body, 40),
		validates(),
	)
}
