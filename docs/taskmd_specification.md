# taskmd Specification

**Version:** 1.0
**Last Updated:** 2026-02-08

## Overview

The taskmd format is a markdown-based task definition system that uses YAML frontmatter to store structured metadata and markdown for rich task descriptions. Each task is stored as a separate `.md` file with a standardized schema.

This specification defines the conventions, formats, and best practices for creating and managing taskmd files.

## File Format

### Basic Structure

Every taskmd file consists of two parts:

1. **YAML Frontmatter** - Structured metadata enclosed in `---` delimiters
2. **Markdown Body** - Rich text content describing the task

```markdown
---
id: "001"
title: "Task title"
status: pending
priority: high
effort: medium
dependencies: []
tags:
  - tag1
  - tag2
created: 2026-02-08
---

# Task Title

## Objective

Description of what this task aims to accomplish.

## Tasks

- [ ] Subtask 1
- [ ] Subtask 2

## Acceptance Criteria

- Criterion 1
- Criterion 2
```

## Frontmatter Schema

### Required Fields

#### `id` (string)
- **Type:** String
- **Format:** Zero-padded numeric ID (e.g., `"001"`, `"015"`, `"142"`)
- **Description:** Unique identifier for the task within its scope
- **Constraints:** Must be unique across all tasks in the project
- **Example:** `id: "019"`

#### `title` (string)
- **Type:** String
- **Description:** Brief, descriptive title for the task
- **Constraints:** Should be concise and action-oriented
- **Example:** `title: "Implement user authentication"`

#### `status` (enum)
- **Type:** String (enum)
- **Valid Values:**
  - `pending` - Task has not been started
  - `in-progress` - Task is currently being worked on
  - `completed` - Task has been finished
  - `blocked` - Task cannot proceed due to dependencies or blockers
  - `cancelled` - Task will not be completed (kept for historical reference)
- **Description:** Current state of the task
- **Example:** `status: pending`

### Optional Fields

#### `priority` (enum)
- **Type:** String (enum)
- **Valid Values:**
  - `low` - Nice to have, can be deferred
  - `medium` - Standard priority
  - `high` - Important for project success
  - `critical` - Urgent, must be addressed immediately
- **Description:** Relative importance of the task
- **Default:** `medium` (if not specified)
- **Example:** `priority: high`

#### `effort` (enum)
- **Type:** String (enum)
- **Valid Values:**
  - `small` - Quick task, typically < 2 hours
  - `medium` - Moderate task, typically 2-8 hours
  - `large` - Substantial task, typically > 8 hours or multi-day
- **Description:** Estimated effort or complexity
- **Example:** `effort: medium`

#### `dependencies` (array)
- **Type:** Array of strings
- **Description:** List of task IDs that must be completed before this task can start
- **Format:** Array of task ID strings
- **Default:** `[]` (empty array)
- **Example:**
  ```yaml
  dependencies:
    - "001"
    - "015"
  ```

#### `tags` (array)
- **Type:** Array of strings
- **Description:** Labels for categorization and filtering
- **Format:** Array of lowercase, hyphen-separated strings
- **Default:** `[]` (empty array)
- **Conventions:**
  - Use lowercase
  - Use hyphens for multi-word tags (e.g., `user-auth`)
  - Common tags: `setup`, `infrastructure`, `core`, `ui`, `testing`, `documentation`
- **Example:**
  ```yaml
  tags:
    - core
    - api
    - authentication
  ```

#### `group` (string)
- **Type:** String
- **Description:** Logical grouping for the task
- **Behavior:**
  - If specified in frontmatter, uses the explicit value
  - If omitted, derived from the parent directory name
  - Root-level tasks have no group
- **Example:** `group: "cli"`

#### `created` (date)
- **Type:** Date string
- **Format:** ISO 8601 date format (`YYYY-MM-DD`)
- **Description:** Date when the task was created
- **Example:** `created: 2026-02-08`

#### `description` (string)
- **Type:** String
- **Description:** Optional brief description (alternative to markdown body)
- **Usage:** Useful for task lists where full markdown is unnecessary
- **Example:** `description: "Set up CI/CD pipeline with GitHub Actions"`

