---
id: "web-018"
title: "URL-based routing with deep linking"
status: completed
priority: high
effort: medium
dependencies: ["web-015"]
tags:
  - ui
  - ux
  - infrastructure
created: 2026-02-08
---

# URL-Based Routing with Deep Linking

## Objective

Replace tab-based state switching with URL-based routing so users can bookmark and share links to specific views. Currently the app uses React state for navigation — refreshing the page always returns to the Tasks tab.

## Tasks

- [x] Add `react-router-dom` dependency
- [x] Set up routes: `/tasks`, `/board`, `/graph`, `/stats`
- [x] Add task detail route: `/tasks/:id` for individual task pages
- [x] Update `Shell.tsx` navigation to use `<NavLink>` instead of onClick state
- [x] Default route `/` redirects to `/tasks`
- [x] Preserve query parameters for board groupBy (e.g., `/board?groupBy=priority`)
- [x] Ensure SPA fallback in Go server works with all routes (already using `/{path...}`)
- [x] Update Vite proxy config if needed (no changes needed)

## Acceptance Criteria

- Each view has its own URL path
- Task detail page accessible at `/tasks/:id`
- Browser back/forward buttons work correctly
- Refreshing the page stays on the current view
- Bookmarks work for any view
- Board groupBy selection is preserved in the URL
