---
id: "web-021"
title: "Responsive layout for mobile and tablet"
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

# Responsive Layout for Mobile and Tablet

## Objective

Make the web dashboard usable on smaller screens. Currently the layout assumes desktop width — tables overflow, board columns are cramped, and the graph may not scale.

## Tasks

- [ ] Make Shell header responsive (collapsible nav or hamburger menu on mobile)
- [ ] Task table: horizontal scroll on narrow screens, or switch to card layout on mobile
- [ ] Board view: single-column stack on mobile, horizontal scroll on tablet
- [ ] Graph view: ensure Mermaid diagram is scrollable/zoomable on small screens
- [ ] Stats view: stack metric cards vertically on mobile
- [ ] Test at common breakpoints: 375px (phone), 768px (tablet), 1024px+ (desktop)

## Acceptance Criteria

- Dashboard is usable on phone-sized screens (no content clipped)
- Navigation works on all screen sizes
- Tables don't break layout on narrow screens
- Board columns are readable on tablet
