package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFileName = ".taskmd.yaml"

// Config is the top-level project configuration.
type Config struct {
	Sync SyncConfig `yaml:"sync"`
}

// SyncConfig holds the sync-specific configuration.
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

// LoadConfig reads the project config from dir/.taskmd.yaml and returns the sync section.
func LoadConfig(dir string) (*SyncConfig, error) {
	path := filepath.Join(dir, configFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", configFileName, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", configFileName, err)
	}

	syncCfg := &cfg.Sync

	if len(syncCfg.Sources) == 0 {
		return nil, fmt.Errorf("no sync sources defined in %s", configFileName)
	}

	for i, s := range syncCfg.Sources {
		if s.Name == "" {
			return nil, fmt.Errorf("source at index %d has no name", i)
		}
		if s.OutputDir == "" {
			return nil, fmt.Errorf("source %q has no output_dir", s.Name)
		}
	}

	return syncCfg, nil
}
