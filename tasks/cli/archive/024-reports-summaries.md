---
id: "archive-024"
title: "Reports & summaries"
status: pending
priority: medium
effort: medium
dependencies: ["archive-017"]
tags:
  - cli
  - go
  - reports
  - archived
created: 2026-02-08
---

# Reports & Summaries

## Objective

Add CLI subcommands for generating adhoc summaries, reports, and analysis of tasks — useful for standups, sprint planning, or project overviews. Output is printed to stdout (no TUI required).

## Tasks

- [ ] Add subcommand structure using cobra or simple flag-based routing
- [ ] `taskmd summary` — overview: total tasks, breakdown by status, priority distribution, groups, tags
- [ ] `taskmd list` — non-interactive flat list of tasks (for piping/scripting)
- [ ] `taskmd report` — detailed report: blocked tasks, dependency chains, stale tasks
- [ ] `taskmd deps` — show dependency graph as an ASCII tree or indented list
- [ ] Support `--format` flag: `text` (default), `json`, `csv`
- [ ] Support `--status`, `--priority`, `--group`, `--tag` filters on all subcommands
- [ ] Ensure default behavior (bare `taskmd` with no subcommand) still launches the TUI

## Acceptance Criteria

- `taskmd summary` prints a concise project overview to stdout
- `taskmd list` prints tasks in a scriptable format
- `taskmd report` identifies blocked and stale tasks
- `--format=json` outputs valid JSON
- Bare `taskmd` still opens the interactive TUI
