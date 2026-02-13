---
id: "070"
title: "Clickable tags navigate to filtered list view"
status: completed
priority: high
effort: small
dependencies:
  - "024"
tags:
  - web
  - typescript
  - ux
  - navigation
  - mvp
created: 2026-02-13
---

# Clickable Tags Navigate to Filtered List View

## Objective

Make tags clickable throughout the web app so that clicking a tag navigates the user to the task list page with that tag pre-selected as a filter. This applies to:

1. **Task detail page** — tags displayed in the metadata section
2. **Task list table** — tags displayed in each row (already clickable for local filtering, but should also work when navigating from other pages)

## Context

Task 024 implemented interactive tag filtering in the list view — clicking a tag in a table row toggles it as a filter. However, tags on the **task detail page** are plain `<span>` elements with no interactivity. Users should be able to click a tag anywhere and land on the filtered list.

The list view already supports tag filtering via `selectedTags` state in `TaskTable.tsx`. The missing piece is:
- Accepting an initial tag filter via URL search params (e.g. `/tasks?tag=cli`)
- Making detail page tags into links/buttons that navigate to `/tasks?tag=<value>`
- Ensuring list-view tag clicks in table rows also update the URL for shareability

## Tasks

### URL-based tag filter state

- [ ] Update `TasksPage.tsx` to read `tag` search param from the URL on mount
- [ ] Pass the initial tag value down to `TaskTable.tsx` to seed `selectedTags` state
- [ ] When `selectedTags` changes, sync back to URL search params (replace, not push)
- [ ] Support both single and multiple tags via repeated params (e.g. `?tag=cli&tag=web`)

### Task detail page — clickable tags

- [ ] In `TaskDetailPage.tsx`, replace tag `<span>` elements with `<Link>` (or `<button>` + `useNavigate`)
- [ ] Each tag navigates to `/tasks?tag=<tagValue>`
- [ ] Style tags with the same interactive look as the list view (hover state, pointer cursor)

### Task list table — sync tag clicks with URL

- [ ] When a tag is toggled in the table via click, update the URL search params to reflect the current tag filter
- [ ] On page load, initialize `selectedTags` from URL params if present

## Acceptance Criteria

- Clicking a tag on the task detail page navigates to `/tasks?tag=<tag>`
- The list view loads with that tag pre-selected in the filter bar
- Clicking a tag in a list table row still toggles the filter AND updates the URL
- Navigating to `/tasks?tag=cli` directly applies the filter on load
- Multiple tags supported: `/tasks?tag=cli&tag=web` filters by both
- Tag styling on the detail page matches the interactive style used in the table
- Browser back/forward works correctly with tag filter URL state

## Files to Modify

- `apps/web/src/pages/TaskDetailPage.tsx` — make tags clickable links
- `apps/web/src/pages/TasksPage.tsx` — read URL search params, pass to TaskTable
- `apps/web/src/components/tasks/TaskTable.tsx` — accept initial tags, sync state to URL
- `apps/web/src/components/tasks/TaskTable/columns.tsx` — no changes needed (already clickable)

## References

- Task 024 (enhanced tag filtering — already completed)
- Task 017 (task detail view)
- Task 018 (URL routing)