### Extension Fields

Implementations MAY support additional custom fields. Unknown fields SHOULD be preserved during read/write operations to maintain forward compatibility.

## File Naming Conventions

### Standard Pattern

Task files SHOULD follow this naming pattern:

```
NNN-descriptive-title.md
```

Where:
- `NNN` is the zero-padded task ID (e.g., `001`, `042`, `137`)
- `descriptive-title` is a lowercase, hyphen-separated slug derived from the task title
- `.md` is the markdown file extension

### Examples

- `001-project-scaffolding.md`
- `015-go-cli-scaffolding.md`
- `019-validate-command.md`
- `142-implement-user-authentication.md`

### Alternative Pattern

For projects preferring semantic naming, the ID prefix MAY be omitted:

```
descriptive-title.md
```

In this case, the `id` field in frontmatter becomes the sole source of task identification.

## Directory Structure

### Overview

Tasks can be organized into subdirectories for logical grouping. The directory structure provides implicit grouping that complements the frontmatter `group` field.

### Group Resolution

The `group` field is resolved using the following priority:

1. **Explicit frontmatter** - If `group` is specified in frontmatter, use that value
2. **Directory-based** - If `group` is omitted, derive from parent directory name
3. **No group** - Root-level tasks have no group (empty/null)

### Example Structure

```
tasks/
├── 001-taskmd-specification.md     # Root task, no group
├── web/                             # Group: "web"
│   ├── 001-project-scaffolding.md
│   ├── 002-typescript-types.md
│   └── 003-markdown-parsing.md
└── cli/                             # Group: "cli"
    ├── 015-go-cli-scaffolding.md
    ├── 016-task-model-markdown-parsing.md
    ├── archive/                     # Group: "cli" (inherited from parent)
    │   └── 018-file-watcher.md
    └── README.md                    # Non-task file (ignored)
```

### Directory Conventions

- **Flat structure** - All tasks in a single `tasks/` directory
- **Grouped structure** - Tasks organized by feature area, component, or milestone
- **Mixed structure** - Root-level tasks + grouped subdirectories
- **Archive subdirectories** - Use `archive/` or similar for completed/deprecated tasks

## Dependencies

### Reference Format

Dependencies reference other tasks by their `id` field:

```yaml
dependencies:
  - "001"
  - "015"
```

### Dependency Rules

1. **ID-based references** - Always reference by task ID, never by filename
2. **String format** - IDs should be quoted strings in YAML
3. **Array format** - Always use array syntax, even for single dependencies
4. **Cross-group dependencies** - Tasks can depend on tasks in different groups
5. **Circular dependencies** - MUST NOT create circular dependency chains
6. **Missing dependencies** - Tools SHOULD warn about references to non-existent tasks

### Validation

The `validate` command checks for:
- Circular dependency cycles
- Missing task references
- Self-dependencies (task depending on itself)

## Status Lifecycle

### Status Flow

```
pending → in-progress → completed
   ↓            ↓            ↓
   ↓         blocked        ↓
   ↓            ↓           ↓
   └──→ cancelled ←─────────┘
```

### Status Definitions

| Status | Description | When to Use |
|--------|-------------|-------------|
| `pending` | Not started | Initial state for new tasks |
| `in-progress` | Actively being worked on | When work begins |
| `completed` | Finished and verified | When all acceptance criteria are met |
| `blocked` | Cannot proceed | When dependencies are incomplete or blockers exist |
| `cancelled` | Will not be completed | When requirements change, task is superseded, or permanently deprioritized |

### Status Transitions

- `pending` → `in-progress` - Work begins
- `in-progress` → `completed` - Work finishes successfully
- `in-progress` → `blocked` - Blocker encountered
- `blocked` → `pending` - Blocker resolved, not yet restarted
- `blocked` → `in-progress` - Blocker resolved, work resumes
- `pending` → `blocked` - Blocker identified before work begins
- `pending` → `cancelled` - Task deprioritized or requirements changed before work begins
- `in-progress` → `cancelled` - Active work stopped, task will not be completed
- `blocked` → `cancelled` - Blocked task will not be unblocked or completed
- `completed` → `cancelled` - Rare; used when task was incorrectly marked as completed

