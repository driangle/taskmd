# Core Concepts

## Tasks

Tasks are markdown files with YAML frontmatter. Each task file has two parts:

1. **Frontmatter** - structured metadata enclosed in `---` delimiters
2. **Body** - markdown content for descriptions, subtasks, and acceptance criteria

```markdown
---
id: "001"
title: "Implement feature X"
status: pending
priority: high
effort: medium
dependencies: []
tags:
  - feature
  - backend
created: 2026-02-08
---

# Implement Feature X

## Objective
Build the new feature X that allows users to...

## Tasks
- [ ] Design API endpoints
- [ ] Implement backend logic
- [ ] Write tests
```

### Required Fields

- **`id`** - Unique identifier, typically zero-padded (e.g., `"001"`, `"042"`)
- **`title`** - Brief, action-oriented description

### Optional Fields

- **`status`** - Current state of the task
- **`priority`** - Importance level (`low`, `medium`, `high`, `critical`)
- **`effort`** - Estimated complexity (`small`, `medium`, `large`)
- **`dependencies`** - List of task IDs that must complete first
- **`tags`** - Labels for categorization and filtering
- **`group`** - Logical grouping (derived from directory if omitted)
- **`created`** - Creation date in `YYYY-MM-DD` format

## Status

Tasks move through these states:

| Status | Meaning |
|--------|---------|
| `pending` | Not started (initial state) |
| `in-progress` | Currently being worked on |
| `completed` | Finished and verified |
| `blocked` | Cannot proceed due to a blocker |
| `cancelled` | Will not be completed |

Typical flow:

```
pending → in-progress → completed
   ↓            ↓            ↓
   ↓         blocked        ↓
   ↓            ↓           ↓
   └──→ cancelled ←─────────┘
```

## Dependencies

Tasks can depend on other tasks. A dependency means "this task cannot start until the dependency is completed."

```yaml
dependencies:
  - "001"  # Must complete task 001 first
  - "005"  # And task 005
```

Dependencies create a directed acyclic graph (DAG) that taskmd uses to:
- **Recommend next tasks** - only suggests tasks with satisfied dependencies
- **Visualize relationships** - interactive dependency graphs
- **Calculate critical paths** - identify the longest chain of dependent tasks
- **Find blockers** - which tasks are preventing the most work

## Priority and Effort

These are independent dimensions for planning:

| Priority | Use Case |
|----------|----------|
| `low` | Nice to have, can be deferred |
| `medium` | Standard work items |
| `high` | Important for project success |
| `critical` | Urgent, must address immediately |

| Effort | Typical Duration |
|--------|------------------|
| `small` | Less than 2 hours |
| `medium` | 2-8 hours |
| `large` | More than 8 hours / multi-day |

A task can be high priority but small effort (urgent bug fix), or low priority but large effort (nice-to-have feature). The `next` command uses both to make intelligent recommendations.

## Tags

Tags are labels for categorization and filtering:

```yaml
tags:
  - feature
  - backend
  - api
```

Use lowercase, hyphen-separated strings. Keep your tag vocabulary consistent across tasks.

## File Organization

### File Naming

Task files follow the pattern `NNN-descriptive-title.md`:

```
tasks/
├── 001-project-setup.md
├── 002-user-authentication.md
└── 003-api-endpoints.md
```

### Directory Structure

Use subdirectories to organize by area:

```
tasks/
├── 001-specification.md        # Root-level task
├── cli/                         # Group: "cli"
│   ├── 015-go-cli-scaffolding.md
│   └── 016-task-parsing.md
└── web/                         # Group: "web"
    ├── 010-board-view.md
    └── 011-graph-view.md
```

The `group` field is automatically derived from the parent directory name.

## Task Discovery

taskmd scans directories recursively for `.md` files with valid task frontmatter:

```bash
# Scan specific directory
taskmd list ./tasks

# Scan subdirectory
taskmd list ./tasks/cli
```

## Next Steps

- [CLI Guide](/guide/cli) - Learn all CLI commands
- [Task Specification](/reference/specification) - Full format reference
- [Configuration](/reference/configuration) - Set up your project
