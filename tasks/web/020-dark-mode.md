---
id: "web-020"
title: "Dark mode support"
status: pending
priority: low
effort: medium
dependencies: ["web-015"]
tags:
  - ui
  - ux
  - polish
created: 2026-02-08
---

# Dark Mode Support

## Objective

Add dark mode to the web dashboard, respecting system preference by default with a manual toggle.

## Tasks

- [ ] Configure Tailwind v4 dark mode (class-based strategy)
- [ ] Add dark mode color tokens for all UI elements:
  - Background, text, borders
  - Status/priority badge colors
  - Table row hover/stripe colors
  - Board column backgrounds
  - Graph/Mermaid theme
- [ ] Create a theme toggle button in the Shell header
- [ ] Persist preference in localStorage
- [ ] Default to system preference (`prefers-color-scheme: dark`)
- [ ] Ensure Mermaid diagrams render correctly in dark mode

## Acceptance Criteria

- Dashboard respects system dark mode preference on first visit
- Toggle switches between light and dark modes
- Preference persists across page refreshes
- All UI elements are readable in both modes
- Mermaid graphs adapt to the active theme
