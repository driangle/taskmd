---
id: "021"
title: "snapshot command - Static machine-readable output"
status: completed
priority: high
effort: medium
dependencies: ["017", "019"]
tags:
  - cli
  - go
  - commands
  - export
created: 2026-02-08
---

# Snapshot Command - Static Machine-Readable Output

## Objective

Implement the `snapshot` command to produce a frozen, machine-readable representation of tasks for CI/CD pipelines and automation.

## Tasks

- [x] Create `internal/cli/snapshot.go` for snapshot command
- [x] Support output formats: `json`, `yaml`, `md`
- [x] Implement `--core` flag: output only core fields (id, title, duration, dependencies)
- [x] Implement `--derived` flag: include computed fields (blocked status, depth, topological order, critical path)
- [x] Implement `--group-by <field>` for grouping tasks in output
- [x] Implement `--out <file>` for writing to file instead of stdout
- [x] Calculate derived fields:
  - Blocked status (dependencies not met)
  - Dependency depth
  - Topological order
  - Critical path membership
- [x] Default: output all fields in JSON format

## Acceptance Criteria

- `taskmd snapshot` outputs all task data in JSON format
- `--format yaml` produces valid YAML
- `--format md` produces markdown output
- `--core` includes only essential fields
- `--derived` includes computed dependency analysis
- `--group-by status` groups tasks by status field
- `--out file.json` writes to file
- Works with stdin and explicit file paths

## Examples

```bash
taskmd snapshot > snapshot.json
taskmd snapshot --format yaml --out snapshot.yaml
taskmd snapshot --core --format json
taskmd snapshot --derived --group-by status
cat tasks.md | taskmd snapshot --stdin
```
