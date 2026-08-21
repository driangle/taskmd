## 2026-08-21T20:03:35Z

Implemented the worklogs config key rather than removing the documentation, per the task's recommendation.

**Semantics decided:** default-on. Absent or `worklogs: true` means enabled; `worklogs: false` is the opt-out. Default-off would have silently stopped worklog writes for every existing project, so the skills (which assumed absent = disabled) were the side that was wrong.

**Scope of the gate:** writes only. `taskmd worklog --add` now returns an error explaining why instead of silently succeeding. Reads (`worklog` view, `get`, `feed --source worklog`, web API/export) and `taskmd rm`'s worklog cleanup stay ungated — they operate on files that already exist. The web API has no worklog write endpoint yet; task 179 already plans to check the config before allowing edits.

**Changes:** `worklogsEnabled()` in `apps/cli/internal/cli/worklog.go` (viper `IsSet` || `GetBool`); `worklogs` added to `knownConfigKeys` in `sdk/go/validator/validator.go`; new `## Worklogs` section in the canonical spec plus `make sync-spec`; six skills, three init templates, two eval workspace templates, docs site (configuration, best-practices, cli guide), `docs/templates/CLAUDE.md`, and the lite plugin SPEC_REFERENCE realigned; the stale note in `evals/fixtures/README.md` rewritten.

**Tests:** three unit tests in `worklog_test.go` (explicit true, false, read-with-false) plus a comment marking the existing add test as the absent case; a validator test for the known key; three e2e groups in `internal/e2e/config_test.go` covering enabled/disabled/absent writes and clean validation of both values. `make test`, `make e2e`, `make lint` all pass.

## 2026-08-21T20:11:16Z

Reversed the default per user direction: worklogs are now opt-in, not opt-out. Only `worklogs: true` permits a write; an absent key means disabled, matching what the skills and most docs already claimed. The CLI moved instead of the skills.

This is a behavior change for any project that relied on `taskmd worklog --add` working with no config, so this repo's `.taskmd.yaml` now sets `worklogs: true` explicitly — without it the project's own AGENTS.md worklog instruction would have silently stopped working. `taskmd init` writes no `.taskmd.yaml` at all, so fresh projects start with worklogs off, which the generated templates now state correctly.

The decline message names the key to set: "worklogs are disabled for this project, so no entry was written for task X; set `worklogs: true` in .taskmd.yaml to enable them".

All docs, templates, six skills, and the eval fixtures README flipped back to opt-in wording. Tests restructured: the absent-key case now sits alongside explicit-false in the disabled table tests (unit and e2e), and the two pre-existing add tests seed `worklogs: true` via a new `enableWorklogs` helper. make test, make e2e, make lint pass.
