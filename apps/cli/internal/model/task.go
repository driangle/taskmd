package model

import (
	"fmt"
	"time"
)

// Status represents the current state of a task
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in-progress"
	StatusCompleted  Status = "completed"
	StatusBlocked    Status = "blocked"
	StatusCancelled  Status = "cancelled"
)

// Priority represents the importance level of a task
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// Effort represents the estimated effort required
type Effort string

const (
	EffortSmall  Effort = "small"
	EffortMedium Effort = "medium"
	EffortLarge  Effort = "large"
)

// VerifyStep represents a single verification check defined in task frontmatter.
type VerifyStep struct {
	Type  string `yaml:"type" json:"type"`
	Run   string `yaml:"run,omitempty" json:"run,omitempty"`
	Dir   string `yaml:"dir,omitempty" json:"dir,omitempty"`
	Check string `yaml:"check,omitempty" json:"check,omitempty"`
}

// ValidateVerifySteps checks that each step has a type and the required fields for that type.
// Returns a list of human-readable validation errors.
func ValidateVerifySteps(steps []VerifyStep) []string {
	var errs []string
	for i, s := range steps {
		if s.Type == "" {
			errs = append(errs, fmt.Sprintf("verify[%d]: missing required field 'type'", i))
			continue
		}
		switch s.Type {
		case "bash":
			if s.Run == "" {
				errs = append(errs, fmt.Sprintf("verify[%d]: bash step missing required field 'run'", i))
			}
		case "assert":
			if s.Check == "" {
				errs = append(errs, fmt.Sprintf("verify[%d]: assert step missing required field 'check'", i))
			}
		}
	}
	return errs
}

// Task represents a parsed task from a markdown file
type Task struct {
	// Frontmatter fields
	ID           string       `yaml:"id" json:"id"`
	Title        string       `yaml:"title" json:"title"`
	Status       Status       `yaml:"status" json:"status"`
	Priority     Priority     `yaml:"priority" json:"priority,omitempty"`
	Effort       Effort       `yaml:"effort" json:"effort,omitempty"`
	Dependencies []string     `yaml:"dependencies" json:"dependencies"`
	Tags         []string     `yaml:"tags" json:"tags"`
	Touches      []string     `yaml:"touches" json:"touches,omitempty"`
	Context      []string     `yaml:"context" json:"context,omitempty"`
	Group        string       `yaml:"group" json:"group,omitempty"`
	Owner        string       `yaml:"owner" json:"owner,omitempty"`
	Parent       string       `yaml:"parent,omitempty" json:"parent,omitempty"`
	Created      time.Time    `yaml:"created" json:"created"`
	Verify       []VerifyStep `yaml:"verify,omitempty" json:"verify,omitempty"`

	// Content fields
	Body     string `json:"-"`
	FilePath string `json:"file_path"`
}

// IsValid checks if the task has required fields
func (t *Task) IsValid() bool {
	return t.ID != "" && t.Title != ""
}

// GetGroup returns the group, prioritizing frontmatter over derived value
func (t *Task) GetGroup() string {
	return t.Group
}
