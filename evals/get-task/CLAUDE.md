# get-task suite

See [../CLAUDE.md](../CLAUDE.md) for the rules that apply to every suite, and
[../list-tasks/CLAUDE.md](../list-tasks/CLAUDE.md) for the read-only grading model this
suite inherits (`check_output` on stdin, plus `no-mutation`, plus `tool_not_used`).

## One task is the answer, not a set

`list-tasks` grades an exact ID set. `get-task` grades a **focus**: `assertFocused` in
`workspace/.verify/detail.go` takes an explicit forbidden list rather than deriving it
from the fixture. That is not laziness — 003 legitimately names its dependency 002, so an
exact-set assertion would fail correct answers to `get-by-id`.

## Three traps specific to this fixture and CLI

- **`auth` matches `authenticate`.** `titleTokens["005"]` is `{"refactor"}` here, where
  `list-tasks` uses `{"auth"}`. 001's body reads *"users authenticate via SSO"*, so the
  `auth` token would fire on a correct, 001-only answer to `get-by-keyword` and score it
  as having leaked 005. Pinned by `TestByKeyword/full_body_mentioning_authentication`.
- **`taskmd get` is interactive on a fuzzy match.** `taskmd get 042` matches 002 at 67%,
  opens a selection prompt, and with no tty exits 1 with `Error: invalid selection`. Same
  for `taskmd get sso`. `--exact` is the non-interactive form, and **neither skill
  mentions it** — that is a finding, not a bug in the suite. It is also why `get-missing`
  must assert the near-match was not reported as the answer.
- **`get --format json` emits an object, `list --format json` emits an array.**
  `taskObjects` accepts both, and jumps past a parsed value so the nested
  `dependencies.depends_on[0]` object — which carries the *dependency's* id — is not
  counted as a second task. Pinned by `TestTaskObjectsSkipsNestedDependency`.

## `get-missing` has been the buggy grader every time

All five grader bugs found so far — two by the Go tests, one by the smoke run, two by auditing
the full run — were in the not-found path, and every one was a **false negative**: the agent
was right and the grader's notion of "the title is attached to 042" was too loose, or its
vocabulary for "does not exist" too narrow. If you touch `assertNotFound` or `mislabeledAs042`,
assume the same and add cases in both directions.

Two rules that are easy to undo there:

- **The span around 042 breaks on another task ID, not just on punctuation.** If a real ID
  sits between 042 and a title, the title belongs to that ID — *"no task 042 — the highest
  task ID present is 006 (…006-export-reports-csv.md)"* is correct, not a mislabel.
- **Address the captured digits, never the match bounds.** `missingIDPattern` consumes a
  boundary character on each side; measuring from the match end steps over a sentence-ending
  period and swallows the following sentence.

## After touching a grader

```bash
cd workspace/.verify && GOWORK=off go test ./...   # both directions, zero tokens
cp -R workspace/.verify/. workspace-bare/.verify/  # the two must stay identical
```

`no-mutation` is filesystem-backed and not covered by the Go tests — hand-test it from a
copy **outside this repo** (`taskmd list` inside the checkout walks up to taskmd's own
task set).

## Deliberate grading judgment calls

Record these in any report; they are choices, not accidents.

- **`assertFields` is presence-only.** The stronger "no contradictory value on a line
  naming this field" rule misfires when an agent correctly prints 003's dependency 002
  along with 002's own status and priority. Two-sidedness comes from `assertFocused`.
- **`get-by-keyword` fails an answer that offers 001 *and* 005.** The prompt names an SSO
  bug and only 001 is one, so presenting both is a failure to disambiguate.
- **`get-missing` accepts the CLI's own miss output** (*"No exact match found for 042. Did
  you mean: 1. 002…"*). It offers a candidate without claiming it *is* 042, which answers
  the question. What fails is asserting 042's identity onto a fixture task, or answering
  about the near-match without ever mentioning 042.
- **`get-blocked-state` needs a startability verdict, not just the dependency.** The CLI
  hands the agent `Depends on: 002` for free, so the dependency half alone proves nothing.

## Other

- Fixture is the shared six-task base, forked verbatim — no edits, so
  [`../fixtures/README.md`](../fixtures/README.md)'s verified ground truth applies as-is.
  Graders assume six tasks.
- Keep `run:` in the `/bin/sh -c '…'` form.
- The task file for this suite (`tasks/01kk60r52-…`) proposed `get-blocked-state` as *"can
  I start task 006?"* while naming 003 as the target. 006 has no dependencies; **003** is
  the blocked task, and that is what the eval asks about.
