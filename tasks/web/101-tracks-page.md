---
id: "101"
title: "Tracks page for parallel work visualization"
status: in-progress
priority: medium
effort: large
tags:
  - mvp
created: 2026-02-14
---

# Tracks Page for Parallel Work Visualization

## Objective

Add a "Tracks" page to the web UI that visualizes parallel work tracks, mirroring the functionality of the CLI `tracks` command. The page should show tasks grouped into parallel lanes based on scope overlap (via the `touches` field), helping teams see at a glance which tasks can be worked on simultaneously without merge conflicts.

## Tasks

- [ ] Add `/api/tracks` endpoint that returns track assignment data (reuse CLI tracks algorithm via Go backend)
- [ ] Create `TracksPage` component with route at `/tracks`
- [ ] Render track lanes as vertical columns or horizontal swim lanes, each labeled with its scopes
- [ ] Show tasks within each lane ordered by score (priority, critical path, downstream impact)
- [ ] Display a "Flexible" section for tasks with no `touches` field
- [ ] Add task cards showing ID, title, priority, effort, and touched scopes
- [ ] Support filtering (by tag, group, priority) consistent with other pages
- [ ] Add navigation link to Tracks page in the sidebar/header
- [ ] Ensure responsive layout for varying numbers of tracks
- [ ] Add empty state when no actionable tasks exist

## Acceptance Criteria

- Tracks page is accessible from the main navigation
- Tasks sharing a `touches` scope appear in separate tracks (never in the same lane)
- Tasks without `touches` appear in a clearly labeled flexible section
- Each track lane displays its aggregated scopes
- Filtering works consistently with the rest of the web UI
- Page handles edge cases: no tasks, no tracks (all flexible), single track
- Track data comes from the same algorithm used by the CLI `tracks` command
