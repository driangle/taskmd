---
id: "01kz3c23t"
title: "Remove stale PLAN.md and fix project-overview references"
status: completed
priority: high
effort: small
type: docs
phase: critical-feedback
dependencies: []
tags: [docs, cleanup, credibility]
created_at: 2026-08-03
completed_at: 2026-08-03
---

# Remove stale PLAN.md and fix project-overview references

## Objective

`PLAN.md` describes a **Next.js 14 + shadcn/ui + SWR** web app with a `src/app/api/` route
structure and a 14-task breakdown — a stack that was never built. The actual product is a Go
CLI plus a **Vite + React 19** SPA. Worse, `AGENTS.md` cites `PLAN.md` under "Documentation
Locations" as the canonical **"Project overview"**, so the project's own agent guide points
contributors at a document describing an architecture that does not exist. Remove the stale
plan and repoint the overview to something accurate.

## Tasks

- [x] Delete `PLAN.md` (or replace it with an accurate one-paragraph architecture overview)
- [x] Update `AGENTS.md` "Documentation Locations" to stop pointing "Project overview" at `PLAN.md`
- [x] Point the overview at an accurate source (README architecture section or `docs/`)
- [x] Grep the repo for other references to `PLAN.md` and fix or remove them
- [x] Fix the `AGENTS.md` "Go 1.22+" prerequisite to match `go.work` (`go 1.25.0`)

## Acceptance Criteria

- No committed doc describes a Next.js/shadcn stack for this project
- `AGENTS.md` "Project overview" pointer resolves to an accurate, current description
- Go version stated in docs matches `go.work`
- `grep -r "PLAN.md"` returns no stale references
