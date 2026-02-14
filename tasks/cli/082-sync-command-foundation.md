---
id: "082"
title: "Sync command foundation with pluggable source architecture"
status: completed
priority: low
effort: medium
dependencies: []
tags:
  - cli
  - go
  - integration
  - mvp
created: 2026-02-14
---

# Sync Command Foundation

## Objective

Set up the core `taskmd sync` command with a pluggable source architecture. This lays the groundwork so that individual providers (GitHub Issues, Jira, etc.) can be added independently by implementing a simple interface.

## Tasks

- [ ] Design a `Source` interface in `internal/sync/` (e.g. `FetchTasks()`, `Name()`, `ValidateConfig()`)
- [ ] Implement a provider registry that discovers and registers available sources
- [ ] Define the sync config format (`.taskmd-sync.yaml`) for specifying sources, credentials, project/board IDs, and field mappings
- [ ] Implement the `taskmd sync` CLI command that reads config, fetches from the configured source, and writes/updates markdown files
- [ ] Handle conflict resolution: detect when a local file was modified and the remote changed too
- [ ] Map external fields (status, priority, assignee, labels) to taskmd frontmatter fields via configurable field mappings
- [ ] Support `--dry-run` flag to preview what would be created/updated/deleted
- [ ] Support `--source` flag to sync a specific source when multiple are configured
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

- `taskmd sync` command exists and reads `.taskmd-sync.yaml` config
- Adding a new source only requires implementing the `Source` interface and registering it
- `--dry-run` shows a preview without writing files
- Field mapping is configurable per source
- Existing local tasks are updated in place, not duplicated
- Tests cover the sync engine and interface contract