### Best Practices

- Update status immediately when work begins or completes
- Use `blocked` with a comment explaining the blocker
- Consider using subtasks in markdown body for granular progress tracking

## Priority Levels

### Priority Scale

| Priority | Use Case | Examples |
|----------|----------|----------|
| `low` | Nice to have, no deadline | Documentation improvements, minor refactors |
| `medium` | Standard work items | Feature development, routine bug fixes |
| `high` | Important for project success | Core features, blocking issues |
| `critical` | Urgent, severe impact | Production outages, security vulnerabilities |

### Priority Guidelines

- Use `critical` sparingly - only for true emergencies
- Default to `medium` for standard development work
- Reserve `high` for tasks that block other work or are on critical path
- Use `low` for backlog items that can be deferred indefinitely

## Effort Estimates

### Effort Scale

| Effort | Typical Duration | Characteristics |
|--------|------------------|-----------------|
| `small` | < 2 hours | Simple changes, clear scope, minimal uncertainty |
| `medium` | 2-8 hours | Moderate complexity, some unknowns, multiple steps |
| `large` | > 8 hours or multi-day | Complex, significant unknowns, requires design/planning |

### Estimation Guidelines

- Focus on complexity, not just time
- Consider uncertainty and risk
- Break down `large` tasks into smaller ones when possible
- Re-estimate if scope changes significantly

## Tag Conventions

### Naming Patterns

- **Lowercase** - Always use lowercase letters
- **Hyphens** - Use hyphens for multi-word tags (e.g., `user-auth`, `api-client`)
- **Singular** - Prefer singular form (e.g., `test` not `tests`)
- **Descriptive** - Choose clear, meaningful names

### Common Tags

#### By Type
- `setup` - Initial configuration and scaffolding
- `infrastructure` - Build tools, deployment, environment
- `core` - Core business logic or functionality
- `ui` - User interface components
- `api` - API endpoints or integration
- `test` - Testing code or infrastructure
- `documentation` - Docs, specs, README files

#### By Technology
- `go` - Go language tasks
- `typescript` - TypeScript/JavaScript tasks
- `react` - React framework
- `cli` - Command-line interface

#### By Activity
- `refactor` - Code restructuring
- `bugfix` - Bug fixes
- `enhancement` - Improvements to existing features
- `feature` - New functionality

### Best Practices

- Limit to 2-5 tags per task
- Be consistent across the project
- Use tags that enable useful filtering
- Document project-specific tag conventions

## Markdown Body

### Structure Recommendations

While the markdown body is free-form, these sections are commonly used:

#### Minimal Structure
```markdown
# Task Title

Brief description of what needs to be done.
```

#### Standard Structure
```markdown
# Task Title

## Objective

What this task aims to accomplish and why it matters.

## Tasks

- [ ] Subtask 1
- [ ] Subtask 2
- [ ] Subtask 3

## Acceptance Criteria

- Criterion 1
- Criterion 2
```

#### Comprehensive Structure
```markdown
# Task Title

## Objective

What this task aims to accomplish.

## Context

Background information, related work, or design decisions.

## Tasks

- [ ] Step 1
- [ ] Step 2
- [ ] Step 3

## Acceptance Criteria

- Must satisfy requirement 1
- Must satisfy requirement 2

## Testing

How to verify the task is complete.

## Notes

Additional information, gotchas, or considerations.

## References

- [Related documentation](url)
- [Design doc](url)
```

### Formatting Guidelines

- **Heading hierarchy** - Start with `#` (h1) for the main title
- **Checklists** - Use `- [ ]` for subtasks
- **Code blocks** - Use triple backticks with language identifiers
- **Links** - Use standard markdown link syntax
- **Emphasis** - Use `**bold**` for important terms, `*italic*` for emphasis
- **Lists** - Use `-` for unordered, `1.` for ordered

## Examples

### Minimal Task

