---
id: "012"
title: "Inline status and priority editing"
status: completed
priority: high
effort: medium
dependencies:
  - "010"
tags:
  - ui
  - tasks
  - editing
created: 2026-02-08
---

# Inline Status and Priority Editing

## Objective

Allow users to change a task's status and priority directly from the table without opening a separate edit view. Changes are immediately written back to the `.md` file via the API.

## Tasks

- [ ] Make status badges clickable in the table
  - Click opens a dropdown with all status options
  - Selecting a new status triggers an optimistic update
  - Calls `PATCH /api/tasks/[id]` with `{ status: newStatus }`
  - On error, reverts the optimistic update and shows a toast
- [ ] Make priority badges clickable in the table
  - Same pattern as status: click → dropdown → optimistic update → API call
  - Calls `PATCH /api/tasks/[id]` with `{ priority: newPriority }`
- [ ] Implement optimistic updates in `use-tasks.ts`
  - Update the local SWR cache immediately
  - Revalidate from the server after the mutation completes
  - Rollback on error
- [ ] Add visual feedback:
  - Brief highlight/flash on the row when a change is saved
  - Loading spinner or subtle indicator while the API call is in flight
  - Toast notification on error

## Acceptance Criteria

- Clicking a status badge opens a dropdown to change the status
- Clicking a priority badge opens a dropdown to change the priority
- Changes appear instantly in the table (optimistic update)
- Changes are persisted to the `.md` file on disk
- Failed updates revert the change and show an error message
- The markdown body is not affected by inline field updates
