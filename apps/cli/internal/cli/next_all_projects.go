package cli

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/model"
)

// mergeProjectRecs merges per-project ranked lists into one list, keeping each
// project's own order. At each step it takes the best head across projects
// (earlier project wins ties): by priority tier first when strictPriority is
// set, then by score.
//
// Each list is already ranked by next.Recommend, which may put a lower-scoring
// task first (strict phase or priority tiers). Re-sorting the combined list by
// score would discard that order; merging only the heads preserves it. When no
// strict ordering is active each list is score-sorted, so the result is the
// same as a plain sort by score. Phases are never compared across projects:
// phase names are project-local.
func mergeProjectRecs(perProject [][]ProjectRecommendation, strictPriority bool) []ProjectRecommendation {
	total := 0
	for _, recs := range perProject {
		total += len(recs)
	}

	merged := make([]ProjectRecommendation, 0, total)
	heads := make([]int, len(perProject))
	for len(merged) < total {
		best := -1
		for p, recs := range perProject {
			if heads[p] >= len(recs) {
				continue
			}
			if best == -1 || ranksAbove(recs[heads[p]], perProject[best][heads[best]], strictPriority) {
				best = p
			}
		}
		merged = append(merged, perProject[best][heads[best]])
		heads[best]++
	}
	return merged
}

// ranksAbove reports whether a should be merged ahead of b.
func ranksAbove(a, b ProjectRecommendation, strictPriority bool) bool {
	if strictPriority {
		ta, tb := priorityTier(a.Priority), priorityTier(b.Priority)
		if ta != tb {
			return ta > tb
		}
	}
	return a.Score > b.Score
}

// priorityTier orders priorities for --strict-priority; unset ranks with low,
// matching the tiering next.Recommend applies within a project.
func priorityTier(p string) int {
	switch model.Priority(p) {
	case model.PriorityCritical:
		return 3
	case model.PriorityHigh:
		return 2
	case model.PriorityMedium:
		return 1
	default:
		return 0
	}
}

// loadProjectPhaseOrder returns the phase ids configured in a registered
// project's .taskmd.yaml. Phase names are project-local, so --all-projects
// ranks each project against its own phase order.
func loadProjectPhaseOrder(projectPath string) []string {
	phases := loadProjectPhases(projectPath)
	if len(phases) == 0 {
		return nil
	}
	ids := make([]string, len(phases))
	for i, p := range phases {
		ids[i] = p.ID
	}
	return ids
}

// loadProjectEffortScale returns the effort vocabulary configured in a
// registered project's .taskmd.yaml, or the default when it is absent or
// invalid — the same fallback resolveEffortScale applies to the current project.
func loadProjectEffortScale(projectPath string) effort.Scale {
	data, err := os.ReadFile(filepath.Join(projectPath, configFilename))
	if err != nil {
		return effort.Default()
	}
	var cfg map[string]any
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return effort.Default()
	}
	values, errs := parseEffortConfig(cfg[effortConfigKey])
	if len(errs) > 0 || len(values) == 0 {
		return effort.Default()
	}
	scale, err := effort.NewScale(values)
	if err != nil {
		return effort.Default()
	}
	return scale
}
