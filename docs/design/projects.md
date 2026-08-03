# Global Project Registry

**Status:** Draft
**Date:** 2026-03-22

## Problem

taskmd is directory-scoped. `initConfig()` walks up from `cwd` to find the nearest `.taskmd.yaml`, then falls back to `~/.taskmd.yaml`. You must `cd` into a project (or pass `--task-dir`) to interact with it. There's no way to say "show me all my projects" or "what's next across everything I'm working on" from an arbitrary directory.

## Goals

1. Register project roots in a global config so taskmd knows where they live
2. List, filter, and query tasks across registered projects from anywhere
3. Keep it simple — a project is just a directory with a `.taskmd.yaml`
4. No changes to the task file format, no new frontmatter fields

## Non-goals

- Sub-project groupings within a single repo (use existing `group`, `tags`, or `phase` for that)
- Cross-project dependencies
- Project-scoped phases or scopes (each project already has its own `.taskmd.yaml` for that)
- Monorepo workspace support (see `monorepo-workspaces.md`)

## Design

### Config

The home directory config (`~/.taskmd.yaml`) gains a `projects` section — a list of known project roots on disk:

```yaml
# ~/.taskmd.yaml
projects:
  - id: taskmd
    name: "taskmd"
    path: ~/workplace/gg/taskmd/taskmd-1
  - id: myapp
    name: "My App"
    path: ~/workplace/myapp
  - id: dotfiles
    name: "Dotfiles"
    path: ~/dotfiles
```

| Field | Required | Description |
|-------|----------|-------------|
| `id` | No | Short identifier for `--project` flag (falls back to directory basename if omitted) |
| `name` | No | Human-readable label (falls back to `id` if omitted) |
| `path` | Yes | Absolute or `~`-relative path to the project root (directory containing `.taskmd.yaml`) |

Each entry is a pointer. The project's full config (phases, scopes, ID strategy, task dir) comes from its own `.taskmd.yaml` — the global registry just knows where to find it.

### Config resolution

Current behavior is unchanged for local use:
1. Walk up from `cwd` to find nearest `.taskmd.yaml`
2. Fall back to `~/.taskmd.yaml`

The global registry is only consulted when:
- Running `taskmd projects` (always lists registered projects)
- Using `--project <id>` while not inside that project's directory tree
- Using `--all-projects`

When `--project <id>` is used:
1. If `cwd` is inside a registered project's path, use local config (no registry lookup needed)
2. Otherwise, find the matching entry in the global registry, resolve its `path`, and load that project's `.taskmd.yaml`

### What a "project" is

A project is a directory that contains a `.taskmd.yaml`. Nothing more. The registry doesn't add metadata beyond a short id and a path. Everything else — phases, scopes, workflow, ID strategy — lives in the project's own config file where it belongs.

## CLI changes

### `taskmd projects`

List all registered projects with summary stats:

```bash
$ taskmd projects
PROJECT   PATH                               TASKS  PENDING  IN-PROGRESS  COMPLETED
taskmd    ~/workplace/gg/taskmd/taskmd-1     40     18       4            18
myapp     ~/workplace/myapp                  12     8        2            2
dotfiles  ~/dotfiles                         3      1        0            2
```

Flags:
- `--format json|yaml|table` — output format (default: table)

### `taskmd project register`

Register the current directory (or a given path) as a project:

```bash
# Register cwd
taskmd project register

# Register with explicit id
taskmd project register --id myapp

# Register a specific path
taskmd project register --path ~/workplace/myapp --id myapp
```

Behavior:
- Derives `id` from directory basename if not provided
- Errors if the directory has no `.taskmd.yaml`
- Errors if `id` is already taken in the registry
- Appends to `~/.taskmd.yaml` `projects` list

### `taskmd project unregister`

Remove a project from the registry:

```bash
taskmd project unregister            # Unregister cwd
taskmd project unregister --id myapp # Unregister by id
```

