# taskmd Specification

**Version:** 1.2
**Last Updated:** 2026-02-15

## Quick Reference

Each task is a `.md` file with YAML frontmatter and a markdown body.

```yaml
---
id: "001"
title: "Task title"
status: pending
---

# Task Title

Description and subtasks go here.
```

### Field Summary

| Field | Type | Required | Values / Format |
|-------|------|----------|-----------------|
| `id` | string | **Yes** | Zero-padded number (e.g., `"001"`, `"042"`) |
| `title` | string | **Yes** | Brief, descriptive text |
| `status` | enum | Recommended | `pending`, `in-progress`, `completed`, `blocked`, `cancelled` |
| `priority` | enum | No | `low`, `medium`, `high`, `critical` |
| `effort` | enum | No | `small`, `medium`, `large` |
| `dependencies` | array | No | List of task ID strings (e.g., `["001", "015"]`) |
| `tags` | array | No | Lowercase, hyphen-separated strings |
| `group` | string | No | Logical grouping (derived from directory if omitted) |
| `owner` | string | No | Free-form assignee name or identifier |
| `touches` | array | No | Abstract scope identifiers (e.g., `["cli/graph", "cli/output"]`) |
| `parent` | string | No | Single task ID (e.g., `"045"`) |
| `created` | date | No | `YYYY-MM-DD` |
| `verify` | array | No | List of typed verification checks (see below) |

## Frontmatter Schema

### Required Fields

**`id`** — Unique identifier for the task. Use zero-padded numeric IDs (e.g., `"001"`, `"042"`). Must be unique across all tasks in the project.

**`title`** — Brief, action-oriented description of the task.

### Optional Fields

**`status`** — Current state of the task (recommended for all tasks):

| Status | Meaning |
|--------|---------|
| `pending` | Not started (initial state) |
| `in-progress` | Currently being worked on |
| `completed` | Finished and verified |
| `blocked` | Cannot proceed due to a blocker |
| `cancelled` | Will not be completed |

```
pending → in-progress → completed
   ↓            ↓            ↓
   ↓         blocked        ↓
   ↓            ↓           ↓
   └──→ cancelled ←─────────┘
```

**`priority`** — Importance level:

| Priority | Use Case |
|----------|----------|
| `low` | Nice to have, can be deferred |
| `medium` | Standard work items (default) |
| `high` | Important for project success |
| `critical` | Urgent, must address immediately |

**`effort`** — Estimated complexity:

| Effort | Typical Duration |
|--------|------------------|
| `small` | < 2 hours |
| `medium` | 2–8 hours |
| `large` | > 8 hours / multi-day |

**`dependencies`** — List of task IDs that must be completed before this task can start. Always reference by ID, always use array format:

```yaml
dependencies: ["001", "015"]
```

**`tags`** — Labels for categorization and filtering. Use lowercase, hyphen-separated strings:

```yaml
tags:
  - core
  - api
```

**`group`** — Logical grouping. If omitted, derived from the parent directory name. Root-level tasks have no group.

**`owner`** — Free-form string for assigning a task to a person or team. Used for filtering and display; no validation is applied.

**`touches`** — List of abstract scope identifiers declaring which code areas a task modifies. Used by the `tracks` command to detect spatial overlap and assign tasks to parallel work tracks. Two tasks that share a scope should not be worked on simultaneously (risk of merge conflicts).

```yaml
touches:
  - cli/graph
  - cli/output
```

Scopes are user-defined identifiers. Concrete scope-to-path mappings can be configured in `.taskmd.yaml`:

```yaml
# .taskmd.yaml
scopes:
  cli/graph:
    description: "Graph visualization and dependency rendering"
    paths:
      - "apps/cli/internal/graph/"
      - "apps/cli/internal/cli/graph.go"
  cli/output:
    paths:
      - "apps/cli/internal/cli/format.go"
```

The optional `description` field provides a human-readable explanation of what a scope covers. When present, it is included in validation error messages for easier debugging.

