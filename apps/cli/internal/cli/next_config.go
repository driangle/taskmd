package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/driangle/taskmd/sdk/go/next"
)

// nextUnphasedConfigKey is the .taskmd.yaml key choosing where strict phase
// ordering ranks tasks with no phase.
const nextUnphasedConfigKey = "next.unphased"

// projectNextConfig reads just the next settings from a project's .taskmd.yaml.
type projectNextConfig struct {
	Next struct {
		Unphased any `yaml:"unphased"`
	} `yaml:"next"`
}

// resolveUnphasedPlacement reads next.unphased from the current project's
// config. Unlike effort, an invalid value is an error: silently ranking
// differently from what the user configured would be worse than refusing.
func resolveUnphasedPlacement() (next.UnphasedPlacement, error) {
	return parseUnphasedPlacement(viper.Get(nextUnphasedConfigKey))
}

// loadProjectUnphasedPlacement reads next.unphased from a registered project's
// .taskmd.yaml. A missing or unreadable config means the default.
func loadProjectUnphasedPlacement(projectPath string) (next.UnphasedPlacement, error) {
	data, err := os.ReadFile(filepath.Join(projectPath, configFilename))
	if err != nil {
		return next.UnphasedCurrent, nil
	}
	var cfg projectNextConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return next.UnphasedCurrent, nil
	}
	return parseUnphasedPlacement(cfg.Next.Unphased)
}

// parseUnphasedPlacement converts a raw next.unphased value. Absent means the
// default, current.
func parseUnphasedPlacement(raw any) (next.UnphasedPlacement, error) {
	if raw == nil {
		return next.UnphasedCurrent, nil
	}
	valid := []string{string(next.UnphasedCurrent), string(next.UnphasedLast)}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%s: invalid %s: must be %s or %s, but found a %s", configFilename,
			nextUnphasedConfigKey, next.UnphasedCurrent, next.UnphasedLast, yamlContainerKind(raw))
	}
	switch next.UnphasedPlacement(value) {
	case next.UnphasedCurrent, next.UnphasedLast:
		return next.UnphasedPlacement(value), nil
	default:
		return "", fmt.Errorf("%s: %w", configFilename, invalidValueError(nextUnphasedConfigKey, value, valid))
	}
}
