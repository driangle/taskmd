---
id: "013"
title: "Task creation dialog"
status: completed
priority: medium
effort: medium
dependencies:
  - "007"
  - "010"
tags:
  - ui
  - tasks
created: 2026-02-08
---

# Task Creation Dialog

## Objective

Build a dialog/modal for creating new tasks. The dialog collects the required fields, calls the API to create the task file, and refreshes the table.

## Tasks

- [ ] Create `src/components/tasks/create-task-dialog.tsx`
  - Modal form with fields:
    - **Title** (required) — text input
    - **Status** — select dropdown, defaults to "pending"
    - **Priority** — select dropdown, defaults to "medium"
    - **Effort** — select dropdown, optional
    - **Tags** — comma-separated text input, optional
    - **Description** (body) — textarea for markdown content, optional
  - ID is auto-generated (not user-editable)
- [ ] Add form validation
  - Title is required and non-empty
  - Show inline validation errors
- [ ] On submit:
  - Call `POST /api/tasks` with the form data
  - Show loading state on the submit button
  - On success: close dialog, refresh task list, show success toast
  - On error: show error message, keep dialog open
- [ ] Add a "Create Task" button to the page header / above the table
  - Button opens the dialog
- [ ] Support keyboard shortcuts:
  - Enter to submit (when not in textarea)
  - Escape to close

## Acceptance Criteria

- Dialog opens from a prominent "Create Task" button
- All fields are present and have appropriate defaults
- Title is validated as required
- Submitting creates a new `.md` file in the project folder
- The new task appears in the table immediately after creation
- The auto-generated ID follows the existing numbering sequence
