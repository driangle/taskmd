---
id: "011"
title: "Table sorting and filtering"
status: pending
priority: medium
effort: medium
dependencies:
  - "010"
tags:
  - ui
  - tasks
  - table
created: 2026-02-08
---

# Table Sorting and Filtering

## Objective

Add client-side sorting and filtering to the task table. Users should be able to sort by any column and filter by status, priority, and free-text search.

## Tasks

- [ ] Create `src/components/tasks/task-filters.tsx`
  - Filter bar above the table with:
    - **Text search** — filters across title, ID, tags, and body
    - **Status filter** — multi-select dropdown to show/hide statuses
    - **Priority filter** — multi-select dropdown to show/hide priorities
  - Show active filter count / "clear filters" button
- [ ] Enable TanStack Table sorting
  - Click column headers to sort (ascending → descending → none)
  - Visual indicator (arrow) on sorted column
  - Default sort: by ID ascending
  - Support multi-column sort (shift+click)
- [ ] Enable TanStack Table filtering
  - Wire filter controls to TanStack Table's column filter state
  - Text search uses a global filter function
  - Status/priority use column-level faceted filters
- [ ] Show result count (e.g., "Showing 5 of 12 tasks")
- [ ] Persist filter state in URL search params (optional, nice-to-have)

## Acceptance Criteria

- Clicking column headers sorts the table
- Sort direction is visually indicated
- Text search filters tasks in real-time as the user types
- Status and priority filters work as multi-select (show only selected values)
- Filters can be cleared individually or all at once
- Filter + sort combinations work together correctly
