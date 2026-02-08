---
id: "web-015"
title: "web start command - Serve web dashboard from CLI"
status: pending
priority: high
effort: large
dependencies: ["027"]
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

Implement `taskmd web start --port 8000 [directory]` to launch a Go HTTP server that serves a TypeScript web dashboard. The Go server provides JSON API endpoints that reuse existing CLI packages (scanner, graph, metrics, validator). The frontend is a thin Vite + React SPA that presents data without business logic.

## Architecture

**Single Go process** serving both the API and the static SPA:

```
taskmd web start --port 8000 ./tasks
         │
         ▼
┌─ Go HTTP Server ──────────────────────────────┐
│  /api/tasks        → scanner + model → JSON    │
│  /api/board?groupBy → scanner + grouping → JSON│
│  /api/graph        → scanner + graph pkg → JSON│
│  /api/graph/mermaid→ scanner + graph → Mermaid │
│  /api/stats        → scanner + metrics → JSON  │
│  /api/validate     → scanner + validator → JSON│
│  /api/events       → SSE (fsnotify)            │
│  /*                → static SPA files           │
│                                                 │
│  DataProvider: caches scan results,             │
│  invalidated by fsnotify watcher                │
└─────────────────────────────────────────────────┘
         │
         ▼
┌─ Browser (Vite + React SPA) ──────────────────┐
│  SWR hooks fetch /api/* endpoints              │
│  SSE listener triggers SWR revalidation        │
│  TanStack Table for task list presentation     │
│  Mermaid.js renders dependency graphs          │
└────────────────────────────────────────────────┘
```

**Key design decisions:**
- Go HTTP server with JSON API (not static file generation) for maximum reuse of existing packages
- Replace Next.js scaffold with Vite + React: produces a plain `dist/` embeddable via `embed.FS`, no Node.js runtime needed
- Cached scan results with fsnotify invalidation (DataProvider pattern)
- SSE for live updates (simpler than WebSocket, unidirectional)
- Frontend does zero business logic; all computation in Go

## Tasks

### Go: File Watcher (dependency: task 027)

- [ ] Create `internal/watcher/watcher.go` with fsnotify-based file watching
- [ ] Watch scan directory recursively for `.md` file changes
- [ ] Debounce rapid changes (200ms default)
- [ ] Provide `onChange` callback for consumers
- [ ] Handle new subdirectory creation (re-watch)
- [ ] Promote `fsnotify` from indirect to direct dependency in `go.mod`
- [ ] Add tests with temporary filesystem changes

### Go: Web Server Package

- [ ] Create `internal/web/server.go` — HTTP server setup with `http.ServeMux`
- [ ] Create `internal/web/data_provider.go` — cached scan results with dirty-flag invalidation
- [ ] Create `internal/web/handlers.go` — API endpoint handlers calling existing packages:
  - `GET /api/tasks` → scanner → JSON task array
  - `GET /api/board?groupBy=status` → scanner + groupTasks → JSON grouped tasks
  - `GET /api/graph` → scanner + `graph.NewGraph().ToJSON()` → JSON nodes/edges
  - `GET /api/graph/mermaid` → scanner + `graph.NewGraph().ToMermaid()` → Mermaid string
  - `GET /api/stats` → scanner + `metrics.Calculate()` → JSON metrics
  - `GET /api/validate` → scanner + `validator.Validate()` → JSON issues
- [ ] Create `internal/web/sse.go` — SSE endpoint broadcasting file change events
- [ ] Create `internal/web/middleware.go` — CORS middleware for dev mode
- [ ] Graceful shutdown via `context.Context` + signal handling
- [ ] Add handler and server tests

### Go: CLI Command

- [ ] Create `internal/cli/web.go` — Cobra `web` parent command + `start` subcommand
- [ ] Flags: `--port` (int, default 8080), `--dev` (bool), `--open` (bool)
- [ ] `runWebStart` wires up server with watcher and starts listening
- [ ] Follow existing command patterns (args[0] for scan directory, default ".")

### Go: Refactoring for Reuse

