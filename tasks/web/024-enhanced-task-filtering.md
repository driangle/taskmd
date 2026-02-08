---
id: "024"
title: "Enhanced task filtering with pills and interactive tags"
status: pending
priority: medium
effort: medium
dependencies:
  - "011"
tags:
  - ui
  - tasks
  - filter
  - ux
created: 2026-02-08
---

# Enhanced Task Filtering with Pills and Interactive Tags

## Objective

Enhance the task table filtering UX by replacing dropdowns with visual pill-based filters for status and priority, and enable filtering by tags through direct interaction with tag elements in the table rows.

## Context

Task 011 implemented basic filtering with dropdowns. This task improves the UX by making filters more visual and interactive:
- **Status pills** — Visual chips that can be toggled on/off to show/hide statuses
- **Priority pills** — Visual chips for priority filtering
- **Clickable tags** — Click any tag in the table to filter by that tag
- More intuitive, visual filtering without dropdown menus

## Tasks

### Update Filter Bar Component

- [ ] Replace status dropdown with pill-based multi-select
  - Display all possible statuses as toggleable pills/chips
  - Active filters shown with colored background (matching status colors)
  - Inactive/unselected shown as outlined/ghost variant
  - Click to toggle on/off
  - Default: all statuses selected (show all)

- [ ] Replace priority dropdown with pill-based multi-select
  - Display all priorities (low, medium, high, critical) as pills
  - Use priority color scheme (matching existing badge colors)
  - Click to toggle on/off
  - Default: all priorities selected (show all)

- [ ] Add active tag filter display
  - Show currently active tag filters as removable pills
  - Click X icon to remove tag filter
  - Visual indicator when tag filters are active

### Add Tag Click Interaction

- [ ] Make tags in table rows clickable
  - Add hover state to tag badges
  - Cursor changes to pointer on hover
  - Click handler adds tag to active filters
  - If tag already filtered, clicking again removes it (toggle)

- [ ] Update tag badge styling
  - Visual indicator that tags are interactive
  - Hover effects (slightly darker/brighter)
  - Active state when tag is in active filters (e.g., border highlight)

### Filter State Management

- [ ] Update filter state to handle pill selections
  - Track selected statuses, priorities, and tags in state
  - When all items in a category are selected, treat as "show all"
  - When none selected, show none (empty state message)

- [ ] Update URL params for filter persistence (optional)
  - Encode status, priority, and tag filters in URL
  - Support deep-linking to filtered views
  - Restore filter state from URL on page load

### Visual Polish

- [ ] Add clear filters button
  - Only shown when filters are active (not default "all")
  - Resets to show all statuses, priorities, and clears tags
  - Visual indicator of how many filters are active

- [ ] Add filter count badge
  - Show number of active filters (when not "show all")
  - Position near filter bar or in tab/page header

- [ ] Empty state when no tasks match filters
  - Friendly message: "No tasks match your filters"
  - Quick action to clear filters

### Animation and Transitions

- [ ] Add smooth transitions when filters change
  - Fade in/out for task rows
  - Pill color transitions on toggle
  - Count updates with subtle animation

## Design Considerations

### Status Pills Layout
```
Status: [Pending] [In Progress] [Blocked] [Completed]
         ^active   ^active       ^inactive  ^inactive
```

### Priority Pills Layout
```
Priority: [Low] [Medium] [High] [Critical]
```

### Tag Filter Display
```
Tags: [cli ×] [bugfix ×] [ui ×]
```

### Tag Badge in Table Row
When hovering over a tag in a task row, show subtle highlight and pointer cursor to indicate it's clickable.

## Acceptance Criteria

- Status filters displayed as visual pills, not dropdowns
- Priority filters displayed as visual pills, not dropdowns
- Clicking status/priority pills toggles filter on/off
- Visual distinction between active and inactive pills
- Clicking tags in table rows adds them to active filters
- Clicking an already-filtered tag removes it from filters
- Tag badges show hover state indicating they're interactive
- Active tag filters displayed as removable pills in filter bar
- Clear filters button resets to "show all" state
- Filter state persists in URL (optional nice-to-have)
- Empty state shown when no tasks match filters
- Smooth transitions when applying/removing filters
- Result count updates correctly: "Showing X of Y tasks"

## References

- Task 011 (existing filtering implementation)
- shadcn/ui Badge component for pills
- TanStack Table filtering API