When scopes are configured, `touches` values not found in the config produce a warning. When no scopes config exists, all values are accepted silently.

**`parent`** — Task ID of the parent task for hierarchical grouping. A task can have at most one parent. Children are computed dynamically (not stored in frontmatter) by finding all tasks whose `parent` matches a given ID.

- Purely organizational — does not imply blocking or dependency
- No status cascading — completing all children does not auto-complete the parent
- Must reference an existing task ID; self-references and cycles are flagged by validation

```yaml
parent: "045"
```

**`created`** — Date when the task was created, in `YYYY-MM-DD` format.

**`verify`** — List of typed acceptance checks for validating task completion. Each entry is a map with a `type` field that determines the check kind. Run checks with `taskmd verify --task-id <ID>`.

| Type | Fields | Behavior |
|------|--------|----------|
| `bash` | `run` (required), `dir` (optional) | Runs `run` in a shell subprocess; pass if exit code 0, fail otherwise |
| `assert` | `check` (required) | Displays `check` text for an agent to evaluate (not executed) |

- `dir` is relative to the project root (where `.taskmd.yaml` lives); defaults to `.`
- Unknown types are preserved in the file but produce a warning and are skipped during execution

```yaml
verify:
  - type: bash
    run: "go test ./internal/api/... -run TestPagination"
    dir: "apps/cli"
  - type: bash
    run: "npm test"
    dir: "apps/web"
  - type: assert
    check: "Pagination links appear in the API response headers"
  - type: assert
    check: "Page size defaults to 20 when not specified"
```

Unknown frontmatter fields are preserved during read/write operations.

## File Organization

### File Naming

Task files follow this pattern:

```
NNN-descriptive-title.md
```

Where `NNN` is the zero-padded task ID and `descriptive-title` is a lowercase hyphen-separated slug. Examples:

- `001-project-scaffolding.md`
- `042-implement-user-auth.md`

The ID prefix may be omitted if the `id` field in frontmatter is the sole identifier.

### Directory Structure

Tasks can be organized into subdirectories for grouping:

```
tasks/
├── 001-taskmd-specification.md     # No group
├── web/                             # Group: "web"
│   ├── 001-project-scaffolding.md
│   └── 002-typescript-types.md
└── cli/                             # Group: "cli"
    ├── 015-go-cli-scaffolding.md
    └── 016-task-model-parsing.md
```

Group resolution priority:
1. Explicit `group` in frontmatter
2. Parent directory name
3. No group (root-level tasks)

## Validation

A valid taskmd file **must**:

1. Have YAML frontmatter enclosed in `---` delimiters
2. Include required fields: `id`, `title`
3. Use valid enum values for `status`, `priority`, `effort`
4. Have unique IDs across the project
5. Reference only existing tasks in `dependencies`
6. Have no circular dependency chains
7. Reference an existing task in `parent` (if set), with no self-reference or parent cycles

A valid taskmd file **should**:

1. Follow the `NNN-task-name.md` naming pattern
2. Include a creation date
3. Have a descriptive markdown body

## Examples

### Minimal Task

```markdown
---
id: "001"
title: "Fix login button alignment"
status: pending
---

# Fix Login Button Alignment

The login button on the homepage is misaligned. Update the CSS to center it.
```

### Full Task

```markdown
---
id: "015"
title: "Implement user authentication"
status: in-progress
priority: high
effort: large
dependencies: ["012", "013"]
parent: "012"
tags:
  - auth
  - security
  - api
created: 2026-02-08
---

# Implement User Authentication

## Objective

Add JWT-based authentication to the API.

## Tasks

- [x] Design authentication flow
- [x] Implement JWT signing and verification
- [ ] Create login endpoint
- [ ] Create logout endpoint
- [ ] Add authentication middleware
- [ ] Write integration tests

## Acceptance Criteria

- Users can log in with email and password
- JWT tokens expire after 24 hours
- Protected routes require valid JWT
- All endpoints have > 90% test coverage
```
