---
id: "web-021"
title: "Responsive layout for mobile and tablet"
status: completed
priority: low
effort: medium
dependencies: ["web-015"]
tags:
  - ui
  - ux
  - polish
  - post-mvp
created: 2026-02-08
---

# Responsive Layout for Mobile and Tablet

## Objective

Make the web dashboard usable on smaller screens. Currently the layout assumes desktop width — tables overflow, board columns are cramped, and the graph may not scale.

## Tasks

- [x] Make Shell header responsive (collapsible nav or hamburger menu on mobile)
- [x] Task table: horizontal scroll on narrow screens, hide secondary columns on mobile
- [x] Board view: single-column stack on mobile, horizontal scroll on tablet
- [x] Graph view: responsive height and min-height on small screens
- [x] Stats view: already responsive with grid-cols-2/sm:grid-cols-4
- [x] Test at common breakpoints: 375px (phone), 768px (tablet), 1024px+ (desktop)

## Acceptance Criteria

- Dashboard is usable on phone-sized screens (no content clipped)
- Navigation works on all screen sizes
- Tables don't break layout on narrow screens
- Board columns are readable on tablet
