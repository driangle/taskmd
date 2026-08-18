package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// baselineIDs are the fixture tasks present before the agent runs. Anything
// else in the task list was created by the agent.
var baselineIDs = map[string]bool{
	"001": true, "002": true, "003": true, "004": true, "005": true,
}

// Task mirrors the subset of `taskmd list --format json` we assert on.
type Task struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Priority     string   `json:"priority"`
	Effort       string   `json:"effort"`
	Type         string   `json:"type"`
	Group        string   `json:"group"`
	Tags         []string `json:"tags"`
	Dependencies []string `json:"dependencies"`
	FilePath     string   `json:"file_path"`
}

// listTasks returns every task taskmd can see in the working directory.
//
// `-d tasks` is explicit rather than relying on `.taskmd.yaml`: the bare-project
// variant has no config file, and without the flag taskmd would scan from `.`
// and report root-level tasks as group "tasks". The flag makes every variant
// grade identically.
func listTasks() ([]Task, error) {
	out, err := exec.Command("taskmd", "-d", "tasks", "list", "--format", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("`taskmd -d tasks list --format json` failed: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, fmt.Errorf("could not parse task list: %w", err)
	}
	return tasks, nil
}

// newTask returns the single task the agent added, and errors if the agent
// added none or more than one.
func newTask() (Task, error) {
	tasks, err := listTasks()
	if err != nil {
		return Task{}, err
	}

	var added []Task
	for _, t := range tasks {
		if !baselineIDs[t.ID] {
			added = append(added, t)
		}
	}

	switch len(added) {
	case 1:
		return added[0], nil
	case 0:
		// More files than fixtures but nothing new means the agent reused an
		// existing ID — a real failure, and a confusing one to report as
		// "nothing was created".
		if len(tasks) > len(baselineIDs) {
			return Task{}, fmt.Errorf("created task reuses an existing fixture ID (%d tasks listed, none with a new ID)", len(tasks))
		}
		return Task{}, fmt.Errorf("no new task was created (still only the %d fixture tasks)", len(tasks))
	default:
		return Task{}, fmt.Errorf("expected exactly 1 new task, found %d: %s", len(added), titles(added))
	}
}

// readTaskFile returns the markdown body of a task file on disk.
func readTaskFile(t Task) (string, error) {
	path := filepath.Join("tasks", t.FilePath)
	body, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read created task file %s: %w", path, err)
	}
	return string(body), nil
}

// validates asserts that the whole task set still passes `taskmd validate`.
func validates() error {
	out, err := exec.Command("taskmd", "-d", "tasks", "validate").CombinedOutput()
	if err != nil {
		return fmt.Errorf("`taskmd validate` failed: %s", out)
	}
	return nil
}

func titles(tasks []Task) string {
	out := ""
	for i, t := range tasks {
		if i > 0 {
			out += ", "
		}
		out += t.ID + " " + t.Title
	}
	return out
}
