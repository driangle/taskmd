---
id: "web-017"
title: "Fix task detail page to show full markdown content"
status: completed
priority: high
effort: medium
dependencies: ["web-015", "web-018"]
tags:
  - ui
  - tasks
  - ux
  - bug
created: 2026-02-08
updated: 2026-02-08
---

# Fix Task Detail Page to Show Full Markdown Content

## Current Status

TaskDetailPage.tsx was created in task web-018 but it's not working properly. The page appears empty because:
1. There's no dedicated `GET /api/tasks/:id` endpoint that returns the task body
2. The current `/api/tasks` endpoint excludes the body field (`json:"-"`)
3. Frontend needs a markdown renderer to display the body content

## Objective

Fix the task detail page at `/tasks/:id` to show the full task content including the rendered markdown body.

## Tasks

### Backend

- [x] Add `GET /api/tasks/:id` endpoint in `internal/web/server.go` that returns a single task
  - Should include the full task body/content from the markdown file
  - Ensure markdown body is properly escaped in JSON
  - Return 404 if task not found
- [x] Consider creating a separate Task type that includes the body field for this endpoint

### Frontend

- [x] Install a markdown renderer library (e.g., `react-markdown` or `marked` + `dompurify`)
- [x] Create `src/api/task-detail.ts` hook to fetch single task: `GET /api/tasks/:id`
- [x] Update `TaskDetailPage.tsx` to:
  - Use the new data fetching hook for individual task
  - Render markdown body content with the markdown renderer
  - Show file path if available
  - Display all frontmatter fields properly
  - Handle loading and error states

## Acceptance Criteria

- Navigating to `/tasks/:id` shows the full task detail page with content
- Task metadata (status, priority, tags, dependencies, created date) is displayed clearly
- Markdown body content is rendered with proper formatting (headers, lists, code blocks, etc.)
- Back button returns to `/tasks` view
- 404 state is shown for invalid task IDs
- Loading states work correctly
