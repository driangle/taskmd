---
id: "090"
title: "Add drag-and-drop to board view"
status: completed
priority: medium
effort: medium
tags:
  - web
  - mvp
created: 2026-02-14
---

# Add Drag-and-Drop to Board View

## Objective

Enable drag-and-drop functionality on the Board web interface so users can move task cards between status columns (e.g., from "pending" to "in-progress"). Each task card should have a visible drag handle to initiate the drag. Dropping a card into a different column updates the task's status via the API.

## Tasks

- [x] Choose and integrate a drag-and-drop library (e.g., dnd-kit or react-beautiful-dnd)
- [x] Add a drag handle element to each task card in the board view
- [x] Implement drag-and-drop between board columns
- [x] Update the task status via the API when a card is dropped into a new column
- [x] Add visual feedback during drag (placeholder, card shadow, column highlight)
- [x] Handle edge cases (dropping back to same column, empty columns, API errors)
- [x] Ensure drag-and-drop works on touch devices
- [ ] Write tests for drag-and-drop interactions and status update logic

## Acceptance Criteria

- Task cards have a visible drag handle
- Cards can be dragged from one status column to another
- Dropping a card into a different column updates the task's status and persists the change
- Visual feedback is provided during the drag operation
- Dropping a card back into its original column is a no-op
- Works on both desktop and touch devices
