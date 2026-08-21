---
title: "Honor the worklogs config key or stop documenting it"
id: "01m0jfw0q"
status: pending
priority: medium
type: bug
tags: ["worklogs", "config", "docs"]
created: "2026-08-21"
---

# Honor the worklogs config key or stop documenting it

## Objective

`worklogs: true|false` in `.taskmd.yaml` is documented in eleven places — including the agent
templates `taskmd init` itself generates — and six shipped skills gate behavior on it. No Go code
reads it, and `taskmd validate` reports it as an unknown key. A user who follows our own generated
instructions gets warned at for doing so.

Close the loop in one direction or the other. Recommended: implement the key, since the skills
already depend on it and the convention is useful — the missing piece is the CLI honoring what it
is told.

## Steps to Reproduce

1. `taskmd init` in an empty directory
2. Follow the generated `tasks/CLAUDE.md`, which says worklogs are recorded "when worklogs are
   enabled (`worklogs: true` in `.taskmd.yaml`)" — add that key to `.taskmd.yaml`
3. `taskmd add "probe"`
4. `taskmd validate`
5. Separately, with **no** `worklogs` key at all, run `taskmd worklog <id> --add "entry"`

## Expected Behavior

Setting a documented config key validates cleanly, and the CLI respects it: with
`worklogs: false` (or the key absent, if absence means disabled) `taskmd worklog` declines to
write and says why.

## Actual Behavior

Step 4 warns, with the same message a typo produces:

```
unknown config key: 'worklogs'
  File: /tmp/.../.taskmd.yaml

Validated 1 task(s) with 1 warning(s)
```

Step 5 writes `tasks/.worklogs/001.md` regardless — `taskmd worklog` has no gate at all. There is
no `viper.Get*("worklogs")` anywhere in `apps/cli` or `sdk/go`; config is viper-backed, so the key
parses fine and is then simply never read.

The two halves disagree about what "enabled" means, which is the actual defect: the skills treat
*absent* as disabled ("skip all worklog steps silently"), while the CLI treats absent as enabled.

## Tasks

- [ ] Decide the semantics and write them down: does an absent key mean enabled or disabled?
      The skills assume disabled; the CLI behaves as enabled. Changing the CLI to default-off is a
      behavior change for existing users, so favour default-on and make `worklogs: false` the
      opt-out — then fix the skills, which are the side that is wrong
- [ ] Read the key in the CLI and gate `taskmd worklog` writes on it, with a clear message when
      declining rather than silent success
- [ ] Add `worklogs` to the known top-level config keys so `taskmd validate` stops flagging it
      (see `loadConfigForValidation` in `apps/cli/internal/cli/validate.go` and the validator's
      `ConfigData.TopKeys`)
- [ ] Check the other write paths for the same gap — `taskmd rm`, `feed`, and the web export all
      touch worklogs
- [ ] Align the six skills that check the key: `do-task`, `complete-task` and `divide-and-conquer`
      in both `claude-code-plugin/` and `claude-code-plugin-lite/`
- [ ] Align the docs to whatever is implemented: `CLAUDE.md:277`, `tasks/CLAUDE.md:80`, the
      `taskmd init` templates (`CLAUDE.md`, `GEMINI.md`, `CODEX.md`, each line 98), and
      `claude-code-plugin-lite/SPEC_REFERENCE.md:92`
- [ ] Document the key in `docs/taskmd_specification.md`, which currently does not mention it at
      all — then `make sync-spec`
- [ ] Add a CLI test covering enabled, explicitly disabled, and absent
- [ ] Update the note in `evals/fixtures/README.md`, which records the current broken behavior for
      eval authors

## Acceptance Criteria

- `worklogs: true` and `worklogs: false` both validate without warnings
- `taskmd worklog` observes the setting, and says why when it declines to write
- Absent-key semantics are stated in the spec and match between the CLI and all six skills
- No doc or template instructs the user to set a key the CLI ignores
- Tests cover enabled / disabled / absent

## Notes

Found while migrating the skill benchmarks to skival. A `do-task` eval was written asserting the
correct behavior was **no worklog file**, reasoning from the documented convention — the opposite
of what the CLI does. Verified against `taskmd` 0.4.1.
