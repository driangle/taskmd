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
	ID           string    `yaml:"id"`
	Title        string    `yaml:"title"`
	Status       Status    `yaml:"status"`
	Priority     Priority  `yaml:"priority"`
	Effort       Effort    `yaml:"effort"`
	Dependencies []string  `yaml:"dependencies"`
	Tags         []string  `yaml:"tags"`
	Group        string    `yaml:"group"` // From frontmatter or derived from directory
	Created      time.Time `yaml:"created"`

	// Content fields
	Body     string // Markdown body content
	FilePath string // Source file path
}

// IsValid checks if the task has required fields
func (t *Task) IsValid() bool {
	return t.ID != "" && t.Title != ""
}

// GetGroup returns the group, prioritizing frontmatter over derived value
func (t *Task) GetGroup() string {
	return t.Group
}
