---
title: "Serve merged worktree view in MCP server and web API"
id: "01m0n67bx"
status: completed
priority: medium
type: feature
tags: ["worktrees", "mcp", "web"]
created: "2026-08-22"
dependencies: ["01m0nk00d"]
effort: medium
phase: worktree-support
completed_at: 2026-08-23
---

# Serve merged worktree view in MCP server and web API

## Objective

Serve the merged worktree view through the MCP server and the web server's data
layer/API so agent and browser surfaces see the same effective state the CLI does,
and keep that view live when sibling worktrees change. Server-side only — the React
frontend rendering is a separate task. Spec: `docs/specs/worktree-support.md` §4, §9.

## Tasks

- [x] MCP read tools return effective status/owner plus provenance fields (`worktree`, `branch`, `remote_only`) when the overlay is active; no breaking shape change to existing fields
- [x] MCP mutation tools keep local-only writes and return the sibling-only guard error
- [x] Web `DataProvider` builds the overlay so every endpoint (list, board, graph, stats, detail) serves effective status; task payloads carry the additive provenance fields
- [x] Web mutation endpoints keep local-only writes and return the guard error in the response
- [x] Extend the filesystem watcher to sibling worktrees' tasks dirs (same invalidate + SSE broadcast path and debounce)
- [x] Re-enumerate worktrees on membership change: watch `<common-dir>/worktrees/`, with re-list-on-invalidation as fallback
- [x] Static export bakes effective status and provenance at export time
- [x] Tests: MCP/web handler tests with injected discovery; live-refresh e2e — edit a task in a sibling worktree, assert cache invalidation, SSE broadcast, and updated effective status in the next payload

## Acceptance Criteria

- An MCP client listing tasks from worktree B sees a task claimed in worktree A as `in-progress` with provenance
- A claim made in a sibling worktree reaches connected browsers via SSE within the watcher debounce, without touching the serving worktree's files
- `git worktree add`/`remove` while the server runs updates the overlay without a restart
- Existing MCP clients that ignore the new fields keep working (additive only — `taskmd-mcp` plugin takes a minor bump per ADR 0003)
- Exported static site reflects effective status; single-worktree exports are unchanged
- No behavior change for either surface when the overlay is inactive