Does not delete any files — only removes the entry from `~/.taskmd.yaml`.

### `--project` flag

Commands that scan tasks gain a `--project` flag:

```bash
taskmd list --project taskmd
taskmd next --project taskmd
taskmd graph --project taskmd
taskmd metrics --project taskmd
```

When `--project` is specified and you're not already inside that project's directory tree, taskmd resolves the project path from the global registry and scans from there. When you're already inside a project directory, `--project` is a no-op (you're already scoped).

### `--all-projects`

Aggregate across all registered projects:

```bash
taskmd list --all-projects
taskmd next --all-projects
```

Task output includes a `project` column so you can tell which project each task belongs to. Task IDs are qualified as `<project-id>:<task-id>` when displayed in aggregate mode to avoid ambiguity.

### `default_project`

Optional convenience for users who primarily work on one project:

```yaml
# ~/.taskmd.yaml
default_project: taskmd

projects:
  - id: taskmd
    path: ~/workplace/gg/taskmd/taskmd-1
```

When set and you're not inside any project directory, commands like `taskmd list` and `taskmd next` automatically scope to the default project instead of failing with "no .taskmd.yaml found."

## Data model changes

No changes to the Task struct or frontmatter schema. Projects are a config-level concept, not a task-level one.

### New config struct

```go
// GlobalProjectEntry represents a registered project in ~/.taskmd.yaml
type GlobalProjectEntry struct {
    ID   string `yaml:"id"`
    Name string `yaml:"name"`
    Path string `yaml:"path"`
}
```

### Home config loading

Today, `~/.taskmd.yaml` is loaded as a fallback by viper's config search. For the global registry, we need to read `~/.taskmd.yaml` independently of the local config walk-up — specifically, just the `projects` list. This avoids conflicts where the home config's `dir` or `phases` bleed into a local project.

```go
func LoadGlobalRegistry() ([]GlobalProjectEntry, error) {
    // Always read ~/.taskmd.yaml (or $TASKMD_HOME_CONFIG)
    // Parse only the "projects" key
    // Resolve ~ in paths
}
```

## Validation

### Registry validation

On `taskmd projects` or any `--project`/`--all-projects` use:

1. Each entry must have `path`
2. `path` must exist and be a directory
3. `path` should contain a `.taskmd.yaml` (warn if missing)
4. `id` values must be unique across the registry

### No task-level validation changes

Since there's no `project` frontmatter field, the existing task validation is unaffected.

## Migration

No migration. Users opt in by running `taskmd project register` from their project directories, or by manually adding entries to `~/.taskmd.yaml`.

## Implementation order

1. **Home config parsing** — Read `projects` list from `~/.taskmd.yaml` independently of viper walk-up
2. **`taskmd projects`** — List registered projects with task count stats
3. **`taskmd project register/unregister`** — Manage the registry (read/write `~/.taskmd.yaml`)
4. **`--project` flag** — Resolve project path from registry, load its config, scan its tasks
5. **`--all-projects`** — Iterate over all registered projects, aggregate results with qualified IDs
6. **`default_project`** — Fallback when not inside any project directory

## Open questions

1. **Home config file location.** Currently `~/.taskmd.yaml` serves double duty as both a fallback project config and the global registry. Should the registry live in a dedicated file like `~/.config/taskmd/projects.yaml`? One file is simpler; separate files are cleaner. Leaning toward keeping `~/.taskmd.yaml` for now and splitting later if needed.

2. **Qualified ID syntax.** `taskmd:042` is used for display in `--all-projects` mode. Should this syntax also work as input to commands like `taskmd set taskmd:042 --status completed`? Useful but adds parsing complexity. Could defer to v2.

3. **Stale registry entries.** What happens when a registered path no longer exists (directory deleted, moved)? Current plan: warn on `taskmd projects`, skip gracefully on `--all-projects`. Could add `taskmd project check` to audit the registry.
