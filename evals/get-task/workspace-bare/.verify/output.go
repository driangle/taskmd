package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

// This file grades the agent's *reported output*, which is what a read-only
// skill produces. skival's `check_output` verifier pipes the agent's final text
// to stdin; everything here reads from there.
//
// Two rules from evals/README.md shape the matching:
//
//   - Assert two-sided. The expected IDs must be reported *and* the competing
//     ones must not, or an agent that dumps every task passes a "pending only"
//     eval.
//   - Match stable tokens, not prose. IDs and title words are matchable;
//     sentence phrasing and table layout are not. Grading the plugin skill's
//     pretty-printed format would measure formatting compliance and unfairly
//     fail `no-skill`.

// titleTokens maps each fixture task to distinctive lowercase fragments of its
// title or filename. A bare ID is not enough evidence that a task was reported:
// "001" also shows up in prose like "001 is in-progress, so I left it out".
// Requiring a title token near the ID keeps those mentions from counting.
//
// 005's token is "refactor", not the "auth" the list-tasks suite uses. 001's
// body reads "users authenticate via SSO", and a substring match on "auth"
// fires on *authenticate* — so a correct, 001-only answer to the keyword eval
// would be scored as having leaked 005. That collision is live in this suite in
// a way it was not in list-tasks: get-by-keyword deliberately asks an auth
// question in an auth-heavy fixture.
var titleTokens = map[string][]string{
	"001": {"sso", "login"},
	"002": {"search"},
	"003": {"xss", "vulnerab", "security"},
	"004": {"readme"},
	"005": {"refactor"},
	"006": {"csv", "export"},
}

// idPattern matches a fixture ID as a standalone token, so it does not fire on
// digits inside dates ("2026-03-08") or counts ("6 tasks").
var idPattern = regexp.MustCompile(`(?:^|[^0-9])(00[1-6])(?:[^0-9]|$)`)

// listItemPattern matches a line that *begins* with an ID, allowing for the
// markup agents wrap it in: table pipes, bullets, bold, headings, numbering.
// Starting the line with the ID is the structural signal that separates a
// listing row from a sentence that happens to mention a task.
//
// The optional `group/` prefix is not decoration — agents really do label rows
// `web/003` and `cli/001` when the fixture has task groups, and without it those
// rows read as prose and drop out of the results.
var listItemPattern = regexp.MustCompile("^\\s*(?:\\d{1,2}[.)]\\s*)?[\\s\\-*•#>|()\\[\\]`]*\\**\\s*(?:[a-z][a-z0-9_.-]*/)*(00[1-6])(?:[^0-9]|$)")

// exclusionCues mark text as commentary about a task the agent deliberately
// left out, rather than as a line reporting it. Without this, an agent that
// correctly filters *and* explains itself — "005 (Refactor authentication
// module) is completed, so it's excluded" — would be scored as having leaked
// 005 into its results.
var exclusionCues = regexp.MustCompile(`(?i)exclud|omitt|left out|filtered out|not (?:shown|included|listed|returned)|skipp|does ?n[o']t match`)

// readOutput returns the agent's final text from stdin.
func readOutput() (string, error) {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("could not read agent output from stdin: %w", err)
	}
	if strings.TrimSpace(string(raw)) == "" {
		return "", fmt.Errorf("agent produced no output")
	}
	return string(raw), nil
}

// reportedIDs returns the fixture tasks the agent actually reported.
//
// A task counts as reported when its ID and one of its title tokens appear
// within one line of each other. The one-line window covers the block layouts
// agents use ("### 001" followed by "Title: Fix login SSO bug") without letting
// a bare ID elsewhere in the answer count as a listing.
//
// Structure decides which lines are eligible at all, because agents routinely
// end a correct filtered answer with a sentence naming what they left out —
// "The other two tasks (006 export CSV, 004 update README) are in the Polish
// phase, not MVP" — and those names must not count as results:
//
//   - If the answer contains any list items (table rows, bullets, headings —
//     anything starting with the ID), then *only* those lines count. The agent
//     chose a list format, so its prose is commentary about the list.
//   - If it contains none, the answer is prose, and the window match is used
//     instead. Exclusion cues then veto a match, since prose is where an agent
//     explains itself.
//
// Trying to catch the commentary case with cue words alone does not work: the
// sentence above says "not MVP", the next agent will phrase it another way, and
// evals/README.md is explicit that graders match stable tokens, not prose.
func reportedIDs(output string) map[string]bool {
	lines := strings.Split(strings.ToLower(output), "\n")

	// Two list items, not one: wrapped prose can put a single ID at the start of
	// a continuation line by accident ("…, medium),\n003 (Patch XSS…"), and
	// treating that as a list would discard the rest of the sentence. Every
	// expected result set in this suite has at least two tasks.
	listItems := 0
	for _, line := range lines {
		if listItemPattern.MatchString(line) {
			listItems++
		}
	}
	listFormatted := listItems >= 2

	found := map[string]bool{}
	for i, line := range lines {
		isListItem := listItemPattern.MatchString(line)
		if listFormatted && !isListItem {
			continue
		}

		for _, m := range idPattern.FindAllStringSubmatch(line, -1) {
			id := m[1]
			if found[id] {
				continue
			}

			win := window(lines, i)
			if !containsToken(win, titleTokens[id]) {
				continue
			}

			// A list item is vetoed only by a cue on its own line; prose is
			// vetoed by a cue anywhere in its window, which is where the
			// multi-line "…is completed, so it's excluded" case lands.
			cueScope := win
			if isListItem {
				cueScope = line
			}
			if exclusionCues.MatchString(cueScope) {
				continue
			}

			found[id] = true
		}
	}
	return found
}

// window returns line i together with its immediate neighbours, joined.
func window(lines []string, i int) string {
	lo, hi := i-1, i+1
	if lo < 0 {
		lo = 0
	}
	if hi >= len(lines) {
		hi = len(lines) - 1
	}
	return strings.Join(lines[lo:hi+1], "\n")
}

func containsToken(haystack string, tokens []string) bool {
	for _, tok := range tokens {
		if strings.Contains(haystack, tok) {
			return true
		}
	}
	return false
}

// The exact-set assertion `list-tasks` uses lives in detail.go here instead, as
// `assertFocused`: get-task grades one task as the answer, and a strict
// "these IDs and no others" rule would fail a correct answer to get-by-id,
// which legitimately names 003's dependency 002.

func sortedSet(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
