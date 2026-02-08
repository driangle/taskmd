---
id: "web-001"
title: "Project scaffolding (Vite + React + Tailwind)"
status: completed
priority: high
effort: small
dependencies: []
tags:
  - setup
  - infrastructure
created: 2026-02-08
---

# Project Scaffolding (Next.js + Tailwind)

## Objective

Initialize the Next.js project with TypeScript, Tailwind CSS, and the App Router. Set up the foundational project structure that all other tasks build upon.

## Tasks

- [ ] Run `npx create-next-app@latest` with TypeScript, Tailwind CSS, ESLint, App Router, and `src/` directory enabled
- [ ] Verify the dev server starts cleanly (`npm run dev`)
- [ ] Create the directory structure under `src/`:
  - `src/lib/`
  - `src/hooks/`
  - `src/components/ui/`
  - `src/components/layout/`
  - `src/components/projects/`
  - `src/components/tasks/`
- [ ] Install core dependencies: `gray-matter`, `swr`, `@tanstack/react-table`
- [ ] Clean up default Next.js boilerplate from `page.tsx` and `globals.css`
- [ ] Add a minimal `page.tsx` that renders a placeholder heading

## Acceptance Criteria

- `npm run dev` starts without errors
- Project uses App Router (files in `src/app/`)
- Tailwind CSS is configured and working
- All core dependencies are installed
- Directory structure is in place
