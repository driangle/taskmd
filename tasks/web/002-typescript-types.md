---
id: "002"
title: "TypeScript types and shared constants"
status: pending
priority: high
effort: small
dependencies:
  - "001"
tags:
  - types
  - core
created: 2026-02-08
---

# TypeScript Types and Shared Constants

## Objective

Define all TypeScript types, interfaces, and shared constants used across the application. This forms the contract between the API layer, parsing logic, and UI components.

## Tasks

- [ ] Create `src/lib/types.ts` with the following types:
  - `TaskStatus` — union type: `"pending"` | `"in_progress"` | `"completed"` | `"cancelled"`
  - `TaskPriority` — union type: `"low"` | `"medium"` | `"high"` | `"critical"`
  - `TaskFrontmatter` — the YAML frontmatter fields (id, title, status, priority, effort, dependencies, tags, created, updated)
  - `Task` — full task object including frontmatter fields, `filePath`, and `body` (markdown content)
  - `Project` — `{ id: string; name: string; path: string }`
  - `AppConfig` — `{ projects: Project[]; activeProjectId: string | null }`
- [ ] Define constants arrays: `TASK_STATUSES`, `TASK_PRIORITIES`, `EFFORT_LEVELS`
- [ ] Create `src/lib/utils.ts` with the `cn()` utility (clsx + tailwind-merge) — standard shadcn pattern
- [ ] Install `clsx` and `tailwind-merge`

## Acceptance Criteria

- All types are exported and importable from `@/lib/types`
- Constants arrays match their respective union types
- `cn()` utility works for conditional class merging
- No `any` types — everything is properly typed