```markdown
---
id: "001"
title: "Fix login button alignment"
status: pending
---

# Fix Login Button Alignment

The login button on the homepage is misaligned by 2px. Update the CSS to center it properly.
```

### Standard Task

```markdown
---
id: "015"
title: "Implement user authentication"
status: in-progress
priority: high
effort: large
dependencies: ["012", "013"]
tags:
  - auth
  - security
  - api
created: 2026-02-08
---

# Implement User Authentication

## Objective

Add JWT-based authentication to the API with login, logout, and token refresh endpoints.

## Tasks

- [x] Design authentication flow
- [x] Implement JWT signing and verification
- [ ] Create `/auth/login` endpoint
- [ ] Create `/auth/logout` endpoint
- [ ] Create `/auth/refresh` endpoint
- [ ] Add authentication middleware
- [ ] Write integration tests

## Acceptance Criteria

- Users can log in with email and password
- JWT tokens expire after 24 hours
- Refresh tokens work for 30 days
- Protected routes require valid JWT
- All endpoints have > 90% test coverage
```

### Task with Complex Body

```markdown
---
id: "042"
title: "Migrate database to PostgreSQL"
status: blocked
priority: critical
effort: large
dependencies: ["040", "041"]
tags:
  - database
  - migration
  - infrastructure
created: 2026-02-01
---

# Migrate Database to PostgreSQL

## Objective

Migrate the application database from SQLite to PostgreSQL for production scalability and better concurrent access handling.

## Context

Current SQLite implementation works for development but has limitations:
- No concurrent write access
- Limited scalability
- No production-grade replication

PostgreSQL will provide:
- Better concurrent access
- Advanced querying capabilities
- Production-ready replication and backups

## Tasks

- [x] Set up local PostgreSQL instance
- [x] Create migration scripts
- [ ] Update database schema for PostgreSQL-specific types
- [ ] Migrate seed data
- [ ] Update ORM configuration
- [ ] Run full test suite against PostgreSQL
- [ ] Create backup/restore procedures
- [ ] Deploy to staging environment
- [ ] Performance testing
- [ ] Production cutover plan

## Blockers

**Status:** Blocked on infrastructure provisioning (task 040)

The RDS instance needs to be provisioned before migration scripts can be tested against a production-like environment.

## Acceptance Criteria

- All data migrated without loss
- Application runs on PostgreSQL with no functionality regression
- Test suite passes at 100%
- Backup procedures documented and tested
- Rollback plan documented and tested
- Performance meets or exceeds SQLite baseline

## Testing

1. Run migration on copy of production data
2. Verify data integrity with checksums
3. Run full application test suite
4. Perform load testing
5. Test backup and restore procedures
6. Test rollback procedures

## Rollback Plan

If issues occur:
1. Stop application
2. Restore SQLite database from backup
3. Revert application configuration
4. Restart application
5. Document issues for next attempt

## References

- [Migration Guide](https://example.com/docs/migration)
- [PostgreSQL Best Practices](https://example.com/docs/pg-practices)
- [Rollback Procedures](https://example.com/docs/rollback)
```

## Validation Rules

A valid taskmd file MUST:

1. **Have YAML frontmatter** - Enclosed in `---` delimiters
2. **Include required fields** - `id`, `title`, `status`
3. **Use valid enum values** - For `status`, `priority`, `effort`
4. **Have unique IDs** - No duplicate IDs within the project
5. **Have valid dependencies** - All referenced task IDs must exist
6. **Have no circular dependencies** - Dependency graph must be acyclic

A valid taskmd file SHOULD:

1. **Follow naming conventions** - Use `NNN-task-name.md` pattern
2. **Include creation date** - Track when tasks were created
3. **Have descriptive titles** - Clear, action-oriented titles
4. **Include acceptance criteria** - Define "done" clearly
5. **Use appropriate granularity** - Not too broad, not too narrow

## Best Practices

### Task Granularity

- **Too small** - "Change variable name" (too trivial)
- **Too large** - "Build entire application" (too broad)
- **Just right** - "Implement user login endpoint" (focused, achievable)

