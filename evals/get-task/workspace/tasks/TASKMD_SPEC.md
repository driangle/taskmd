# taskmd Specification

> Trimmed for agent use. Full spec: `taskmd spec --stdout`

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

### Fields

| Field | Required | Values / Format |
|-------|----------|-----------------|
| `id` | **Yes** | Unique string (e.g., `"001"`, `"cli-049"`) |
| `title` | **Yes** | Brief, descriptive text |
| `status` | Recommended | `pending`, `in-progress`, `completed`, `in-review`, `blocked`, `cancelled` |
| `priority` | No | `low`, `medium`, `high`, `critical` |
| `effort` | No | `small` (< 2h), `medium` (2-8h), `large` (> 8h) |
| `type` | No | `feature`, `bug`, `improvement`, `chore`, `docs` |
| `dependencies` | No | Task IDs that must complete first (e.g., `["001", "015"]`) |
| `tags` | No | Lowercase, hyphen-separated (e.g., `[core, api]`) |
| `group` | No | Logical grouping (derived from directory if omitted) |
| `owner` | No | Free-form assignee |
| `phase` | No | Phase identifier (e.g., `"v0.2"`) — see Configuration below |
| `touches` | No | Scope identifiers for conflict detection (e.g., `["cli/graph"]`) |
| `context` | No | File paths relevant to the task (e.g., `["docs/api-design.md"]`) |
| `parent` | No | Parent task ID — organizational only, no blocking |
| `created_at` | No | `YYYY-MM-DD` (alias: `created`) |
| `verify` | No | Acceptance checks (see below) |
| `pr` | No | Pull request URLs — managed via `taskmd set --add-pr <url>` |

### Status Flow

```
pending -> in-progress -> in-review -> completed
   |            |              |
   v            v              v
blocked         +-> cancelled <+
```

### Verify Checks

Run with `taskmd verify <ID>`:

```yaml
verify:
  - type: bash
    run: "go test ./..."
    dir: "apps/cli"          # optional working directory
  - type: assert
    check: "API returns paginated results"  # agent-evaluated, not executed
```

## File Organization

Files: `ID-descriptive-title.md` (e.g., `042-implement-user-auth.md`)

```
tasks/
  001-spec.md                   # No group
  cli/                          # Group: "cli"
    015-go-cli-scaffolding.md
  web/                          # Group: "web"
    001-project-scaffolding.md
```

Group resolution: explicit `group` field > parent directory name > none.

## Configuration (.taskmd.yaml)

Project settings in `.taskmd.yaml` at the repository root.

### Phases

Time-based groupings (sprints, releases). Tasks reference a phase via the `phase` field.

```yaml
phases:
  - id: core-cli
    name: "Core CLI"
    description: "Core CLI features"
    due: 2026-04-01
  - id: web-dashboard
    name: "Web Dashboard"
    due: 2026-06-01
```

| Field | Required | Description |
|-------|----------|-------------|
| `id` | No | Stable key matched by task `phase` (falls back to `name`) |
| `name` | Yes | Display label |
| `description` | No | Description |
| `due` | No | Target date (`YYYY-MM-DD`) |

When introducing a new phase, add it to the `phases` list in `.taskmd.yaml` before assigning it to tasks.

### Other Settings

| Key | Values | Description |
|-----|--------|-------------|
| `workflow` | `solo` (default), `pr-review` | `pr-review`: tasks go through `in-review` with a linked PR |
| `worklogs` | `true` / `false` | Enable worklog entries |
| `id` | `sequential`, `prefixed`, `random`, `ulid` | ID generation strategy for `taskmd add` |

## Validation

A valid task **must**:

1. Have YAML frontmatter with `---` delimiters
2. Include `id` and `title`
3. Use valid enum values for `status`, `priority`, `effort`, `type`
4. Have unique IDs across the project
5. Reference only existing tasks in `dependencies` and `parent`
6. Have no circular dependencies or parent cycles

## Example

```markdown
---
id: "015"
title: "Implement user authentication"
status: in-progress
priority: high
effort: large
type: feature
phase: "v1.0"
dependencies: ["012", "013"]
tags: [auth, security, api]
created_at: 2026-02-08
verify:
  - type: bash
    run: "go test ./internal/auth/..."
  - type: assert
    check: "JWT tokens expire after 24 hours"
---

# Implement User Authentication

## Objective

Add JWT-based authentication to the API.

## Tasks

- [x] Design authentication flow
- [ ] Create login endpoint
- [ ] Add authentication middleware

## Acceptance Criteria

- Users can log in with email and password
- Protected routes require valid JWT
```
