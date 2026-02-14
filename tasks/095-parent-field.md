---
id: "095"
title: "Add parent field for hierarchical task relations"
status: pending
priority: medium
effort: large
related: ["094"]
tags:
  - feature
  - spec
  - cli
  - web
  - mvp
created: 2026-02-14
---

# Add Parent Field for Hierarchical Task Relations

## Objective

Add a `parent` frontmatter field that lets a task declare itself as a subtask of another task. This is a single task ID string (not an array — a task has exactly one parent). The inverse ("children") is computed dynamically, not stored.

This complements `dependencies` (blocking order) and `related` (loose association) with a third relationship type: hierarchical grouping.

## Tasks

### Specification
- [ ] Add `parent` field to `docs/taskmd_specification.md` as an optional string (single task ID)
- [ ] Document semantics: hierarchical grouping, no implicit blocking, no status cascading
- [ ] Clarify that children are computed (not stored) — find all tasks whose `parent` matches a given ID

### Model & Parser
- [ ] Add `Parent string` field to the Task struct in `internal/model/task.go`
- [ ] Ensure YAML/JSON serialization tags are correct (`yaml:"parent,omitempty" json:"parent,omitempty"`)
- [ ] Verify parser handles `parent` field correctly

### Validation
- [ ] Validate that parent task ID references an existing task
- [ ] Warn if a task lists itself as parent
- [ ] Detect parent cycles (A parent of B, B parent of A)
- [ ] Add tests for parent field validation

### CLI — `get` command
- [ ] Display parent task in `taskmd get` output (e.g., "Parent: #045 — Homebrew distribution")
- [ ] Display computed children list when viewing a parent task
- [ ] Add tests

### CLI — `set` command
- [ ] Support `--parent 045` flag to set parent task
- [ ] Support `--parent ""` to clear parent
- [ ] Add tests

### CLI — `list` command
- [ ] Support `--parent 045` filter to show only children of a task
- [ ] Support `parent=true/false` filter to find tasks that have/don't have a parent
- [ ] Add tests

### CLI — `graph` command
- [ ] Consider rendering parent-child as a distinct edge style (or subgraph clustering)
- [ ] Add tests

### Web UI
- [ ] Display parent task as a clickable link in task detail view
- [ ] Display children list in parent task's detail view

## Design Decisions

- **Single parent only** — a task belongs to at most one parent (no multi-parent)
- **No status cascading** — completing all children does NOT auto-complete the parent
- **No implicit blocking** — parent/child is purely organizational, not a dependency
- **Children are computed** — no `children` field in frontmatter; derived by scanning for tasks with matching `parent`
- **Cross-directory** — parent-child can span task directories (unlike directory-based grouping)

## Non-Goals

- No auto-completion of parent when all children complete
- No inheritance of fields (priority, tags, etc.) from parent to children
- No nested parent chains beyond what naturally occurs (A → B → C is fine, but no special deep-hierarchy features)

## Acceptance Criteria

- `parent` field is documented in the specification
- Tasks can declare a parent via frontmatter: `parent: "045"`
- `taskmd get` displays parent and computed children
- `taskmd set --parent` updates the field
- `taskmd list --parent 045` filters to children
- Validation catches references to non-existent parent tasks and self-references
- All new functionality has tests
- Web UI shows parent and children in detail view
