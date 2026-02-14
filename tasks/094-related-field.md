---
id: "094"
title: "Add related field for non-dependency task relations"
status: pending
priority: medium
effort: large
tags:
  - feature
  - spec
  - cli
  - web
  - mvp
  - mvp
created: 2026-02-14
---

# Add Related Field for Non-Dependency Task Relations

## Objective

Add a `related` frontmatter field that lets tasks reference other tasks they are conceptually connected to, without implying any blocking or ordering. This is a flat list of task IDs with the same shape as `dependencies`, meaning "these tasks are connected/relevant to each other."

Bidirectional by convention: if task A lists task B as related, B is related to A even if B doesn't list A.

## Tasks

### Specification
- [ ] Add `related` field to `docs/taskmd_specification.md` as an optional `array` of task ID strings
- [ ] Document semantics: non-blocking, non-ordering, bidirectional by convention

### Model & Parser
- [ ] Add `Related []string` field to the Task struct in `internal/model/task.go`
- [ ] Ensure YAML/JSON serialization tags are correct (`yaml:"related" json:"related"`)
- [ ] Verify parser handles `related` field correctly (omitempty behavior)

### Validation
- [ ] Validate that related task IDs reference existing tasks (reuse dependency validation pattern)
- [ ] Warn if a task lists itself as related
- [ ] Add tests for related field validation

### CLI — `get` command
- [ ] Display related tasks in `taskmd get` output
- [ ] Add tests

### CLI — `set` command
- [ ] Support `--related 058,063` flag to set related tasks
- [ ] Add tests

### CLI — `graph` command
- [ ] Render related edges as dashed/dotted lines (visually distinct from dependency edges)
- [ ] Mermaid: use dotted arrow syntax (`-.->`)
- [ ] DOT: use `style=dashed`
- [ ] ASCII: separate "Related" section or annotation
- [ ] JSON: add `relatedEdges` array alongside existing `edges`
- [ ] Add tests

### Filtering
- [ ] Support `related=true/false` filter in the filter package
- [ ] Add tests

### Web UI
- [ ] Display related tasks in the task detail view as clickable links

## Non-Goals

- No effect on `next` command scoring or actionability
- No cycle detection for relations (non-directional, non-blocking)
- No cascading status changes
- No typed relations (parent, blocks, etc.) — keep it simple for now

## Acceptance Criteria

- `related` field is documented in the specification
- Tasks can declare related tasks via frontmatter: `related: ["058", "063"]`
- `taskmd get` displays related tasks
- `taskmd set --related` updates the field
- `taskmd graph` renders related edges distinctly from dependency edges
- Validation catches references to non-existent tasks
- All new functionality has tests
- Web UI shows related tasks in detail view
