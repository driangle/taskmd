---
id: "web-017"
title: "Task detail page"
status: pending
priority: high
effort: medium
dependencies: ["web-015", "web-018"]
tags:
  - ui
  - tasks
  - ux
created: 2026-02-08
---

# Task Detail Page

## Objective

Create a dedicated task detail page at `/tasks/:id` that shows the full details of a task including its rendered markdown body. Currently the API returns tasks without the body field (`json:"-"`). This requires both a backend change to expose task body and a frontend page to display it.

## Tasks

### Backend

- [ ] Add `GET /api/tasks/:id` endpoint that returns a single task including its markdown body
- [ ] Ensure markdown body is properly escaped in JSON

### Frontend

- [ ] Create `src/pages/TaskDetailPage.tsx` — a dedicated page routed at `/tasks/:id`
  - Display task title, status, priority, effort, tags as metadata header
  - Render markdown body content (use a markdown renderer or `dangerouslySetInnerHTML` with sanitized HTML)
  - Show file path, dependencies, and other frontmatter fields
  - Back navigation to return to the previous view
- [ ] Add data fetching hook for the individual task endpoint

## Acceptance Criteria

- Navigating to `/tasks/:id` shows the full task detail page
- Task metadata (status, priority, tags, dependencies) is displayed clearly
- Markdown body content is rendered
- Back button returns to the previous view (tasks list, board, or graph)
