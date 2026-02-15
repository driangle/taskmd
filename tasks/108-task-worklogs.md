---
id: "108"
title: "Support worklogs for tasks"
status: completed
priority: medium
effort: large
tags:
  - worklogs
  - convention
  - ai
  - mvp
touches:
  - cli
  - web
created: 2026-02-14
---

# Support Worklogs for Tasks

## Objective

Introduce a worklog convention where each task can have a companion worklog file that records progress notes, decisions, blockers, and session summaries. Agents (human or AI) working on a task are encouraged to create and append to this worklog. The worklogs are then surfaced via a new CLI command and a web page view.

## Tasks

### Convention & Specification
- [x] Define the worklog file convention (e.g., `tasks/.worklogs/<task-id>.md` or `tasks/cli/.worklogs/<task-id>.md` alongside the task)
- [x] Document the worklog format: timestamped entries with optional author, status updates, and free-form notes
- [x] Update `docs/taskmd_specification.md` with the worklog convention
- [x] Add worklog guidance to agent templates (CLAUDE.md, GEMINI.md, CODEX.md) explaining purpose, when to write entries, and examples of good worklogs

### Scanner & Parser
- [x] Extend the scanner to discover worklog files associated with tasks
- [x] Parse worklog entries (timestamp, author, content)
- [x] Link worklogs to their parent task by ID

### CLI Command
- [x] Add `taskmd worklog --task-id <ID>` command to view a task's worklog
- [x] Add `taskmd worklog --task-id <ID> --add "message"` to append a new entry with timestamp
- [x] Support `--format` flag (text, json, yaml) for worklog output
- [x] Show worklog summary in `taskmd get --task-id <ID>` output (e.g., entry count, last updated)
- [x] Add tests for worklog CLI commands

### Web UI
- [x] Add worklog display to the task detail view
- [x] Show worklog entries in chronological order with timestamps and authors
- [x] Add a visual indicator on task cards when a worklog exists

## Acceptance Criteria

- Worklog files follow a documented convention and live alongside task files
- `taskmd worklog --task-id 042` displays the worklog for task 042
- `taskmd worklog --task-id 042 --add "Started implementation"` appends a timestamped entry
- `taskmd get --task-id 042` shows worklog metadata (entry count, last update)
- Web task detail view displays worklog entries
- Worklogs are optional — tasks without worklogs behave exactly as before
- Convention is documented in the specification
- Tests cover worklog creation, appending, viewing, and edge cases (missing worklog, empty worklog)
