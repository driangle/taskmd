---
id: "102"
title: "Document new CLI commands in the CLI guide (sync, report, tracks, archive, next-id, get, set, tags)"
status: pending
priority: medium
effort: medium
tags:
  - documentation
  - cli
  - mvp
created: 2026-02-14
---

# Document New CLI Commands in the CLI Guide

## Objective

The CLI guide (`docs/guides/cli-guide.md`) is missing dedicated reference sections for 8 commands that have been added since v0.0.4. Add a section for each command under the existing `## Command Reference` heading, following the same style and depth as the existing entries (list, validate, next, graph, stats, board, snapshot, web).

## Background

The guide currently documents: `list`, `validate`, `next`, `graph`, `stats`, `board`, `snapshot`, `web`. The following commands exist in the CLI but have no reference section:

| Command | Source | Notes |
|---------|--------|-------|
| `sync` | `internal/cli/sync.go` | New sync infrastructure with GitHub Issues source. Also needs `.taskmd-sync.yaml` config documented. |
| `report` | `internal/cli/report.go` | Generates markdown, HTML, and JSON reports with grouping and critical-path analysis. |
| `tracks` | `internal/cli/tracks.go` | Assigns parallel work tracks based on `touches`/scope overlap. |
| `archive` | `internal/cli/archive.go` | Archives or deletes completed/cancelled tasks. |
| `next-id` | `internal/cli/nextid.go` | Returns the next available task ID (useful for scripting). |
| `get` | `internal/cli/get.go` | Show details of a single task by ID or filepath. |
| `set` | `internal/cli/set.go` | Update a task's frontmatter fields (status, priority, etc). |
| `tags` | `internal/cli/tags.go` | List all tags with usage counts. |

## Tasks

- [ ] Add `### get - View Task Details` section with usage, flags, and examples
- [ ] Add `### set - Update Task Fields` section with usage, flags, and examples
- [ ] Add `### tags - List Tags` section with usage and examples
- [ ] Add `### archive - Archive Completed Tasks` section with usage, flags, and examples
- [ ] Add `### next-id - Get Next Available ID` section with usage and examples
- [ ] Add `### report - Generate Reports` section covering md/html/json formats, grouping options, and critical-path analysis
- [ ] Add `### tracks - Parallel Work Tracks` section explaining scope overlap detection and the `touches` field
- [ ] Add `### sync - Sync External Sources` section covering GitHub Issues source, `.taskmd-sync.yaml` config, field mapping, and conflict strategies
- [ ] Add sync configuration to the `## Configuration` section (document `.taskmd-sync.yaml`)

## Acceptance Criteria

- Every command listed above has a dedicated subsection under Command Reference
- Each section includes: description, usage syntax, flags table, and 2–3 practical examples
- The sync section documents `.taskmd-sync.yaml` structure and field mapping
- Style and depth matches existing command sections in the guide
