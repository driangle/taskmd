## 2026-08-22T20:07:06Z

Started work. Approach: implement owner-aware actionability in sdk/go/next (additive: new field on config or filter), add --for flag to CLI next command, update --explain output, then tests (sdk unit, CLI, e2e) and help-text docs. Reading sdk/go/next and internal/cli/next.go first.

## 2026-08-22T21:12:56Z

Implementation complete. SDK (sdk/go/next, additive only): Options.For, exported ExcludedByOwner predicate (in-progress + non-empty owner != For), owner-exclusion in filterActionable, owned-by-For as primary sort key with 'owned by <for>' reason, Recommendation.Owner field. CLI: next --for flag (wired into runNext and all-projects path), --explain now appends an 'Excluded by ownership:' section naming each owner, new 'owner' column, claim convention documented in next and set help text. Key decision: pending tasks with an owner are assigned-not-claimed and stay actionable for everyone (per spec §6 the rule applies only to in-progress). Tests: sdk/go/next/owner_test.go (7 cases), internal/cli/next_owner_test.go (6 tests), e2e TestNext_OwnerAware. make test, make e2e, lint (cli + sdk) all green; manual smoke with taskmd-dev verified. No spec-sync needed; sdk pin rides CI auto-heal (patch).

## 2026-08-22T21:23:17Z

REVERTED same day. After discussion, owner-based coordination was rejected as a scope decision: the coordination unit is one agent per worktree, where the overlay's effective status alone prevents double-assignment; repurposing the display-only owner field as a claim would silently hide a solo user's own in-progress tasks from plain 'next' the moment they fill the field in (violating the no-behavior-change acceptance criterion in spirit), and it created an unresolved identity-convention problem (human vs agent names, fleets per human). All code changes were discarded before commit. Spec updated: goal removed, rationale recorded under Non-goals, §6 replaced with a tombstone (numbering kept for §7–§9 cross-references), 01m0nk00d dependency and owner wording dropped. Claims sidecar (spec open question 1) remains the designated answer if same-checkout coordination ever becomes real.
