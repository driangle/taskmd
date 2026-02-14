---
id: "web-019"
title: "Polish loading and error states"
status: completed
priority: low
effort: small
dependencies: ["web-015"]
tags:
  - ui
  - polish
  - ux
  - web
  - mvp
created: 2026-02-08
---

# Polish Loading and Error States

## Objective

Replace bare "Loading..." text and "Error: ..." messages with proper UI components. Add skeleton loaders, retry buttons, and empty state illustrations.

## Tasks

- [x] Create a reusable `LoadingSpinner` component (or skeleton loader)
- [x] Create a reusable `ErrorMessage` component with retry button
- [x] Update all pages to use these components:
  - TasksPage: skeleton table rows while loading
  - BoardPage: skeleton columns while loading
  - GraphPage: placeholder while Mermaid renders
  - StatsPage: skeleton cards while loading
- [x] Add empty state handling:
  - Tasks table: "No tasks found" message when empty
  - Board: "No tasks" message (already partial)
  - Graph: "No dependencies to display" when graph is empty
  - Stats: "No data" when no tasks exist
- [x] Handle API connection errors gracefully (server not running)

## Acceptance Criteria

- All pages show visual loading indicators instead of text
- Error states include a retry button that re-fetches data
- Empty states show helpful messages
- Connection errors show a clear "cannot reach server" message
