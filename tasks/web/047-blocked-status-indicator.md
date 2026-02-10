---
id: "047"
title: "Add blocked status indicator to tasks table"
status: completed
priority: medium
effort: small
dependencies:
  - "010"
tags:
  - ui
  - tasks
  - blocked
  - ux
created: 2026-02-08
---

# Add Blocked Status Indicator to Tasks Table

## Objective

Enhance the tasks table to clearly display whether each task is blocked or unblocked, and show the count of blocking tickets for blocked tasks.

## Context

Currently, users cannot easily see if a task is blocked when viewing the tasks table. This information is critical for understanding which tasks are actionable vs. waiting on dependencies.

The task data already includes a `blockedBy` array that lists the IDs of blocking tasks. This task leverages that data to provide visual feedback in the UI.

## Tasks

### Add Blocked Status Column

- [ ] Add a new column to the tasks table for blocked status
  - Display blocked/unblocked indicator
  - Position near the status column for visibility
  - Keep column sortable for easy filtering

- [ ] Visual indicator for unblocked tasks
  - Show checkmark icon or "Ready" badge
  - Use positive color (green/success variant)
  - Clear, concise label

- [ ] Visual indicator for blocked tasks
  - Show warning/block icon or "Blocked" badge
  - Use warning color (amber/warning variant)
  - Include count of blocking tickets: "Blocked (3)"

### Data Integration

- [ ] Calculate blocked status from `task.blockedBy`
  - Blocked when `blockedBy.length > 0`
  - Unblocked when `blockedBy.length === 0`

- [ ] Display blocking count
  - Show number: `blockedBy.length`
  - Format: "Blocked (N)" where N is the count
  - For single blocker: "Blocked (1)"
  - For multiple: "Blocked (3)", etc.

### Interactive Features

- [ ] Add tooltip/hover state
  - Show list of blocking task IDs on hover
  - Format: "Blocked by: #001, #003, #005"
  - Optional: Show task titles if available

- [ ] Make blocking task IDs clickable (optional)
  - Click to navigate to blocking task detail
  - Requires integration with task routing (task 018)

### Sorting and Filtering

- [ ] Enable sorting by blocked status
  - Group blocked tasks together when sorted
  - Sort by count (most blocked first) as secondary sort

- [ ] Add filter option (optional)
  - Quick filter to show only blocked tasks
  - Quick filter to show only unblocked/ready tasks
  - Integrate with existing filter bar (task 024)

### Visual Polish

- [ ] Responsive design
  - Column adapts to mobile/tablet layouts
  - Icon-only view on small screens
  - Full text + count on larger screens

- [ ] Accessibility
  - Proper ARIA labels for screen readers
  - Color is not the only indicator (use icons + text)
  - Keyboard navigation support

## Design Considerations

### Desktop Layout
```
Status      | Blocked Status  | Title
----------- | --------------- | -----
In Progress | ✓ Ready         | Task A
Pending     | ⚠ Blocked (2)   | Task B
```

### Mobile Layout (Icon Only)
```
Status      | 🔒 | Title
----------- | -- | -----
In Progress | ✓  | Task A
Pending     | ⚠  | Task B
```

### Tooltip Example
Hovering over "Blocked (2)" shows:
```
Blocked by:
  #003 - Setup database schema
  #007 - API authentication
```

## Acceptance Criteria

- Users can immediately see if a task is blocked in the table view
- Blocked tasks show the exact count of blocking tickets
- Unblocked/ready tasks have clear positive indicator
- Column is sortable to group blocked tasks together
- Visual design uses color, icons, and text for clarity
- Responsive design works on mobile and desktop
- Accessible to screen readers and keyboard navigation
- Tooltip shows blocking task details on hover
- No performance degradation with large task lists

## References

- Task 010 (existing task table implementation)
- Task 024 (enhanced filtering - potential integration point)
- Task 018 (URL routing - for clickable task links)
- TanStack Table API for custom columns
- shadcn/ui Badge component for status indicators
