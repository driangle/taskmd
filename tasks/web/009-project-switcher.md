---
id: "009"
title: "Project switcher and add-project dialog"
status: pending
priority: medium
effort: medium
dependencies:
  - "006"
  - "008"
tags:
  - ui
  - projects
created: 2026-02-08
---

# Project Switcher and Add-Project Dialog

## Objective

Build the UI components for managing projects: a dropdown to switch between configured project folders and a dialog to add new project folders.

## Tasks

- [ ] Create `src/hooks/use-projects.ts`
  - SWR hook that fetches `GET /api/projects`
  - Expose: `projects`, `activeProject`, `isLoading`, `error`
  - Expose mutation helpers: `addProject()`, `removeProject()`, `setActiveProject()`
  - Use SWR's `mutate` for optimistic updates
- [ ] Create `src/components/projects/project-switcher.tsx`
  - Dropdown/select showing all configured projects
  - Display the active project name prominently
  - Selecting a different project calls `setActiveProject()` → `PATCH /api/projects/[id]`
  - "Add project" option at the bottom of the dropdown
  - Show folder path as secondary text for each project
- [ ] Create `src/components/projects/add-project-dialog.tsx`
  - Dialog/modal form with fields: project name, folder path
  - Folder path is a text input (user types/pastes the absolute path)
  - Validates on submit via `POST /api/projects`
  - Shows error if path is invalid or doesn't exist
  - Closes and refreshes project list on success
- [ ] Integrate project switcher into the sidebar

## Acceptance Criteria

- Users can add a new project folder by providing a name and path
- Users can switch between projects from the sidebar
- The active project is visually highlighted
- Adding a project with an invalid path shows an error message
- Switching projects triggers a task list refresh
