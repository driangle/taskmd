---
id: "web-022"
title: "Hyperlink tasks to detail page across all views"
status: pending
priority: high
effort: small
dependencies: ["web-017", "web-018"]
tags:
  - ui
  - tasks
  - ux
created: 2026-02-08
---

# Hyperlink Tasks to Detail Page Across All Views

## Objective

Ensure that every time a task appears in the UI (task list, board, graph), its name/title is a clickable link that navigates to the task detail page (`/tasks/:id`). This gives users a consistent way to drill into task details from any view.

## Tasks

### Tasks Table (`TaskTable.tsx`)

- [ ] Make the task title column a `<Link to={/tasks/${task.id}}>` instead of plain text
- [ ] Style as a subtle link (underline on hover, or blue text)

### Board View (`BoardView.tsx`)

- [ ] Make the task title in each card a `<Link to={/tasks/${task.id}}>`
- [ ] Ensure the card itself remains non-clickable (only the title links)

### Graph View (`GraphView.tsx`)

- [ ] Add click handlers to Mermaid graph nodes that navigate to `/tasks/:id`
  - Mermaid's `securityLevel: "loose"` is already set, which allows click callbacks
  - Use Mermaid's `click` directive in the graph syntax to bind node clicks
  - Or: add post-render event listeners on SVG nodes
- [ ] Ensure the cursor changes to pointer on hoverable nodes

## Acceptance Criteria

- Task titles in the task list table are clickable links to `/tasks/:id`
- Task titles in board cards are clickable links to `/tasks/:id`
- Task nodes in the dependency graph are clickable and navigate to `/tasks/:id`
- All links use client-side routing (no full page reload)
- Visual affordance (cursor, underline, or color) indicates clickability
