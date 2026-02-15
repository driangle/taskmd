package cli

import (
	"github.com/spf13/viper"

	"github.com/driangle/taskmd/apps/cli/internal/taskcontext"
)

// loadScopePathsConfig reads scope definitions from .taskmd.yaml and returns a ScopeMap.
func loadScopePathsConfig() taskcontext.ScopeMap {
	raw := viper.Get("scopes")
	if raw == nil {
		return nil
	}
	scopeMap, ok := raw.(map[string]any)
	if !ok {
		return nil
	}

	entries := parseScopeEntries(scopeMap)
	result := make(taskcontext.ScopeMap, len(entries))
	for name, sc := range entries {
		if len(sc.Paths) > 0 {
			result[name] = sc.Paths
		}
	}
	return result
}
