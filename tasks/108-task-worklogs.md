---
id: "108"
title: "Support worklogs for tasks"
status: in-progress
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
- [ ] Define the worklog file convention (e.g., `tasks/.worklogs/<task-id>.md` or `tasks/cli/.worklogs/<task-id>.md` alongside the task)
- [ ] Document the worklog format: timestamped entries with optional author, status updates, and free-form notes
- [ ] Update `docs/taskmd_specification.md` with the worklog convention
- [ ] Add worklog guidance to agent templates (CLAUDE.md, GEMINI.md, CODEX.md) explaining purpose, when to write entries, and examples of good worklogs

### Scanner & Parser
- [ ] Extend the scanner to discover worklog files associated with tasks
- [ ] Parse worklog entries (timestamp, author, content)
- [ ] Link worklogs to their parent task by ID

### CLI Command
- [ ] Add `taskmd worklog --task-id <ID>` command to view a task's worklog
- [ ] Add `taskmd worklog --task-id <ID> --add "message"` to append a new entry with timestamp
- [ ] Support `--format` flag (text, json, yaml) for worklog output
- [ ] Show worklog summary in `taskmd get --task-id <ID>` output (e.g., entry count, last updated)
- [ ] Add tests for worklog CLI commands

### Web UI
- [ ] Add worklog display to the task detail view
- [ ] Show worklog entries in chronological order with timestamps and authors
- [ ] Add a visual indicator on task cards when a worklog exists

## Acceptance Criteria

- Worklog files follow a documented convention and live alongside task files
- `taskmd worklog --task-id 042` displays the worklog for task 042
- `taskmd worklog --task-id 042 --add "Started implementation"` appends a timestamped entry
- `taskmd get --task-id 042` shows worklog metadata (entry count, last update)
- Web task detail view displays worklog entries
- Worklogs are optional — tasks without worklogs behave exactly as before
- Convention is documented in the specification
- Tests cover worklog creation, appending, viewing, and edge cases (missing worklog, empty worklog)
