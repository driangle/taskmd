---
id: "010"
title: "Task table — column definitions + base"
status: completed
priority: high
effort: large
dependencies:
  - "007"
  - "008"
tags:
  - ui
  - tasks
  - table
created: 2026-02-08
---

# Task Table — Column Definitions + Base

## Objective

Build the core task data table using TanStack Table and shadcn/ui's table component. This is the main view of the application — it displays all tasks from the active project folder.

## Tasks

- [ ] Create `src/hooks/use-tasks.ts`
  - SWR hook that fetches `GET /api/tasks?projectId=xxx`
  - Re-fetches when the active project changes
  - Expose: `tasks`, `isLoading`, `error`, `mutate`
  - Expose mutation helpers: `createTask()`, `updateTask()`, `deleteTask()`
- [ ] Create `src/components/tasks/columns.tsx`
  - Define TanStack Table column definitions:
    - **ID** — short, monospace, e.g., "001"
    - **Title** — primary text, takes most space
    - **Status** — badge with color coding
    - **Priority** — badge with color coding
    - **Effort** — text label
    - **Tags** — comma-separated or badges
    - **Created** — formatted date
    - **Actions** — dropdown menu (edit, delete)
- [ ] Create `src/components/tasks/status-badge.tsx`
  - Colored badge component for task status
  - Colors: pending=gray, in_progress=blue, completed=green, cancelled=red
- [ ] Create `src/components/tasks/priority-badge.tsx`
  - Colored badge component for task priority
  - Colors: low=gray, medium=yellow, high=orange, critical=red
- [ ] Create `src/components/tasks/task-table.tsx`
  - Initialize TanStack Table with column definitions and task data
  - Render using shadcn `<Table>` components
  - Handle empty state (no tasks) with a helpful message
  - Handle loading state with skeleton rows
- [ ] Integrate the task table into `src/app/page.tsx`

## Acceptance Criteria

- Table renders all tasks from the active project
- Columns display correctly formatted data
- Status and priority badges are color-coded
- Empty state shows a message and CTA to create a task
- Loading state shows skeleton placeholders
- Table is responsive and handles long titles gracefully
