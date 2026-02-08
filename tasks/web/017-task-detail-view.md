---
id: "web-017"
title: "Task detail view - Click to expand task body"
status: pending
priority: medium
effort: medium
dependencies: ["web-015"]
tags:
  - ui
  - tasks
  - ux
created: 2026-02-08
---

# Task Detail View - Click to Expand Task Body

## Objective

Allow users to click a task row in the table to view its full markdown body. Currently the API returns tasks without the body field (`json:"-"`). This requires both a backend change to expose task body and a frontend panel to display it.

## Tasks

### Backend

- [ ] Add `GET /api/tasks/:id` endpoint that returns a single task including its markdown body
- [ ] Or: add a query parameter `GET /api/tasks?include=body` to include body in list response
- [ ] Ensure markdown body is properly escaped in JSON

### Frontend

- [ ] Create `src/components/tasks/TaskDetail.tsx` — slide-out panel or expandable row
  - Display task title, status, priority, effort, tags as metadata header
  - Render markdown body (simple rendering or raw markdown display)
  - Show file path and dependencies
  - Close button / click-outside to dismiss
- [ ] Make task rows clickable in `TaskTable.tsx`
- [ ] Add a hook or fetch for individual task data if using a detail endpoint

## Acceptance Criteria

- Clicking a task row opens a detail panel showing the full task
- Task metadata (status, priority, tags, dependencies) is displayed clearly
- Markdown body content is visible
- Panel can be closed to return to the table view
