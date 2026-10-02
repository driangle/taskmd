---
id: "01m3yn6f8"
title: "Resolve efforts and next.unphased per project in the web project switcher"
status: pending
priority: medium
effort: small
type: bug
phase: web-ui
dependencies: ["01m3ygtgq"]
tags: ["web", "config", "projects"]
created: "2026-10-02"
---

# Resolve efforts and next.unphased per project in the web project switcher

## Objective

When the web server switches project (`?project=<id>`), it resolves the scan
dir and **phases** from that project's own `.taskmd.yaml`
(`buildResolveProject` in `apps/cli/internal/cli/web.go`,
`ProjectResolverFunc` in `apps/cli/internal/web/project_resolver.go`), but the
**effort vocabulary** and **`next.unphased`** still come from the project the
server was started in (`web.Config.Efforts` / `web.Config.UnphasedPlacement`).
So a switched-to project with a custom effort scale gets its edits validated
(`PUT /api/tasks/{id}`), its board/tracks/validate views computed, and its
`/api/next` ranked against another project's config.

Surfaced while doing 01m3ygtgq, which added `next.unphased`; the efforts gap
predates it.

Out of scope: the MCP server's per-call `task_dir` uses startup config for
everything, and documents that as a deliberate limitation in
`apps/cli/internal/mcp/server.go`.

## Tasks

- [ ] Have the project resolver return a per-project config (phases, efforts,
      unphased placement) instead of phases alone, read standalone from the
      project's `.taskmd.yaml` like `loadProjectPhases`
- [ ] Use the per-project values in every handler that today reads
      `s.config.Efforts` / `s.config.UnphasedPlacement` when a project is selected
      (`/api/tasks/{id}` PUT, `/api/board`, `/api/tracks`, `/api/validate`,
      `/api/next`, and the config/efforts payload the UI reads)
- [ ] Decide how an invalid `next.unphased` in a switched-to project surfaces
      (request error, not a silent fallback)
- [ ] Tests with two projects whose efforts and `next.unphased` differ

## Acceptance Criteria

- With the server started in project A, selecting project B makes `/api/next`
  honour B's `next.unphased` and the effort-dependent endpoints use B's effort scale
- Requests without a project parameter behave exactly as today
- An invalid `next.unphased` in project B produces a clear error on B's requests
