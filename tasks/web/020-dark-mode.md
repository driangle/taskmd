---
id: "web-020"
title: "Dark mode support"
status: completed
priority: low
effort: medium
dependencies: ["web-015"]
tags:
  - ui
  - ux
  - polish
  - mvp
created: 2026-02-08
---

# Dark Mode Support

## Objective

Add dark mode to the web dashboard, respecting system preference by default with a manual toggle.

## Tasks

- [x] Configure Tailwind v4 dark mode (class-based strategy)
- [x] Add dark mode color tokens for all UI elements:
  - Background, text, borders
  - Status/priority badge colors
  - Table row hover/stripe colors
  - Board column backgrounds
  - Graph/React Flow theme
- [x] Create a theme toggle button in the Shell header
- [x] Persist preference in localStorage
- [x] Default to system preference (`prefers-color-scheme: dark`)
- [x] Ensure React Flow graphs render correctly in dark mode

## Acceptance Criteria

- Dashboard respects system dark mode preference on first visit
- Toggle switches between light and dark modes
- Preference persists across page refreshes
- All UI elements are readable in both modes
- React Flow graphs adapt to the active theme
