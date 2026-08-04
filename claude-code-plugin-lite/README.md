# taskmd-lite -- CLI-free Claude Code Plugin

A zero-dependency taskmd plugin that uses Claude's native tools (Read, Write, Edit, Glob, Grep) instead of the taskmd CLI binary. No installation of Go, no compiled binaries -- just plain markdown task management powered by Claude Code's built-in capabilities.

## Prerequisites

None. This plugin requires no CLI binary, no runtime, and no external dependencies.

## Installation

```bash
claude plugin install --marketplace https://github.com/driangle/taskmd
```

Then select `taskmd-lite` from the list of available plugins.

## Available Skills

| Skill | Description | Example |
|-------|-------------|---------|
| `list-tasks` | List all tasks with optional filtering by status, group, or priority | `/taskmd-lite:list-tasks --status pending` |
| `get-task` | Retrieve a single task by its ID, showing full frontmatter and body | `/taskmd-lite:get-task 042` |
| `get-task-status` | Get just the status of a task by ID | `/taskmd-lite:get-task-status 042` |
| `next-task` | Find the highest-priority pending task with all dependencies met | `/taskmd-lite:next-task` |
| `add-task` | Create a new task file with generated ID and frontmatter | `/taskmd-lite:add-task --title "Add search feature" --priority high` |
| `update-task` | Modify frontmatter fields on an existing task | `/taskmd-lite:update-task 042 --status in-progress` |
| `complete-task` | Mark a task as completed and check off all subtasks | `/taskmd-lite:complete-task 042` |
| `validate-tasks` | Check all task files for schema errors, broken deps, and circular refs | `/taskmd-lite:validate-tasks` |
| `verify-task` | Run a task's verify steps (bash commands, assertions) to confirm completion | `/taskmd-lite:verify-task 042` |
| `do-task` | Pick up the next task and start working on it end-to-end | `/taskmd-lite:do-task` |
| `split-task` | Break a large task into smaller subtasks | `/taskmd-lite:split-task 042` |
| `divide-and-conquer` | Recursively decompose a task tree into actionable units | `/taskmd-lite:divide-and-conquer 042` |
| `import-todos` | Scan source files for TODO/FIXME comments and create tasks from them | `/taskmd-lite:import-todos src/` |

## How It Works

This plugin operates entirely through Claude's native file tools:

1. **Glob** finds task files matching `tasks/**/*.md` patterns
2. **Read** parses YAML frontmatter and markdown body from each file
3. **Edit** and **Write** modify frontmatter fields or create new task files
4. **Grep** searches across task content for filtering and validation

No shell commands are executed. All task operations -- listing, filtering, sorting, dependency resolution, validation -- are performed by Claude directly using file contents.

## Specification Reference

For the full taskmd format specification including frontmatter schema, ID strategies, validation rules, and configuration options, see [SPEC_REFERENCE.md](./SPEC_REFERENCE.md).

## The CLI is the source of truth

Because this plugin runs without the Go binary, its skills re-express core CLI
algorithms as English prose — ID generation (sequential / prefixed / random /
ULID), slug rules, the new-task frontmatter template, and the validation enums
all live here a second time. **When the prose and the CLI disagree, the CLI
wins.** The prose is a derived copy that must follow the CLI's behavior, never
the other way around.

Two Go test suites keep this contract honest, so drift fails the build (they run
in CI via `go test ./...` from `apps/cli`):

- `apps/cli/internal/cli/spec_reference_test.go` guards
  [`SPEC_REFERENCE.md`](./SPEC_REFERENCE.md) against the canonical spec document
  (enum sets and required fields).
- `apps/cli/internal/cli/lite_conformance_test.go` guards the **skill prose**
  against the CLI's actual runtime behavior: it derives each fact from real CLI
  code (the `nextid`/`slug` SDK functions, the `add` command's frontmatter
  writer, the in-package enum lists) and asserts the matching `SKILL.md`
  documents it.

If a conformance test fails, update the named `SKILL.md` so its prose matches
the current CLI behavior — do not weaken the test.
