---
id: "006"
title: "API routes — Projects"
status: pending
priority: medium
effort: small
dependencies:
  - "005"
tags:
  - api
created: 2026-02-08
---

# API Routes — Projects

## Objective

Create the Next.js Route Handlers for managing projects (folder paths). These endpoints allow the UI to list, add, remove, and switch between project folders.

## Tasks

- [ ] Create `src/app/api/projects/route.ts`
  - `GET /api/projects` — return the full config (projects list + active project ID)
  - `POST /api/projects` — add a new project folder
    - Body: `{ name: string; path: string }`
    - Validate that `path` is an absolute path to an existing directory
    - Return the created `Project` object
- [ ] Create `src/app/api/projects/[id]/route.ts`
  - `DELETE /api/projects/[id]` — remove a project from the config
  - `PATCH /api/projects/[id]` — set this project as active
    - Body: `{ active: true }`
- [ ] Consistent error response format: `{ error: string }`
- [ ] Return appropriate HTTP status codes (200, 201, 400, 404)

## API Summary

| Method | Endpoint              | Description              |
|--------|-----------------------|--------------------------|
| GET    | /api/projects         | List all projects + config |
| POST   | /api/projects         | Add a new project folder |
| DELETE | /api/projects/[id]    | Remove a project         |
| PATCH  | /api/projects/[id]    | Set project as active    |

## Acceptance Criteria

- All endpoints return JSON with correct status codes
- Adding a project with an invalid path returns 400
- Removing a non-existent project returns 404
- Setting a project as active persists across requests
