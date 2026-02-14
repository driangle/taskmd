package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFileName = ".taskmd-sync.yaml"

// SyncConfig is the top-level sync configuration.
type SyncConfig struct {
	Sources []SourceConfig `yaml:"sources"`
}

// SourceConfig describes a single external source.
type SourceConfig struct {
	Name      string            `yaml:"name"`
	Project   string            `yaml:"project"`
	BaseURL   string            `yaml:"base_url"`
	TokenEnv  string            `yaml:"token_env"`
	UserEnv   string            `yaml:"user_env"`
	OutputDir string            `yaml:"output_dir"`
	FieldMap  FieldMap          `yaml:"field_map"`
	Filters   map[string]any    `yaml:"filters"`
	Extra     map[string]string `yaml:"extra"`
}

// FieldMap configures how external fields are mapped to taskmd frontmatter.
type FieldMap struct {
	Status          map[string]string `yaml:"status"`
	Priority        map[string]string `yaml:"priority"`
	LabelsToTags    bool              `yaml:"labels_to_tags"`
	AssigneeToOwner bool              `yaml:"assignee_to_owner"`
}

// LoadConfig reads the sync config from dir/.taskmd-sync.yaml.
func LoadConfig(dir string) (*SyncConfig, error) {
	path := filepath.Join(dir, configFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read sync config: %w", err)
	}

	var cfg SyncConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse sync config: %w", err)
	}

	if len(cfg.Sources) == 0 {
		return nil, fmt.Errorf("sync config has no sources defined")
	}

	for i, s := range cfg.Sources {
		if s.Name == "" {
			return nil, fmt.Errorf("source at index %d has no name", i)
		}
		if s.OutputDir == "" {
			return nil, fmt.Errorf("source %q has no output_dir", s.Name)
		}
	}

	return &cfg, nil
}
