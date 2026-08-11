package board

import (
	"fmt"
	"sort"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/model"
)

const defaultGroupKey = "(none)"

// GroupResult holds ordered group keys and the grouped tasks map.
type GroupResult struct {
	Keys   []string
	Groups map[string][]*model.Task
}

// GroupTasks groups tasks by the specified field.
//
// efforts supplies the project's effort vocabulary, which sets the column order
// when grouping by effort. Pass effort.Scale{} for the default vocabulary.
//
//nolint:gocognit,gocyclo,funlen // TODO: refactor to reduce complexity
func GroupTasks(tasks []*model.Task, field string, efforts effort.Scale) (*GroupResult, error) {
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
			Keys:   orderedKeys(groups, efforts.Values()),
			Groups: groups,
		}, nil

	case "type":
		for _, t := range tasks {
			key := string(t.Type)
			if key == "" {
				key = defaultGroupKey
			}
			groups[key] = append(groups[key], t)
		}
		return &GroupResult{
			Keys:   orderedKeys(groups, typeOrder()),
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

	case "phase":
		for _, t := range tasks {
			key := t.Phase
			if key == "" {
				key = defaultGroupKey
			}
			groups[key] = append(groups[key], t)
		}
		return &GroupResult{
			Keys:   sortedKeys(groups),
			Groups: groups,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported group-by field: %s (supported: status, priority, effort, type, group, tag, phase)", field)
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
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Status   string   `json:"status"`
	Priority string   `json:"priority,omitempty"`
	Effort   string   `json:"effort,omitempty"`
	Type     string   `json:"type,omitempty"`
	Phase    string   `json:"phase,omitempty"`
	Tags     []string `json:"tags,omitempty"`
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
				Type:     string(t.Type),
				Phase:    t.Phase,
				Tags:     t.Tags,
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

// ReorderKeys reorders the group result keys to match the given order.
// Keys present in order come first (in order), remaining keys follow alphabetically.
// This is useful for phase grouping where the order comes from configuration.
func ReorderKeys(gr *GroupResult, order []string) {
	gr.Keys = orderedKeys(gr.Groups, order)
}

func statusOrder() []string {
	return []string{
		string(model.StatusPending),
		string(model.StatusInProgress),
		string(model.StatusInReview),
		string(model.StatusBlocked),
		string(model.StatusCompleted),
		string(model.StatusCancelled),
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

func typeOrder() []string {
	return []string{
		string(model.TypeFeature),
		string(model.TypeBug),
		string(model.TypeImprovement),
		string(model.TypeChore),
		string(model.TypeDocs),
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