**Guidelines:**
- Tasks should be completable in a reasonable time frame (hours to a few days)
- Each task should have a clear, single objective
- Break down large tasks into smaller ones with dependencies
- Use subtasks in markdown for fine-grained tracking

### Dependency Management

- Keep dependency chains shallow when possible
- Avoid long linear chains (A → B → C → D → E)
- Parallelize work by minimizing dependencies
- Document why dependencies exist
- Review and remove unnecessary dependencies

### Status Updates

- Update status as soon as work begins
- Mark as `completed` only when acceptance criteria are fully met
- Use `blocked` proactively to surface issues
- Consider adding comments when changing status

### File Organization

- **Small projects** - Flat structure in `tasks/` directory
- **Medium projects** - Group by feature area or component
- **Large projects** - Hierarchical grouping with clear categories
- **Archive completed tasks** - Move to `archive/` or mark with tag

### Writing Task Descriptions

- Start with a clear objective
- List concrete, actionable subtasks
- Define specific acceptance criteria
- Include relevant context and links
- Keep it concise but complete
- Use checklists for tracking progress

### Working with Tags

- Establish a tag taxonomy early
- Document standard tags in project README
- Review and consolidate tags periodically
- Use tags for filtering, not detailed categorization
- Prefer tags over complex group hierarchies

## Tooling Integration

### CLI Commands

The taskmd CLI provides commands that work with this specification:

- `taskmd validate` - Lint and validate tasks against this spec
- `taskmd list` - List tasks with filtering
- `taskmd stats` - Show project metrics
- `taskmd graph` - Visualize dependency graph
- `taskmd snapshot` - Export task state

### Validation

Validation tools should check:

1. **Schema compliance** - Required fields present, correct types
2. **Enum values** - Valid status, priority, effort values
3. **ID uniqueness** - No duplicate task IDs
4. **Dependency integrity** - All referenced tasks exist
5. **Acyclic dependencies** - No circular dependency chains
6. **File naming** - Follows conventions (warning, not error)

### Parsing

Parsers should:

1. **Preserve unknowns** - Keep unrecognized frontmatter fields
2. **Handle missing optionals** - Provide sensible defaults
3. **Preserve body** - Keep markdown content intact
4. **Round-trip fidelity** - parse(serialize(task)) ≈ task
5. **Error gracefully** - Return clear errors for malformed files

## Versioning

This specification follows semantic versioning:

- **Major version** - Breaking changes to schema or behavior
- **Minor version** - New optional fields or non-breaking additions
- **Patch version** - Clarifications and corrections

Current version: **1.0**

## Future Considerations

Potential future enhancements under consideration:

- **Assignees** - Track who is working on a task
- **Due dates** - Add deadline tracking
- **Custom fields** - Project-specific metadata
- **Links** - Bidirectional task relationships beyond dependencies
- **Subtask references** - Link to external task files as subtasks
- **Version history** - Track task changes over time

## Appendix: YAML Frontmatter Reference

### Complete Example

```yaml
---
id: "042"
title: "Implement data export feature"
status: in-progress
priority: high
effort: medium
dependencies:
  - "038"
  - "040"
tags:
  - feature
  - export
  - api
group: "backend"
created: 2026-02-08
description: "Add CSV and JSON export functionality for task data"
---
```

### Field Types

| Field | Type | Required | Default | Format |
|-------|------|----------|---------|--------|
| `id` | string | Yes | - | Zero-padded numbers |
| `title` | string | Yes | - | Free text |
| `status` | enum | Yes | - | pending, in-progress, completed, blocked, cancelled |
| `priority` | enum | No | medium | low, medium, high, critical |
| `effort` | enum | No | - | small, medium, large |
| `dependencies` | array | No | [] | Array of ID strings |
| `tags` | array | No | [] | Array of strings |
| `group` | string | No | (from dir) | Free text |
| `created` | date | No | - | YYYY-MM-DD |
| `description` | string | No | - | Free text |

## License

This specification is released under the [Creative Commons CC0 1.0 Universal](https://creativecommons.org/publicdomain/zero/1.0/) license.

---

**Maintained by:** md-task-tracker project
**Contributions:** Submit issues or PRs to improve this specification
