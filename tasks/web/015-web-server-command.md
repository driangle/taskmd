---
id: "web-015"
title: "web start command - Serve web dashboard from CLI"
status: completed
priority: high
effort: large
dependencies: []
tags:
  - cli
  - web
  - go
  - commands
  - infrastructure
created: 2026-02-08
---

# Web Start Command - Serve Web Dashboard from CLI

## Objective

Implement `taskmd web start` to launch a Go HTTP server that serves a TypeScript web dashboard. The Go server provides JSON API endpoints that reuse existing CLI packages (scanner, graph, metrics, validator). The frontend is a thin Vite + React SPA that presents data without business logic.

## Tasks

### Go: File Watcher

- [x] Create `internal/watcher/watcher.go` with fsnotify-based file watching
- [x] Watch scan directory recursively for `.md` file changes
- [x] Debounce rapid changes (200ms default)
- [x] Provide `onChange` callback for consumers
- [x] Handle new subdirectory creation (re-watch)
- [x] Promote `fsnotify` from indirect to direct dependency in `go.mod`
- [x] Add tests with temporary filesystem changes

### Go: Web Server Package

- [x] Create `internal/web/server.go` — HTTP server setup with `http.ServeMux`
- [x] Create `internal/web/data_provider.go` — cached scan results with dirty-flag invalidation
- [x] Create `internal/web/handlers.go` — API endpoint handlers calling existing packages
- [x] Create `internal/web/sse.go` — SSE endpoint broadcasting file change events
- [x] CORS middleware for dev mode (in server.go)
- [x] Graceful shutdown via `context.Context` + signal handling
- [x] Add handler and server tests

### Go: CLI Command

- [x] Create `internal/cli/web.go` — Cobra `web` parent command + `start` subcommand
- [x] Flags: `--port` (int, default 8080), `--dir` (string, default "."), `--dev` (bool), `--open` (bool)
- [x] `runWebStart` wires up server with watcher and starts listening

### Go: Refactoring for Reuse

- [x] Extract board grouping logic to `internal/board/` package to avoid import cycles
- [x] Export `GroupTasks()` and `GroupResult` for handler reuse

### Go: Static Embedding

- [x] Create `internal/web/embed.go` with `//go:embed static/dist` (build tag: embed_web)
- [x] Create `internal/web/embed_stub.go` with empty FS for CLI-only builds (build tag: !embed_web)
- [x] Fallback HTML page when no assets embedded

### Frontend: Replace Next.js with Vite

- [x] Replace `apps/web/package.json` — swap `next` for `vite` + `@vitejs/plugin-react`
- [x] Create `apps/web/vite.config.ts` with `/api` proxy to Go server for dev
- [x] Create `apps/web/index.html` entry point
- [x] Remove Next.js-specific files
- [x] Keep existing deps: React 19, Tailwind v4, TanStack Table, SWR
- [x] Add `mermaid` for graph rendering

### Frontend: TypeScript Types and API Layer

- [x] Create `src/api/types.ts` — TS interfaces mirroring Go JSON output
- [x] Create `src/api/client.ts` — fetch wrapper with base URL
- [x] Create SWR hooks: `use-tasks.ts`, `use-board.ts`, `use-graph.ts`, `use-stats.ts`
- [x] Create `src/hooks/use-live-reload.ts` — EventSource → SWR `mutate()` on "reload" event

### Frontend: Views

- [x] Create `src/App.tsx` — tab-based navigation shell
- [x] Create `src/components/layout/Shell.tsx` — app layout with header tabs
- [x] Create `src/pages/TasksPage.tsx` — TanStack Table with sorting/filtering
- [x] Create `src/pages/BoardPage.tsx` — kanban columns with groupBy selector
- [x] Create `src/pages/GraphPage.tsx` — Mermaid.js rendering of dependency graph
- [x] Create `src/pages/StatsPage.tsx` — metric cards and breakdowns

### Build Pipeline

- [x] Add `web-build` Makefile target
- [x] Add `web-embed` Makefile target
- [x] Add `build-full` Makefile target
- [x] Add `install-full` Makefile target
