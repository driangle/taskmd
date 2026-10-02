## 2026-10-02T09:25:55Z

Made strict phase ordering the default for `next` on every surface and deprecated `--strict-phases`.

**SDK:** `scoreAndSort` now assigns a `phaseTier` per task (`assignPhaseTiers`). Unphased tasks join the current tier, which is the earliest configured phase among the tasks being ranked. It is computed after filters, so it is the earliest phase in the visible set. Unknown-phase tasks still sort last. If no task is in a configured phase, unphased tasks get tier 0, so unknown-phase tasks stay behind them. There is no API change, so this is a patch bump; the commit message calls out the behaviour change.

**CLI:** `--strict-phases` defaults to true and is marked with `MarkDeprecated`. It still parses and prints a warning, and is hidden from help. `--strict-phases=false` is the escape hatch, though it also prints the warning. Updated the long help and examples. `--all-projects` inherits the default via `recommendForProject`.

**MCP/web/export:** pass `StrictPhases: true`.

**Tests:**
- Added SDK tier tests in `next_phase_tier_test.go`.
- Flipped the CLI default test, and added tests for the deprecation warning, the `=false` opt-out, unphased placement and the no-phases-unchanged case.
- Added strict-by-default tests for MCP, `/api/next` and `next.json`. Their fixtures pit a critical late-phase task against a low early-phase one, so only strict tiering can put the early task first.

**Docs:** `apps/docs/guide/cli.md` now has a phase ordering section and an updated flag table.

**Release notes:** the repo has no changelog file, because notes are written at release time from commits. The behaviour change is recorded in the commit body.

**Verified:** `go test ./...` (SDK and CLI), `make e2e` and `make lint` (CLI) are green. A manual run of the dev binary confirmed the default, the `=false` opt-out and the deprecation warning.
