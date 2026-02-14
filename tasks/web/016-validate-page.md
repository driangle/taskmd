---
id: "web-016"
title: "Validation page - Display task file issues"
status: completed
priority: low
effort: small
dependencies: ["web-015"]
tags:
  - ui
  - validation
  - quality
  - mvp
created: 2026-02-08
---

# Validation Page - Display Task File Issues

## Objective

Add a Validate tab to the web dashboard that displays issues found by the validator (missing fields, broken dependencies, duplicate IDs, etc.). The backend `/api/validate` endpoint already exists.

## Tasks

- [ ] Create `src/hooks/use-validate.ts` — SWR hook fetching `GET /api/validate`
- [ ] Create `src/pages/ValidatePage.tsx` — render validation issues
  - Show issue count summary (errors vs warnings)
  - List issues grouped by file path
  - Each issue shows: severity, message, and affected task ID
  - Color-code by severity (error=red, warning=yellow)
- [ ] Add "Validate" tab to `App.tsx` navigation
- [ ] Handle empty state (no issues = "All tasks are valid" message)
- [ ] Wire up live reload so validation refreshes on file changes

## Acceptance Criteria

- Validate tab shows all issues from `/api/validate`
- Issues are grouped by file and color-coded by severity
- Clean state shows a success message
- Auto-refreshes when task files change via SSE