- [ ] Export `groupTasks()` from `board.go` as `GroupTasks()` for board handler reuse
- [ ] Export `groupResult` as `GroupResult` from `board.go`

### Go: Static Embedding

- [ ] Create `internal/web/embed.go` with `//go:embed static/dist/*`
- [ ] Create `internal/web/embed_noweb.go` with `//go:build noweb` stub for CLI-only builds
- [ ] Create `internal/web/static/dist/.gitkeep` placeholder

### Frontend: Replace Next.js with Vite

- [ ] Replace `apps/web/package.json` — swap `next` for `vite` + `@vitejs/plugin-react`
- [ ] Create `apps/web/vite.config.ts` with `/api` proxy to Go server for dev
- [ ] Create `apps/web/index.html` entry point
- [ ] Remove Next.js-specific files (`next.config.ts`, `next-env.d.ts`, `src/app/`)
- [ ] Keep existing deps: React 19, Tailwind v4, TanStack Table, SWR
- [ ] Drop `gray-matter` (Go handles all parsing)
- [ ] Add `mermaid` for graph rendering

### Frontend: TypeScript Types and API Layer

- [ ] Create `src/api/types.ts` — TS interfaces mirroring Go JSON output (Task, BoardGroup, GraphData, Stats, ValidationResult)
- [ ] Create `src/api/client.ts` — fetch wrapper with base URL
- [ ] Create SWR hooks: `use-tasks.ts`, `use-board.ts`, `use-graph.ts`, `use-stats.ts`
- [ ] Create `src/hooks/use-live-reload.ts` — EventSource → SWR `mutate()` on "reload" event

### Frontend: Views

- [ ] Create `src/App.tsx` — tab-based navigation shell
- [ ] Create `src/components/layout/Shell.tsx` — app layout with sidebar
- [ ] Create `src/pages/TasksPage.tsx` — TanStack Table with sorting/filtering
- [ ] Create `src/pages/BoardPage.tsx` — kanban columns grouped by status/priority/effort
- [ ] Create `src/pages/GraphPage.tsx` — Mermaid.js rendering of dependency graph
- [ ] Create `src/pages/StatsPage.tsx` — metric cards (status/priority/effort distributions)

### Build Pipeline

- [ ] Add `web-build` Makefile target: `cd apps/web && pnpm install && pnpm build`
- [ ] Add `web-embed` Makefile target: copy `apps/web/dist/` → `internal/web/static/dist/`
- [ ] Add `build-full` Makefile target: web-embed + go build

## Acceptance Criteria

- `taskmd web start ./tasks` starts server, serves API and SPA on default port
- `taskmd web start --port 8000 ./tasks` respects port flag
- `curl localhost:8080/api/tasks` returns valid JSON matching CLI `list --format json` output
- `curl localhost:8080/api/stats` returns metrics matching CLI `stats --format json` output
- `curl localhost:8080/api/graph` returns graph JSON matching CLI `graph --format json` output
- Editing a `.md` task file triggers SSE event and frontend auto-refreshes
- `make build-full` produces a single binary that serves everything
- `--dev` mode enables CORS for Vite dev server on separate port
- Frontend renders: task table, kanban board, dependency graph, stats dashboard

## Development Workflow

```bash
# API development
go run ./cmd/taskmd web start --dev --port 8080 ../../tasks/

# Frontend development (hot reload)
# Terminal 1: go run ./cmd/taskmd web start --dev --port 8080 ../../tasks/
# Terminal 2: cd apps/web && pnpm dev  (port 5173, proxies /api → :8080)

# Production build
cd apps/cli && make build-full
./taskmd web start --port 8000 ../../tasks/
```

## Notes

- `fsnotify v1.9.0` is already an indirect dependency via viper
- The `groupTasks` function in `board.go` and `applyFilters`/`sortTasks` in `list.go` are currently unexported
- The graph package already has `ToJSON()` and `ToMermaid()` methods ready for API use
- The metrics package `Calculate()` is already in a separate package, clean to call from handlers
