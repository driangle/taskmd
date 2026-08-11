package cli

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/effort"
)

// effortConfigKey is the .taskmd.yaml key holding the effort vocabulary.
const effortConfigKey = "effort"

// resolveEffortScale returns the project's effort vocabulary from config,
// falling back to the default when the key is absent or unusable.
//
// Invalid config degrades to the default here rather than failing every command;
// `taskmd validate` is what reports it, via parseEffortConfig and the validator.
func resolveEffortScale() effort.Scale {
	values, errs := parseEffortConfig(viper.Get(effortConfigKey))
	if len(errs) > 0 || len(values) == 0 {
		return effort.Default()
	}

	scale, err := effort.NewScale(values)
	if err != nil {
		return effort.Default()
	}
	return scale
}

// parseEffortConfig converts the raw viper value for the effort key into an
// ordered list of effort values.
//
// It returns the values plus any structural problems — a wrong container type or
// a non-string item — phrased for the user. Nil values with no errors means the
// key was absent. Semantic rules (empty, blank, duplicate) belong to
// effort.NewScale and are reported by the validator, so they are not repeated here.
func parseEffortConfig(raw any) (values []string, structuralErrors []string) {
	if raw == nil {
		return nil, nil
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, []string{fmt.Sprintf(
			"effort must be a list of effort values ordered lowest to highest "+
				"(e.g. `effort: [small, medium, large]`), but found a %s. "+
				"See the Effort section of the spec.", yamlContainerKind(raw))}
	}

	values = make([]string, 0, len(items))
	for i, item := range items {
		s, ok := item.(string)
		if !ok {
			structuralErrors = append(structuralErrors, fmt.Sprintf(
				"effort value at index %d must be a string, but found a %s",
				i, yamlContainerKind(item)))
			continue
		}
		values = append(values, s)
	}

	return values, structuralErrors
}
