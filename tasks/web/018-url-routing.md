---
id: "web-018"
title: "URL-based routing with deep linking"
status: pending
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

- [ ] Add `react-router-dom` dependency
- [ ] Set up routes: `/tasks`, `/board`, `/graph`, `/stats`, `/validate`
- [ ] Add task detail route: `/tasks/:id` for individual task pages
- [ ] Update `Shell.tsx` navigation to use `<NavLink>` instead of onClick state
- [ ] Default route `/` redirects to `/tasks`
- [ ] Preserve query parameters for board groupBy (e.g., `/board?groupBy=priority`)
- [ ] Ensure SPA fallback in Go server works with all routes (already using `/{path...}`)
- [ ] Update Vite proxy config if needed

## Acceptance Criteria

- Each view has its own URL path
- Task detail page accessible at `/tasks/:id`
- Browser back/forward buttons work correctly
- Refreshing the page stays on the current view
- Bookmarks work for any view
- Board groupBy selection is preserved in the URL
