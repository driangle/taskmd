# taskmd Specification Reference

This document is an embedded reference for the taskmd format. It covers everything the plugin needs to correctly read, write, and validate task files.

## Frontmatter Schema

Task files use YAML frontmatter delimited by `---` lines at the top of the file.

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique identifier for the task |
| `title` | string | Human-readable task title |

### Optional Fields

| Field | Type | Valid Values / Format | Description |
|-------|------|----------------------|-------------|
| `status` | string | `pending`, `in-progress`, `completed`, `in-review`, `blocked`, `cancelled` | Current task state (default: `pending`) |
| `priority` | string | `low`, `medium`, `high`, `critical` | Task priority level |
| `effort` | string | `small`, `medium`, `large` | Estimated effort size |
| `type` | string | `feature`, `bug`, `improvement`, `chore`, `docs` | Category of work |
| `dependencies` | array of strings | Task ID strings | Tasks that must be completed first |
| `tags` | array of strings | Lowercase, hyphen-separated | Freeform labels |
| `group` | string | Any string | Logical grouping (derived from directory if omitted) |
| `owner` | string | Any string | Who is responsible for the task |
| `phase` | string | Any string | Project phase this task belongs to |
| `touches` | array of strings | Scope identifiers | Areas of the codebase affected |
| `context` | array of strings | File paths | Files relevant to the task |
| `parent` | string | Single task ID | Parent task for subtask relationships |
| `created_at` | string | `YYYY-MM-DD` | Date the task was created (alias: `created`) |
| `verify` | array of objects | See below | Verification steps |
| `pr` | array of strings | URLs | Associated pull request links |
| `external_id` | string | Any string | ID in an external system |

### Verify Field Format

Each entry in the `verify` array is one of:

```yaml
verify:
  - type: bash
    run: "go test ./..."
    dir: "apps/cli"         # optional working directory
  - type: assert
    check: "All tests pass"  # human-readable assertion
```

## File Naming

Task files follow the pattern: `ID-descriptive-title.md`

Examples:
- `001-project-scaffolding.md`
- `cli-049-add-graph-command.md`
- `a3f9x2-implement-search.md`

The ID portion of the filename must match the `id` field in frontmatter.

## Directory Structure

```
tasks/
├── 001-root-task.md              # No group
├── web/                           # Group: "web"
│   └── 001-scaffolding.md
└── cli/                           # Group: "cli"
    └── 015-go-cli.md
```

### Group Resolution

Group is determined in this order:

1. Explicit `group` field in frontmatter
2. Parent directory name (e.g., `tasks/cli/` yields group `cli`)
3. No group (task is at the root of the tasks directory)

## .taskmd.yaml Configuration

Project-level configuration file at the repository root.

```yaml
dir: tasks           # task directory (default: tasks)
id:
  strategy: sequential  # sequential | prefixed | random | ulid
  prefix: ""            # for prefixed strategy
  length: 6             # for random/ulid
  padding: 3            # for sequential
workflow: solo          # solo | pr-review
worklogs: false         # true to allow writing worklog entries
phases:
  - id: phase-id
    name: "Phase Name"
    description: "What this phase covers"
    due: 2026-04-01
```

### Phases Configuration

Phases are time-based groupings (sprints, milestones, releases) defined in `.taskmd.yaml`. Tasks reference a phase by its `id`.

Each phase entry has:

| Field | Required | Description |
|-------|----------|-------------|
| `id` | no | Short stable identifier referenced by task `phase` field (falls back to `name` if omitted) |
| `name` | yes | Human-readable display label |
| `description` | no | What the phase covers |
| `due` | no | Target date (`YYYY-MM-DD`) |

**Adding a new phase:** Append an entry to the `phases` array in `.taskmd.yaml`:

```yaml
phases:
  - id: v2-launch
    name: "V2 Launch"
    description: "Features for the v2.0 release"
    due: 2026-06-01
```

**Assigning a task to a phase:** Set `phase: "<phase-id>"` in the task's frontmatter. The value must match a configured phase `id`. When phases are configured, referencing an undefined phase ID produces a validation warning.

## ID Generation Strategies

| Strategy | Format | Example Filename | Derived ID |
|----------|--------|-----------------|------------|
| sequential (default) | Zero-padded number | `009-add-feature.md` | `009` |
| prefixed | Prefix + number | `dr-001-fix-login.md` | `dr-001` |
| random | Alphanumeric | `a3f9x2-slug-title.md` | `a3f9x2` |
| ulid | ULID | `01h5a3mpk2-fix-bug.md` | `01h5a3mpk2` |

## Validation Rules

A valid task file must satisfy all of the following:

1. **Frontmatter present** -- YAML block between opening `---` and closing `---`
2. **Required fields** -- Both `id` and `title` must be present
3. **Valid enums** -- `status`, `priority`, `effort`, and `type` must use values from their respective allowed lists. The `effort` values above are defaults; a project may define its own ordered vocabulary via `effort:` in `.taskmd.yaml`, which replaces them entirely
4. **Unique IDs** -- No two task files may share the same `id`
5. **Valid dependencies** -- Every ID in `dependencies` must correspond to an existing task
6. **No circular dependencies** -- Following the dependency chain must not lead back to the same task
7. **Valid parent references** -- `parent` must reference an existing task, must not be self-referencing, and must not create cycles

## Workflow Modes

### solo (default)

Tasks can be set directly to `completed` status when finished.

### pr-review

When completing a task:
1. Set status to `in-review` (not `completed`)
2. Open a pull request
3. Stop -- a reviewer will set the task to `completed` after merge

## Next-Task Ranking Algorithm

To determine the next task to work on:

1. **Filter** to tasks with `status: pending`
2. **Exclude** tasks with unmet dependencies (any dependency whose status is not `completed`)
3. **Sort** by the following criteria in order:
   - Priority: `critical` > `high` > `medium` > `low`
   - Effort: lowest first — `small`, then `medium`, then `large` by default
   - Created date: oldest first

The first task in the sorted list is the recommended next task.

## Common Operations

### Finding a Task by ID

1. Use Glob to search for `tasks/**/*ID*.md` (where ID is the target)
2. Read the matched file and confirm the frontmatter `id` field matches
3. If no match is found, list available task IDs for the user

### Reading Frontmatter

1. Read the full file content
2. Find the first line that is exactly `---`
3. Find the second line that is exactly `---`
4. Parse the YAML between those two lines

### Finding the Task Directory

1. Read `.taskmd.yaml` at the project root
2. Use the `dir` field value
3. Default to `tasks` if the file does not exist or `dir` is not set
