---
id: "014"
title: "Polish, error handling, and loading states"
status: completed
priority: medium
effort: medium
dependencies:
  - "009"
  - "011"
  - "012"
  - "013"
tags:
  - ui
  - polish
created: 2026-02-08
---

# Polish, Error Handling, and Loading States

## Objective

Final pass to ensure the app handles all edge cases gracefully, has consistent loading states, and provides a polished user experience.

## Tasks

- [ ] Loading states:
  - Skeleton loading for the task table while fetching
  - Skeleton loading for the project switcher while fetching config
  - Loading spinner on buttons during API calls
  - Full-page loading state on initial load
- [ ] Empty states:
  - No projects configured → show onboarding prompt to add a project folder
  - No tasks in project → show "No tasks yet" with a CTA to create one
  - No tasks matching filters → show "No tasks match your filters" with clear filters button
- [ ] Error handling:
  - API error responses display user-friendly messages
  - Network errors show a retry option
  - Toast notifications for all mutations (success and failure)
  - Global error boundary for unexpected crashes
- [ ] Confirmation dialogs:
  - Delete task → "Are you sure?" confirmation
- [ ] Visual polish:
  - Consistent spacing and alignment across all components
  - Hover states on interactive elements
  - Focus rings for keyboard navigation
  - Responsive layout adjustments for smaller screens
- [ ] Accessibility:
  - All interactive elements are keyboard accessible
  - Screen reader labels on icon-only buttons
  - Focus management when dialogs open/close

## Acceptance Criteria

- No unhandled loading or error states visible to the user
- All destructive actions require confirmation
- Error messages are user-friendly (not raw error strings)
- App is usable with keyboard only
- First-time user experience is clear (guided to add a project)
- No layout shifts during loading
