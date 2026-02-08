package board

import (
	"fmt"
	"sort"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
)

const defaultGroupKey = "(none)"

// GroupResult holds ordered group keys and the grouped tasks map.
type GroupResult struct {
	Keys   []string
	Groups map[string][]*model.Task
}

// GroupTasks groups tasks by the specified field.
//
//nolint:gocognit,gocyclo,funlen // TODO: refactor to reduce complexity
func GroupTasks(tasks []*model.Task, field string) (*GroupResult, error) {
	groups := make(map[string][]*model.Task)

	switch field {
	case "status":
		for _, t := range tasks {
			key := string(t.Status)
			if key == "" {
				key = defaultGroupKey
			}
			groups[key] = append(groups[key], t)
		}
		return &GroupResult{
			Keys:   orderedKeys(groups, statusOrder()),
			Groups: groups,
		}, nil

	case "priority":
		for _, t := range tasks {
			key := string(t.Priority)
			if key == "" {
				key = defaultGroupKey
			}
			groups[key] = append(groups[key], t)
		}
		return &GroupResult{
			Keys:   orderedKeys(groups, priorityOrder()),
			Groups: groups,
		}, nil

	case "effort":
		for _, t := range tasks {
			key := string(t.Effort)
			if key == "" {
				key = defaultGroupKey
			}
			groups[key] = append(groups[key], t)
		}
		return &GroupResult{
			Keys:   orderedKeys(groups, effortOrder()),
			Groups: groups,
		}, nil

	case "group":
		for _, t := range tasks {
			key := t.GetGroup()
			if key == "" {
				key = defaultGroupKey
			}
			groups[key] = append(groups[key], t)
		}
		return &GroupResult{
			Keys:   sortedKeys(groups),
			Groups: groups,
		}, nil

	case "tag":
		for _, t := range tasks {
			if len(t.Tags) == 0 {
				groups[defaultGroupKey] = append(groups[defaultGroupKey], t)
			} else {
				for _, tag := range t.Tags {
					groups[tag] = append(groups[tag], t)
				}
			}
		}
		return &GroupResult{
			Keys:   sortedKeys(groups),
			Groups: groups,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported group-by field: %s (supported: status, priority, effort, group, tag)", field)
	}
}

// JSONGroup is the JSON representation of a board group.
type JSONGroup struct {
	Group string     `json:"group"`
	Count int        `json:"count"`
	Tasks []JSONTask `json:"tasks"`
}

// JSONTask is the JSON representation of a task within a board group.
type JSONTask struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority,omitempty"`
	Effort   string `json:"effort,omitempty"`
}

// ToJSON converts a GroupResult to a JSON-serializable slice.
func ToJSON(gr *GroupResult) []JSONGroup {
	var out []JSONGroup
	for _, key := range gr.Keys {
		tasks := gr.Groups[key]
		jTasks := make([]JSONTask, len(tasks))
		for i, t := range tasks {
			jTasks[i] = JSONTask{
				ID:       t.ID,
				Title:    t.Title,
				Status:   string(t.Status),
				Priority: string(t.Priority),
				Effort:   string(t.Effort),
			}
		}
		out = append(out, JSONGroup{
			Group: key,
			Count: len(tasks),
			Tasks: jTasks,
		})
	}
	return out
}

func statusOrder() []string {
	return []string{
		string(model.StatusPending),
		string(model.StatusInProgress),
		string(model.StatusBlocked),
		string(model.StatusCompleted),
	}
}

func priorityOrder() []string {
	return []string{
		string(model.PriorityCritical),
		string(model.PriorityHigh),
		string(model.PriorityMedium),
		string(model.PriorityLow),
	}
}

func effortOrder() []string {
	return []string{
		string(model.EffortSmall),
		string(model.EffortMedium),
		string(model.EffortLarge),
	}
}

func orderedKeys(groups map[string][]*model.Task, order []string) []string {
	var keys []string
	seen := make(map[string]bool)

	for _, k := range order {
		if _, ok := groups[k]; ok {
			keys = append(keys, k)
			seen[k] = true
		}
	}

	var extra []string
	for k := range groups {
		if !seen[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	keys = append(keys, extra...)

	return keys
}

func sortedKeys(groups map[string][]*model.Task) []string {
	var keys []string
	hasNone := false
	for k := range groups {
		if k == defaultGroupKey {
			hasNone = true
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if hasNone {
		keys = append(keys, defaultGroupKey)
	}
	return keys
}
