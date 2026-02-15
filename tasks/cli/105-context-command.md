---
id: "105"
title: "Add context command for AI agent task context"
status: pending
priority: high
effort: large
tags:
  - ai
  - dx
  - mvp
touches:
  - cli
created: 2026-02-14
---

# Add Context Command for AI Agent Task Context

## Objective

Add a `taskmd context --task-id <ID>` command that resolves all relevant files for a task into a single structured output. The infrastructure already exists — `touches` maps to scopes, scopes map to file paths in `.taskmd.yaml` — but this data is only used internally by validation and tracks. This command surfaces it to agents and humans, turning taskmd into the bridge between "what to do" and "where to look."

Pair this with a new `context` frontmatter field for explicit file references that fall outside of scope mappings (test fixtures, docs, config files):

```yaml
---
id: "042"
title: "Add pagination to task list API"
touches:
  - cli
context:
  - "apps/cli/internal/web/handlers.go"
  - "apps/cli/internal/web/handlers_test.go"
  - "docs/api-design.md"
---
```

The command merges files from both sources (automatic scope resolution + explicit `context` field), deduplicates by path, checks existence, and tags each entry with its source (`scope:<name>` or `explicit`).

## Tasks

### Specification & Model
- [ ] Add `context` field to `docs/taskmd_specification.md` as an optional array of relative file path strings
- [ ] Add `Context []string` to the Task struct in `internal/model/task.go` with `yaml:"context" json:"context,omitempty"`

### Context Resolution Package
- [ ] Create `internal/context/resolve.go`:
  - Accept a task + scope config, return a unified file list
  - Resolve `touches` → scope paths, merge with explicit `context` entries
  - Deduplicate by path, check file/directory existence
  - Support directory expansion (glob files within directory paths)
  - Support content inlining (read file contents into output)
- [ ] Add tests in `internal/context/resolve_test.go`

### CLI Command
- [ ] Create `internal/cli/context.go` with the `context` command
- [ ] Flags:
  - `--task-id` (required) — task to build context for
  - `--format` — `text` (default), `json`, `yaml`
  - `--resolve` — expand directory paths to individual files
  - `--include-content` — inline file contents into output
  - `--include-deps` — also include files from dependency tasks
  - `--max-files` — cap number of files returned (default: no limit)
- [ ] Text output: grouped by source (scope files, explicit files, dependencies)
- [ ] JSON output: flat `files` array with `path`, `source`, `exists`, and optional `content`/`lines` fields
- [ ] Register command with `rootCmd`
- [ ] Add tests in `internal/cli/context_test.go`

### Integration with `get`
- [ ] Add `--context` flag to `taskmd get` that appends the context file list to normal output
- [ ] In JSON/YAML mode, include as a `context_files` field

### Validation
- [ ] Warn (not error) when `context` references a non-existent file — the task may create it

### Skill Update
- [ ] Update `claude-code-plugin/skills/do-task/SKILL.md` to use `taskmd context` before starting work

## Example Output

**Text (default):**

```
Context for task 042 (Add pagination to task list API)

Scope files (cli):
  apps/cli/internal/web/handlers.go
  apps/cli/internal/web/handlers_test.go
  apps/cli/internal/cli/list.go

Explicit files:
  docs/api-design.md

Dependencies:
  038 — Implement API router          (completed)
  041 — Add JSON response helpers     (completed)
```

**JSON** (`--format json --include-content`):

```json
{
  "task_id": "042",
  "title": "Add pagination to task list API",
  "task_body": "## Objective\n\nAdd pagination support to...",
  "files": [
    {
      "path": "apps/cli/internal/web/handlers.go",
      "source": "scope:cli",
      "exists": true,
      "content": "package web\n\nimport (\n...",
      "lines": 142
    },
    {
      "path": "docs/api-design.md",
      "source": "explicit",
      "exists": true,
      "content": "# API Design\n\n...",
      "lines": 45
    }
  ],
  "dependencies": [
    {"id": "038", "title": "Implement API router", "status": "completed"}
  ]
}
```

`content` and `lines` fields are only present with `--include-content`. `task_body` is only present with `--include-content`.

## Non-Goals

- No code intelligence or automatic inference — context comes from explicit human declarations (touches, scopes, context field) only
- No cross-repo resolution

## Acceptance Criteria

- `taskmd context --task-id 042` merges files from both `touches` scope resolution and explicit `context` field into a deduplicated list
- Each file entry is tagged with its source (`scope:<name>` or `explicit`) and existence status
- `--resolve` expands directory paths to individual files
- `--include-content` inlines file contents and task body
- `--include-deps` includes files from dependency tasks' scopes/context
- `taskmd get 042 --context` appends context to normal get output
- Gracefully handles: no scopes config, missing files, no touches, no context field
- Tests cover all flags, both context sources, output formats, and edge cases
