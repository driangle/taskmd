# Markdown Task Tracker — MVP Plan

## Overview

A local web-based task management interface where tasks are stored as markdown files with YAML frontmatter. Users point it at folders on their filesystem, and it reads/writes `.md` files directly — no database.

## Tech Stack

- Next.js 14+ (App Router, TypeScript)
- Tailwind CSS + shadcn/ui
- gray-matter (YAML frontmatter parsing)
- TanStack Table (data table)
- SWR (data fetching)

## Architecture

```
[Browser UI]  <-->  [Next.js API Routes]  <-->  [Local Filesystem (.md files)]
```

- **API layer:** Next.js Route Handlers for CRUD on tasks and project config
- **Config:** `~/.md-task-tracker/config.json` stores project folder paths
- **Parsing:** gray-matter for round-trip frontmatter read/write preserving markdown body
- **Table:** TanStack Table with client-side sort/filter (all tasks loaded at once)
- **Editing:** Optimistic inline updates → PATCH API → write back to `.md` file

## Directory Structure

```
src/
├── app/
│   ├── layout.tsx, page.tsx, globals.css
│   └── api/
│       ├── projects/route.ts
│       └── tasks/route.ts, [id]/route.ts
├── lib/
│   ├── types.ts, markdown.ts, filesystem.ts, config.ts, utils.ts
├── hooks/
│   ├── use-tasks.ts, use-projects.ts
└── components/
    ├── ui/ (shadcn)
    ├── layout/ (app-sidebar, header)
    ├── projects/ (project-switcher, add-project-dialog)
    └── tasks/ (task-table, columns, task-filters, status-badge, priority-badge, create-task-dialog)
```

## Task Breakdown

| ID  | Title                                      | Deps            | Effort |
|-----|--------------------------------------------|-----------------|--------|
| 001 | Project scaffolding (Next.js + Tailwind)   | —               | small  |
| 002 | TypeScript types and shared constants      | 001             | small  |
| 003 | Markdown parsing and serialization service | 002             | medium |
| 004 | Filesystem service (dir scanning, CRUD)    | 003             | medium |
| 005 | Config service (project path persistence)  | 002             | small  |
| 006 | API routes — Projects                      | 005             | small  |
| 007 | API routes — Tasks (CRUD)                  | 004, 006        | large  |
| 008 | Install shadcn/ui components + layout shell| 001             | medium |
| 009 | Project switcher and add-project dialog    | 006, 008        | medium |
| 010 | Task table — column definitions + base     | 007, 008        | large  |
| 011 | Table sorting and filtering                | 010             | medium |
| 012 | Inline status and priority editing         | 010             | medium |
| 013 | Task creation dialog                       | 007, 010        | medium |
| 014 | Polish, error handling, and loading states | 009,011,012,013 | medium |

## Dependency Graph

```
001 (Scaffolding)
├── 002 (Types)
│   ├── 003 (Markdown) → 004 (Filesystem) → 007 (Task API) ──┐
│   └── 005 (Config) → 006 (Project API) ──────────┐          │
└── 008 (shadcn + Layout)                           │          │
    ├── 009 (Project Switcher) ◄────────────────────┘          │
    └── 010 (Task Table) ◄─────────────────────────────────────┘
        ├── 011 (Sorting + Filtering)
        ├── 012 (Inline Editing)
        └── 013 (Task Creation)
            └── 014 (Polish)
```
