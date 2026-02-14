package model

import (
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

// Task represents a parsed task from a markdown file
type Task struct {
	// Frontmatter fields
	ID           string    `yaml:"id" json:"id"`
	Title        string    `yaml:"title" json:"title"`
	Status       Status    `yaml:"status" json:"status"`
	Priority     Priority  `yaml:"priority" json:"priority,omitempty"`
	Effort       Effort    `yaml:"effort" json:"effort,omitempty"`
	Dependencies []string  `yaml:"dependencies" json:"dependencies"`
	Tags         []string  `yaml:"tags" json:"tags"`
	Touches      []string  `yaml:"touches" json:"touches,omitempty"`
	Group        string    `yaml:"group" json:"group,omitempty"`
	Owner        string    `yaml:"owner" json:"owner,omitempty"`
	Parent       string    `yaml:"parent,omitempty" json:"parent,omitempty"`
	Created      time.Time `yaml:"created" json:"created"`

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
