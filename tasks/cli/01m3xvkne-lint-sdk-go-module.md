---
title: "Lint sdk/go in make lint, pre-commit and CI"
id: "01m3xvkne"
status: pending
priority: medium
type: chore
tags: ["cli", "sdk", "lint", "tech-debt"]
created: "2026-10-02"
effort: medium
phase: critical-feedback
---

# Lint sdk/go in make lint, pre-commit and CI

## Objective

The complexity and quality limits in `apps/cli/.golangci.yml` (funlen, gocyclo,
gocognit 20, revive, goconst, gofmt/goimports) are only enforced on `apps/cli`.
Since ADR 0006 moved the task model into the separate `sdk/go` module, that code
has no lint gate: `sdk/go` has no `.golangci.yml`, `make lint` runs only in
`apps/cli`, and the CI/release `golangci-lint-action` steps use
`working-directory: apps/cli`. Running the CLI config against the SDK
(`cd sdk/go && golangci-lint run --config ../../apps/cli/.golangci.yml ./...`)
reported 34 issues on 2026-10-02: gocognit 7 (e.g. `next.filterActionable`,
`next.buildDepComponents`), goconst 10, gofmt 3, goimports 1, revive 13.

Surfaced while doing 01m3xf0rs; pre-existing, not caused by it.

## Tasks

- [ ] Give `sdk/go` a lint config (its own `.golangci.yml`, or share the CLI one)
- [ ] Run SDK lint from `make lint` / `make check-lite` (so the pre-commit hook covers it)
- [ ] Add an `sdk/go` lint step to `.github/workflows/ci.yml` (and `release.yml`)
- [ ] Fix the existing findings, or add justified `//nolint` with a reason; note that
      revive renames of exported SDK symbols (e.g. `worklog.WorklogPath`) are breaking
      API changes — suppress those rather than rename, or ship with `sdk-bump: minor`

## Acceptance Criteria

- `make lint` (and therefore the pre-commit hook) fails on a lint violation in `sdk/go`
- CI runs golangci-lint against `sdk/go` and it passes on main
- No exported SDK symbol is renamed without a declared `sdk-bump: minor`
