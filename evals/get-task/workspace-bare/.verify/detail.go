package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Graders for a *single-task* answer.
//
// list-tasks grades a set: "exactly these IDs and no others". get-task grades a
// focus: one task is the answer, and the fixture's other tasks must not be
// presented as if they were. The difference matters in one specific place —
// 003 legitimately names 002 as its dependency, so an exact-set assertion would
// fail correct answers. Hence `assertFocused`, which takes an explicit
// forbidden list rather than deriving it from the fixture.
//
// Everything here reuses `reportedIDs` from output.go, which requires an ID and
// a title token within one line of each other. That is what keeps a passing
// mention of the word "authentication" from counting as a report of task 005.

// assertFocused asserts the agent reported `want` and did not report any of
// `forbidden`.
func assertFocused(output, want string, forbidden ...string) error {
	got := reportedIDs(output)

	if !got[want] {
		return fmt.Errorf("did not report task %s (reported %v)", want, sortedSet(got))
	}

	var leaked []string
	for _, id := range forbidden {
		if got[id] {
			leaked = append(leaked, id)
		}
	}
	sort.Strings(leaked)
	if len(leaked) > 0 {
		return fmt.Errorf("reported task %s but also presented %v", want, leaked)
	}
	return nil
}

// assertFields asserts each expected frontmatter *value* appears in the answer,
// case-insensitively, as a standalone word.
//
// Presence-only, deliberately. The obvious stronger check — "no contradictory
// value on any line naming this field" — misfires on correct answers: an agent
// describing 003 may also print its dependency 002's own status and priority,
// and those lines are not fabrication. Two-sidedness for these evals comes from
// `assertFocused` instead, which is where the real "close enough, wrong task"
// failure shows up.
func assertFields(output string, values ...string) error {
	lower := strings.ToLower(output)

	var missing []string
	for _, v := range values {
		if !wordPattern(v).MatchString(lower) {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("answer does not state %v", missing)
	}
	return nil
}

// wordPattern matches a value as a standalone token, so "bug" does not fire on
// "debugging" and "pending" does not fire on "depending on 002".
//
// The boundary excludes letters and digits but *allows* a hyphen: agents write
// "a critical-priority bug", and requiring a non-hyphen boundary failed that
// answer. Nothing is lost — a hyphen never joins a status value to a word that
// would make the match wrong.
func wordPattern(v string) *regexp.Regexp {
	return regexp.MustCompile(`(?:^|[^a-z0-9])` + regexp.QuoteMeta(strings.ToLower(v)) + `(?:[^a-z0-9]|$)`)
}

// notFoundCues are ways of saying a task does not exist. Matched near the
// queried ID rather than anywhere in the answer, so "not found" text about
// something else cannot satisfy the eval.
var notFoundCues = regexp.MustCompile(`(?i)no (?:such )?task|not found|no exact match|does ?n[o']?t exist|no match|could ?n[o']?t find|can ?n[o']?t find|unable to find|there is no|isn'?t (?:a|any) task`)

// missingIDPattern matches the queried-but-absent ID as a standalone token.
var missingIDPattern = regexp.MustCompile(`(?:^|[^0-9])042(?:[^0-9]|$)`)

// assertNotFound grades the get-missing eval.
//
// Two sides:
//
//   - The answer must say, near the ID it was asked about, that no such task
//     exists. An agent that silently answers about the nearest match has no cue
//     and fails here.
//   - No line may present a fixture task's title as though it belonged to 042.
//     `taskmd get 042` fuzzy-matches 002 at 67% and offers it, so "reported the
//     near-match as if it were the one asked for" is a real, reachable failure
//     rather than a hypothetical one.
//
// Reported IDs are deliberately *not* forbidden: both skills tell the agent to
// list the available tasks when a lookup misses, so a correct answer names all
// six.
func assertNotFound(output string) error {
	lines := strings.Split(output, "\n")

	var (
		stated     bool
		mentioned  bool
		mislabeled []string
	)

	for i, line := range lines {
		if !missingIDPattern.MatchString(line) {
			continue
		}
		mentioned = true

		if notFoundCues.MatchString(window(lines, i)) {
			stated = true
		}
		for id, tokens := range titleTokens {
			if containsToken(strings.ToLower(line), tokens) {
				mislabeled = append(mislabeled, id)
			}
		}
	}

	sort.Strings(mislabeled)
	switch {
	case !mentioned:
		return fmt.Errorf("answer never mentions the task that was asked for (042)")
	case len(mislabeled) > 0:
		return fmt.Errorf("presented task %v as if it were 042", mislabeled)
	case !stated:
		return fmt.Errorf("answer mentions 042 but never states that no such task exists")
	}
	return nil
}

// dependencyCues mark text that ties an ID to a blocking relationship, as
// opposed to merely listing it.
var dependencyCues = regexp.MustCompile(`(?i)depend|block|prerequisite|require|unblock|waiting on|needs? 002|before`)

// notStartableCues mark a verdict that the task cannot be started yet. This is
// the discriminating half of the blocked-state eval: an agent that pastes
// `taskmd get 003` verbatim already satisfies the dependency half from the
// CLI's own "Depends on: 002" line without ever answering the question.
//
// The ordering-constraint alternations ("…has to be completed first") are
// deliberately loose about the words in between, because that is how agents
// phrase a soft no. They stay safe by anchoring on a completion word: the
// verbatim CLI dump contains "before rendering in the DOM" and "not properly
// sanitized", and matches neither. output_test.go pins that.
var notStartableCues = regexp.MustCompile(`(?i)` + strings.Join([]string{
	`block`,
	`is ?n[o']?t ready`, `not ready`, `not yet`,
	`can(?:no|['’])?t (?:be )?start`, `can not start`,
	`should(?:n['’]?|n)?t start`,
	`not startable`,
	`unmet`, `unsatisfied`, `incomplete dependenc`,
	`(?:complet|finish|done|do)[a-z]*[^.\n]{0,30}first`,
	`first[^.\n]{0,30}(?:complet|finish|done)`,
	`(?:has|have|needs?|must)[^.\n]{0,15}(?:be )?(?:completed|finished|done)`,
	`before (?:you|start|it can|that)`,
}, "|"))

// assertBlocked grades the get-blocked-state eval: the answer must name the
// unmet dependency on 002 *and* say 003 is not startable because of it.
func assertBlocked(output string) error {
	lines := strings.Split(output, "\n")

	dependency := false
	for i, line := range lines {
		if !regexp.MustCompile(`(?:^|[^0-9])002(?:[^0-9]|$)`).MatchString(line) {
			continue
		}
		if dependencyCues.MatchString(window(lines, i)) {
			dependency = true
			break
		}
	}

	verdict := notStartableCues.MatchString(output)

	switch {
	case !dependency && !verdict:
		return fmt.Errorf("answer names neither the dependency on 002 nor that 003 is blocked")
	case !dependency:
		return fmt.Errorf("answer says the task is blocked but never names the dependency on 002")
	case !verdict:
		return fmt.Errorf("answer mentions 002 but never says 003 cannot be started yet")
	}
	return nil
}
