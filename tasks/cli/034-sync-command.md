---
id: "034"
title: "Sync command with pluggable sources"
status: cancelled
priority: low
effort: large
superseded_by:
  - "082"
  - "083"
  - "084"
  - "085"
  - "086"
dependencies: []
tags:
  - cli
  - go
  - integration
  - mvp
created: 2026-02-08
---

# Sync Command with Pluggable Sources

## Objective

Add a `taskmd sync` command that synchronizes tasks from external project management tools (Jira, Asana, Trello, Monday.com, GitHub Issues, etc.) into local markdown task files. The architecture should be extensible so new sources can be added as plugins without modifying core sync logic.

## Tasks

- [ ] Design a `Source` interface in `internal/sync/` that all providers implement (e.g. `FetchTasks()`, `Name()`, `ValidateConfig()`)
- [ ] Implement a provider registry that discovers and registers available sources
- [ ] Add a sync config format (e.g. `.taskmd-sync.yaml`) to define which source to use, credentials, project/board IDs, and field mappings
- [ ] Implement the `taskmd sync` CLI command that reads config, fetches from the source, and writes/updates markdown files
- [ ] Handle conflict resolution: detect when a local file was modified and the remote changed too
- [ ] Map external fields (status, priority, assignee, labels) to taskmd frontmatter fields
- [ ] Support `--dry-run` flag to preview what would be created/updated/deleted
- [ ] Support `--source` flag to sync a specific source when multiple are configured
- [ ] Implement at least one provider (e.g. Jira or Trello) as a reference implementation
- [ ] Write tests for the core sync engine and the source interface contract

## Provider Interface (rough sketch)

```go
type Source interface {
    Name() string
    ValidateConfig(cfg map[string]any) error
    FetchTasks(cfg map[string]any) ([]ExternalTask, error)
}
```

## Acceptance Criteria

- `taskmd sync` fetches tasks from a configured source and writes them as `.md` files
- Adding a new source only requires implementing the `Source` interface and registering it
- `--dry-run` shows a preview without writing files
- Field mapping is configurable per source
- Existing local tasks are updated in place, not duplicated
- Tests cover the sync engine and at least one provider
