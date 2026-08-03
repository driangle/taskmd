---
id: "01kz3c24h"
title: "Refactor nolint-suppressed high-complexity functions"
status: completed
priority: medium
effort: medium
type: improvement
phase: critical-feedback
dependencies: []
tags: [cli, refactor, tech-debt]
created_at: 2026-08-03
completed_at: 2026-08-03
---

# Refactor nolint-suppressed high-complexity functions

## Objective

The project enforces strong complexity limits in `.golangci.yml` (funlen 60/45, gocyclo 15,
gocognit 20), but the worst offenders are **suppressed rather than fixed** with `//nolint`
plus "TODO: refactor" escape hatches. The clearest case is `runGraph` (`graph.go:90`) at
**217 lines — 3.6x the limit** — carrying `//nolint:gocognit,gocyclo,funlen // TODO: refactor`.
Similar suppressions live in `snapshot.go` (`runSnapshot`), `stats.go`, and
`internal/watcher/watcher.go`. These annotations quietly undermine the rule the project claims
to enforce. Pay the debt down.

## Tasks

- [x] Refactor `runGraph` (`graph.go:90`) into focused helpers under the funlen/complexity limits
- [x] Refactor `runSnapshot` (`snapshot.go`) to remove its `//nolint:funlen`
- [x] Refactor the suppressed function in `stats.go`
- [x] Refactor `internal/watcher/watcher.go` to remove its `//nolint:gocognit`
- [x] Remove the corresponding `//nolint` directives and confirm `make lint` passes clean
- [x] Grep for any remaining `//nolint ... TODO: refactor` and file follow-ups or fix them

## Acceptance Criteria

- No function in the CLI carries a `//nolint` complexity/funlen suppression with a refactor TODO
- `make lint` passes with the suppressions removed
- Behavior is unchanged: existing tests for graph/snapshot/stats/watcher still pass
