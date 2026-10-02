package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/driangle/taskmd/sdk/go/next"
)

// nextConfigKey is the .taskmd.yaml section holding next's settings, and
// nextUnphasedConfigKey the key in it choosing where strict phase ordering
// ranks tasks with no phase.
const (
	nextConfigKey         = "next"
	nextUnphasedConfigKey = "next.unphased"
)

// nextConfigFields lists the keys the next section accepts.
var nextConfigFields = []string{"unphased"}

// resolveUnphasedPlacement reads next.unphased from the current project's
// config. Unlike effort, an invalid value is an error: silently ranking
// differently from what the user configured would be worse than refusing.
func resolveUnphasedPlacement() (next.UnphasedPlacement, error) {
	return parseNextConfig(viper.Get(nextConfigKey))
}

// loadProjectUnphasedPlacement reads next.unphased from a registered project's
// .taskmd.yaml. A missing or unreadable config means the default.
func loadProjectUnphasedPlacement(projectPath string) (next.UnphasedPlacement, error) {
	data, err := os.ReadFile(filepath.Join(projectPath, configFilename))
	if err != nil {
		return next.UnphasedCurrent, nil
	}
	var cfg map[string]any
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return next.UnphasedCurrent, nil
	}
	return parseNextConfig(cfg[nextConfigKey])
}

// parseNextConfig validates the raw next section and returns its unphased
// placement. The section's shape is checked too, so a mistyped setting fails
// loudly instead of silently leaving the default in place. Absent means the
// default, current.
func parseNextConfig(raw any) (next.UnphasedPlacement, error) {
	if raw == nil {
		return next.UnphasedCurrent, nil
	}
	section, ok := raw.(map[string]any)
	if !ok {
		return "", fmt.Errorf("%s: %s must be a mapping (e.g. `next: {unphased: last}`), but found a %s",
			configFilename, nextConfigKey, yamlContainerKind(raw))
	}
	keys := make([]string, 0, len(section))
	for key := range section {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if !slices.Contains(nextConfigFields, key) {
			return "", fmt.Errorf("%s: %w", configFilename,
				invalidValueError("key under "+nextConfigKey, key, nextConfigFields))
		}
	}
	return parseUnphasedPlacement(section["unphased"])
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
