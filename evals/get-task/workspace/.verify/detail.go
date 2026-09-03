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
//
// The list is broad because "this does not exist" has many idioms and the eval
// grades the fact, not the phrasing. The 2026-09-03 run failed a correct
// answer — "I don't see a task 042 in this project" — on a list that had
// "no task" but not "don't see".
var notFoundCues = regexp.MustCompile(`(?i)` + strings.Join([]string{
	`no (?:such )?task`, `no match`, `no exact match`, `not found`,
	`there ?(?:'s|s| is| are) no`,
	`do(?:es)? ?n[o']?t (?:see|have|exist|appear|contain)`,
	`is ?n[o']?t (?:a |any )?(?:task|there|present)`,
	`could ?n[o']?t find`, `can ?n[o']?t find`, `unable to find`,
	`nonexistent`, `non-existent`,
}, "|"))

// missingIDPattern matches the queried-but-absent ID as a standalone token. The
// digits are captured so callers can address the ID's own span: the
// surrounding boundary characters are part of the match, and treating the match
// end as the ID's end steps *over* a sentence-ending period — which is how the
// 2026-09-03 run scored "There's no task 042. The closest match is 002: Add
// full-text search" as a mislabel.
var missingIDPattern = regexp.MustCompile(`(?:^|[^0-9])(042)(?:[^0-9]|$)`)

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
		stated    bool
		mentioned bool
	)

	for i, line := range lines {
		if !missingIDPattern.MatchString(line) {
			continue
		}
		mentioned = true
		if notFoundCues.MatchString(window(lines, i)) {
			stated = true
		}
	}

	mislabeled := mislabeledAs042(output)
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

// mislabelWindow is how much text either side of an occurrence of 042 counts as
// "attached to" it. Sentence-scoped, not line-scoped: agents answer this eval in
// a single paragraph, and a whole-line check called this correct answer a
// mislabel —
//
//	There's no task 042. Available task IDs are 001–006. Closest match was
//	task 002 ("Add full-text search"), but none exactly matches "042"
//
// — because the title happens to sit on the same (only) line. Naming a
// near-match in a *later sentence*, having already said the task does not
// exist, is the behavior the skills ask for.
// The windows are asymmetric because mislabeling has a direction: it reads
// "Task 042: Add full-text search", with the title *after* the ID. The backward
// window only needs to cover the rarer "Add full-text search (042)" form, and
// keeping it short matters — the answer above mentions 042 a second time, and a
// 60-character look-back from that one reaches into the preceding clause and
// finds 002's title there.
const (
	mislabelAfter  = 60
	mislabelBefore = 25
)

// mislabeledAs042 returns the fixture tasks whose title is attached directly to
// the missing ID — "Task 042: Add full-text search" — which is the failure the
// CLI's 67% fuzzy match on 002 invites. The span around each 042 is cut at the
// first sentence boundary, so a title mentioned in a neighbouring sentence does
// not count.
func mislabeledAs042(output string) []string {
	lower := strings.ToLower(output)

	var found []string
	seen := map[string]bool{}
	for _, loc := range missingIDPattern.FindAllStringSubmatchIndex(lower, -1) {
		start, end := loc[2], loc[3] // the captured digits, not the boundaries

		span := clipForward(lower[end:min(end+mislabelAfter, len(lower))]) +
			"\n" + clipBackward(lower[max(start-mislabelBefore, 0):start])

		for id, tokens := range titleTokens {
			if !seen[id] && containsToken(span, tokens) {
				seen[id] = true
				found = append(found, id)
			}
		}
	}
	return found
}

// spanBoundary ends the text that counts as attached to the missing ID: a
// sentence or clause break, or — just as decisive — *another task ID*. If a
// real ID sits between 042 and a title, the title belongs to that ID. Without
// this, "There's no task 042 — the highest task ID present is 006
// (tasks/web/006-export-reports-csv.md)" reads as 006 mislabeled as 042.
var spanBoundary = regexp.MustCompile(`[.!?\n;—–]|00[1-6]`)

func clipForward(s string) string {
	if loc := spanBoundary.FindStringIndex(s); loc != nil {
		return s[:loc[0]]
	}
	return s
}

func clipBackward(s string) string {
	if locs := spanBoundary.FindAllStringIndex(s, -1); len(locs) > 0 {
		return s[locs[len(locs)-1][1]:]
	}
	return s
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
