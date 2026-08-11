// Package effort models the project's effort vocabulary: the ordered set of
// values the effort frontmatter field may take.
//
// The vocabulary is ordinal — values are ordered lowest to highest — which is
// what lets filters compare (effort>small), boards and lists sort, and next
// score smaller work more highly. Projects override it via the effort key in
// .taskmd.yaml; when unset the default small, medium, large applies.
package effort

import (
	"fmt"
	"strings"

	"github.com/driangle/taskmd/sdk/go/model"
)

// defaultValues is the built-in effort vocabulary, lowest to highest.
var defaultValues = []string{
	string(model.EffortSmall),
	string(model.EffortMedium),
	string(model.EffortLarge),
}

// Scale is an ordered effort vocabulary, lowest value first.
//
// The zero value is usable and behaves as the default vocabulary, so callers
// with no configuration can pass Scale{} safely.
type Scale struct {
	values []string
}

// Default returns the built-in vocabulary: small, medium, large.
func Default() Scale {
	return Scale{}
}

// NewScale returns a Scale over values, ordered lowest to highest.
//
// It rejects an empty list, blank entries, and duplicates — each of which would
// otherwise produce a vocabulary that cannot be validated or ordered sensibly.
func NewScale(values []string) (Scale, error) {
	if len(values) == 0 {
		return Scale{}, fmt.Errorf("effort must list at least one value")
	}

	seen := make(map[string]bool, len(values))
	cleaned := make([]string, 0, len(values))
	for i, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return Scale{}, fmt.Errorf("effort value at index %d is empty", i)
		}
		if seen[trimmed] {
			return Scale{}, fmt.Errorf("duplicate effort value: %q", trimmed)
		}
		seen[trimmed] = true
		cleaned = append(cleaned, trimmed)
	}

	return Scale{values: cleaned}, nil
}

// Values returns the vocabulary, lowest to highest.
func (s Scale) Values() []string {
	if len(s.values) == 0 {
		return defaultValues
	}
	return s.values
}

// Contains reports whether v is part of the vocabulary.
func (s Scale) Contains(v string) bool {
	return s.Rank(v) >= 0
}

// Rank returns the position of v in the vocabulary (0 is lowest), or -1 when v
// is not part of it.
func (s Scale) Rank(v string) int {
	for i, val := range s.Values() {
		if val == v {
			return i
		}
	}
	return -1
}

// Lowest returns the lowest value in the vocabulary — the one that counts as a
// quick win.
func (s Scale) Lowest() string {
	return s.Values()[0]
}

// Highest returns the highest value in the vocabulary.
func (s Scale) Highest() string {
	vals := s.Values()
	return vals[len(vals)-1]
}

// Len returns the number of values in the vocabulary.
func (s Scale) Len() int {
	return len(s.Values())
}

// String renders the vocabulary as a comma-separated list, for use in error
// messages and help text.
func (s Scale) String() string {
	return strings.Join(s.Values(), ", ")
}

// Points scores v proportionally to its position: the lowest value earns
// maxPoints, the highest earns none, and the values between are spread evenly.
// For the default three-value vocabulary and maxPoints=5 this yields 5, 2, 0.
//
// The bool is false when v contributes no points — because it is unset, unknown
// to the vocabulary, or the highest value.
func (s Scale) Points(v string, maxPoints int) (int, bool) {
	rank := s.Rank(v)
	if rank < 0 {
		return 0, false
	}

	// A single-value vocabulary carries no ordering information, so its one
	// value is treated as the lowest and earns full points.
	steps := s.Len() - 1
	if steps == 0 {
		return maxPoints, true
	}

	points := maxPoints * (steps - rank) / steps
	if points == 0 {
		return 0, false
	}
	return points, true
}
