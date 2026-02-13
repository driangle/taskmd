---
id: "072"
title: "Add spec command to generate task specification file"
status: completed
priority: medium
effort: small
dependencies:
  - "061"
tags:
  - cli
  - go
  - commands
  - dx
  - mvp
created: 2026-02-13
---

# Add `spec` Command to Generate Task Specification File

## Objective

Add a `taskmd spec` command that outputs the taskmd specification document to the user's project directory, giving them a local reference for the task file format.

## Context

The task specification lives at `docs/taskmd_specification.md` in the source repo, but users who install via Homebrew don't have access to it. A `taskmd spec` command follows the same pattern as `taskmd agents init` (embed a file, write it out) and gives users a quick way to generate the spec locally.

Task 061 (simplify specification) should be completed first so we embed the final, simplified version.

## Tasks

- [ ] Embed `docs/taskmd_specification.md` into the binary using `//go:embed`
- [ ] Create `internal/cli/spec.go` with a `spec` cobra command
- [ ] Write the spec to `TASKMD_SPEC.md` in the target directory by default
- [ ] Add `--dir` flag to specify output directory
- [ ] Add `--force` flag to overwrite an existing file
- [ ] Add `--stdout` flag to print to stdout instead of writing a file
- [ ] Refuse to overwrite without `--force` and print a clear message
- [ ] Print success message with the output file path
- [ ] Create `internal/cli/spec_test.go` with comprehensive tests
- [ ] Run `make lint` and `make test` to verify

## Acceptance Criteria

- `taskmd spec` writes `TASKMD_SPEC.md` to the current directory
- `taskmd spec --dir ./docs` writes to a specific directory
- `taskmd spec --force` overwrites an existing file
- `taskmd spec --stdout` prints the spec without writing a file
- Running without `--force` when `TASKMD_SPEC.md` exists returns an error
- Content matches the embedded specification
- All tests pass, lint passes

## Test Cases

- Happy path: writes spec to a temp directory
- Refuses to overwrite existing file without `--force`
- Overwrites with `--force`
- `--stdout` prints to stdout, does not create a file
- `--dir` writes to the specified directory
- `--dir` with non-existent directory returns an error
- Verify written content matches embedded spec

## Examples

```bash
taskmd spec                    # Writes TASKMD_SPEC.md to current directory
taskmd spec --stdout           # Print spec to stdout
taskmd spec --dir ./docs       # Write to docs/ directory
taskmd spec --force            # Overwrite existing file
taskmd spec --stdout | less    # Browse the spec in a pager
```

## References

- `docs/taskmd_specification.md` — the spec to embed
- `internal/cli/init.go` — existing embed + write pattern to follow
- Task 061 — simplify specification (should complete first)
