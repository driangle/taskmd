---
id: "052"
title: "Rename Go module to github.com/driangle/taskmd"
status: completed
priority: high
effort: medium
dependencies: []
tags:
  - refactor
  - infrastructure
  - mvp
created: 2026-02-12
---

# Rename Go Module to github.com/driangle/taskmd

## Objective

Rename the Go module from `github.com/driangle/md-task-tracker` to `github.com/driangle/taskmd` to match the project's identity and simplify the import path.

## Context

The project was originally named `md-task-tracker` but the CLI tool and branding use `taskmd`. The module path should reflect the canonical project name for consistency. This rename is a prerequisite for Homebrew distribution (task 045) and other packaging efforts.

## Tasks

- [x] Rename GitHub repository from `md-task-tracker` to `taskmd` (already done)
- [x] Update `go.mod` module path in `apps/cli/go.mod`
- [x] Update all Go import paths across the codebase
- [x] Update LDFLAGS in release workflow (`.github/workflows/release.yml`)
- [x] Update LDFLAGS in Makefile (`apps/cli/Makefile`) (no LDFLAGS in Makefile)
- [x] Update all references in documentation (`README.md`, `CLAUDE.md`, etc.)
- [x] Update `go install` commands in docs
- [x] Run `go mod tidy` and verify build
- [x] Run full test suite
- [ ] Tag a new release after rename (requires user action)

## Acceptance Criteria

- All Go files use `github.com/driangle/taskmd` as the module path
- `go build ./...` succeeds
- `go test ./...` passes
- Release workflow builds correctly with new module path
- Documentation references the new module path
- GitHub repository is renamed (or redirects are in place)

## Implementation Notes

### Files to Update

The module path appears in:
1. `apps/cli/go.mod` — module declaration
2. All `.go` files with imports referencing the old module path
3. `.github/workflows/release.yml` — LDFLAGS
4. `apps/cli/Makefile` — LDFLAGS
5. `README.md` — `go install` command
6. `CLAUDE.md` — any module references

### Estimated Scope

Approximately 48 files contain the old module path and need updating.

## References

- [Go Module Path Migration](https://go.dev/ref/mod#module-path)
- [GitHub Repository Renaming](https://docs.github.com/en/repositories/creating-and-managing-repositories/renaming-a-repository)
