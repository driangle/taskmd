package sync

import "time"

// ExternalTask represents a task fetched from an external source.
type ExternalTask struct {
	ExternalID  string
	Title       string
	Description string
	Status      string
	Priority    string
	Assignee    string
	Labels      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	URL         string
	Extra       map[string]string
}

// Source defines the interface that external task providers must implement.
type Source interface {
	Name() string
	ValidateConfig(cfg SourceConfig) error
	FetchTasks(cfg SourceConfig) ([]ExternalTask, error)
}
