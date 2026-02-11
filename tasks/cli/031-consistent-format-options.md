---
id: "031"
title: "Consistent formatting options across CLI commands"
status: pending
priority: low
effort: medium
dependencies: ["023"]
tags:
  - cli
  - go
  - quality
  - ux
  - post-mvp
created: 2026-02-08
---

# Consistent Formatting Options Across CLI Commands

## Objective

Standardize the `--format` flag behavior across all CLI commands so users get a predictable experience regardless of which command they run.

## Problem

Each command currently defines its own `--format` flag with different defaults, different supported values, and different output behavior:

| Command | Flag source | Default | Supported formats |
|---------|------------|---------|-------------------|
| root (list, stats, etc.) | `PersistentFlags` | `table` | table, json, yaml |
| graph | local flag | `mermaid` | mermaid, dot, ascii, json |
| board | local flag | `md` | md, txt, json |

This creates several issues:
- **Inconsistent defaults**: `table` vs `mermaid` vs `md`
- **Overlapping local vs global flags**: `graph` and `board` shadow the global `--format` with their own flag, which can confuse users
- **No shared formats**: `json` is the only format available everywhere, but its structure differs per command
- **No `table` support** in graph or board; no `yaml` support in graph or board

## Tasks

- [ ] Audit all commands for `--format` flag usage and output behavior
- [ ] Define a shared set of universal formats all commands must support:
  - `json` - Structured JSON (every command)
  - `yaml` - Structured YAML (every command)
  - At least one human-readable default per command (table, ascii, md, etc.)
- [ ] Decide on flag strategy — either:
  - **(a)** Keep the global `--format` and have each command validate its own supported values, or
  - **(b)** Remove the global `--format` and let each command own its flag entirely
- [ ] Ensure `--format json` produces a consistent envelope structure across commands (e.g., always top-level object with a metadata key)
- [ ] Add a shared `formatOutput` helper or interface so new commands get formatting for free
- [ ] Update help text to clearly list supported formats per command
- [ ] Add tests verifying that every command accepts `json` and returns valid JSON

## Acceptance Criteria

- `--format json` works on every command and produces valid, parseable JSON
- Help text for each command clearly lists its supported formats
- No silent conflict between global and local `--format` flags
- Adding a new command with standard formats requires minimal boilerplate

## Examples

```bash
# These should all produce valid JSON
taskmd list --format json
taskmd stats --format json
taskmd graph --format json
taskmd board --format json

# Human-readable defaults should still work
taskmd list                # table
taskmd graph               # mermaid or ascii
taskmd board               # md
```
