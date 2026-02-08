---
id: "007"
title: "API routes — Tasks (CRUD)"
status: pending
priority: high
effort: large
dependencies:
  - "004"
  - "006"
tags:
  - api
  - tasks
created: 2026-02-08
---

# API Routes — Tasks (CRUD)

## Objective

Create the Next.js Route Handlers for full CRUD operations on tasks. These endpoints read/write `.md` files in the active project's folder via the filesystem service.

## Tasks

- [ ] Create `src/app/api/tasks/route.ts`
  - `GET /api/tasks` — list all tasks in the active project folder
    - Query param: `?projectId=xxx` (optional, defaults to active project)
    - Uses `scanDirectory()` to read all `.md` files
    - Return `{ tasks: Task[] }`
  - `POST /api/tasks` — create a new task
    - Body: `{ title: string; status?: string; priority?: string; body?: string; ... }`
    - Auto-generates ID and filename
    - Uses `writeTaskFile()` to persist
    - Return the created `Task` with 201
- [ ] Create `src/app/api/tasks/[id]/route.ts`
  - `GET /api/tasks/[id]` — get a single task by ID
    - Find the task in the scanned directory matching the ID
  - `PATCH /api/tasks/[id]` — update a task
    - Body: partial `TaskFrontmatter` fields and/or `body`
    - Read the existing file, merge changes, write back
    - This is the endpoint used for inline editing (status, priority changes)
  - `DELETE /api/tasks/[id]` — delete a task file
    - Remove the `.md` file from disk
- [ ] Resolve the project path: look up `projectId` in config, or use active project
- [ ] Handle error cases:
  - No active project configured → 400
  - Project path doesn't exist → 404
  - Task ID not found → 404
  - Invalid field values → 400

## API Summary

| Method | Endpoint          | Description            |
|--------|-------------------|------------------------|
| GET    | /api/tasks        | List all tasks         |
| POST   | /api/tasks        | Create a new task      |
| GET    | /api/tasks/[id]   | Get a single task      |
| PATCH  | /api/tasks/[id]   | Update task fields     |
| DELETE | /api/tasks/[id]   | Delete a task file     |

## Acceptance Criteria

- Full CRUD lifecycle works: create → read → update → delete
- Inline field updates (PATCH with `{ status: "completed" }`) work correctly
- File content is preserved when updating only frontmatter fields
- Proper error responses for all failure cases
- Tasks are always returned with their `filePath` for reference
