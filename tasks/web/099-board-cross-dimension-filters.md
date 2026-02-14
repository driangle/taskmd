---
id: "099"
title: "Add cross-dimension filter pills to board view"
status: pending
priority: medium
effort: medium
tags:
  - web
  - mvp
created: 2026-02-14
---

# Add Cross-Dimension Filter Pills to Board View

## Objective

When the board is grouped by one dimension (e.g., status), add filter pills for other dimensions (e.g., priority, effort, tags) so users can narrow down which cards are visible. For example, grouping by status and filtering by "critical" priority would only show critical tasks across the status columns.

## Tasks

- [ ] Determine which dimensions to offer as filters based on the current groupBy selection
- [ ] Add filter pill UI above the board (reuse or adapt the existing FilterBar pattern from TaskTable)
- [ ] Wire filter state into the board API query or apply client-side filtering to the board data
- [ ] Ensure selected filters persist across groupBy changes where applicable
- [ ] Test that filters work correctly with drag-and-drop (task 090)
- [ ] Verify both light and dark mode styling

## Acceptance Criteria

- When grouped by status, filter pills for priority, effort, and tags are shown
- When grouped by priority, filter pills for status, effort, and tags are shown
- Same pattern for other groupBy values
- Clicking a pill toggles it on/off and immediately filters the visible cards
- Empty columns (after filtering) still display with a "No tasks" placeholder
- Filters do not interfere with drag-and-drop functionality
- Works in both light and dark modes
