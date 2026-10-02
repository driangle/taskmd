# taskmd-lite -- CLI-free Claude Code Plugin

A zero-dependency taskmd plugin that uses Claude's native tools (Read, Write, Edit, Glob, Grep) instead of the taskmd CLI binary. No installation of Go, no compiled binaries -- just plain markdown task management powered by Claude Code's built-in capabilities.

## Prerequisites

None. This plugin requires no CLI binary, no runtime, and no external dependencies.

Use it when the `taskmd` binary cannot be installed. If you can install the CLI, prefer the
[`taskmd`](../claude-code-plugin/README.md) plugin: it exposes the same skills but delegates
ranking, validation, and graph analysis to the CLI instead of re-deriving them from the files.
Do not install both; they provide the same skills under different prefixes. See the
[plugin guide](https://driangle.github.io/taskmd/guide/claude-code-plugin) for a comparison.

## Installation

```bash
claude plugin marketplace add driangle/taskmd
claude plugin install taskmd-lite@taskmd-marketplace --scope project
```

## Available Skills

| Skill | Description | Example |
|-------|-------------|---------|
| `list-tasks` | List all tasks with optional filtering by status, group, or priority | `/taskmd-lite:list-tasks --status pending` |
| `get-task` | Retrieve a single task by its ID, showing full frontmatter and body | `/taskmd-lite:get-task 042` |
| `get-task-status` | Get just the status of a task by ID | `/taskmd-lite:get-task-status 042` |
| `next-task` | Find the highest-priority pending task with all dependencies met | `/taskmd-lite:next-task` |
| `add-task` | Create a new task file with generated ID and frontmatter | `/taskmd-lite:add-task Add search feature, high priority` |
| `update-task` | Modify frontmatter fields on an existing task, described in plain language | `/taskmd-lite:update-task set 042 to high priority and in-progress` |
| `complete-task` | Confirm the task's subtasks and acceptance criteria are met (running its `verify` checks if it has any), then set `status: completed`; in pr-review mode sets `in-review` instead | `/taskmd-lite:complete-task 042` |
| `validate-tasks` | Check all task files for schema errors, broken deps, and circular refs | `/taskmd-lite:validate-tasks` |
| `verify-task` | Run a task's `verify` checks (shell commands via Bash, file and content assertions) and report pass/fail | `/taskmd-lite:verify-task 042` |
| `do-task` | Look up a task by ID or name and work on it end-to-end, checking off subtasks and keeping the worklog as it goes | `/taskmd-lite:do-task 042` |
| `split-task` | Break a large task into smaller subtasks | `/taskmd-lite:split-task 042` |
| `divide-and-conquer` | Work on a task by splitting it into independent workstreams, each run by a subagent in its own git worktree and branch (`dnc/<id>/<slug>`); asks before anything is merged | `/taskmd-lite:divide-and-conquer 042` |
| `import-todos` | Scan source files for TODO/FIXME comments and create tasks from the ones you pick | `/taskmd-lite:import-todos --dir ./src` |

## How It Works

Task management itself runs through Claude's native file tools, with no `taskmd` binary:

1. **Glob** finds task files matching `tasks/**/*.md` patterns
2. **Read** parses YAML frontmatter and markdown body from each file
3. **Edit** and **Write** modify frontmatter fields or create new task files
4. **Grep** searches across task content for filtering and validation

Listing, filtering, sorting, dependency resolution, and validation are performed by Claude directly from file contents and never run a shell command.

Four skills do more than manage task files, and are allowed **Bash**: `do-task` and `divide-and-conquer` carry out the work the task describes, `verify-task` runs the task's `verify` commands, and `complete-task` runs them before marking the task done. `divide-and-conquer` also runs `git` to create worktrees and branches for its subagents. Each skill's `allowed-tools` line in its `SKILL.md` is the exact list.

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

## Versioning

This plugin is on its **own `0.x` semver line**, independent of both the taskmd CLI
release number and the other marketplace plugins. Because it runs no binary, it moves
when the **spec or the prose** moves — a CLI release that does not touch this directory
does not bump it. See
[ADR 0003](https://github.com/driangle/taskmd/blob/main/docs/adr/0003-plugin-versioning-policy.md).

Pre-1.0, skill names and their arguments are not yet a stability promise:

- **Patch** — new skills, new flags on existing skills, prose corrections that bring a
  skill back in line with the CLI.
- **Minor** — a skill renamed or removed, or the arguments it accepts changed.

The authoritative version is the `version` field in
[`.claude-plugin/plugin.json`](./.claude-plugin/plugin.json). The marketplace manifest
deliberately does not repeat it.
