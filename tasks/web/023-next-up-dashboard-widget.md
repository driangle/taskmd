---
id: "web-023"
title: "Next Up widget - Surface recommended tasks in the web UI"
status: pending
priority: low
effort: medium
dependencies: ["032"]
tags:
  - ui
  - productivity
  - api
  - post-mvp
created: 2026-02-08
---

# Next Up Widget - Surface Recommended Tasks in the Web UI

## Objective

Bring the `taskmd next` scoring and ranking logic into the web dashboard so users can see at a glance which tasks deserve attention, without leaving the browser or running a CLI command.

## Problem

The web UI currently shows tasks organized by status, priority, or dependencies — but it never answers "what should I work on next?" Users must switch to the CLI and run `taskmd next` to get recommendations. The web dashboard should surface this insight directly.

## Design

Add a **"Next Up"** panel as a new top-level tab alongside Tasks, Board, Graph, and Stats. The panel shows a ranked list of recommended tasks with scores, reasons, and visual cues that make the recommendations scannable.

### Key elements

- **Ranked card list** — Each recommendation is a card showing rank, title, priority badge, score bar, and reason chips (e.g., "critical path", "unblocks 3 tasks", "quick win").
- **Limit control** — A small selector (3 / 5 / 10) to adjust how many recommendations are shown.
- **Filter bar** — Optional tag or priority filter chips to scope recommendations (mirrors CLI `--filter`).
- **Score breakdown on hover/expand** — Hovering a card (or clicking to expand) reveals the scoring breakdown: base priority points, critical path bonus, downstream bonus, effort bonus.
- **Link to task detail** — Each card title links to the task detail page when `web-017` is available, otherwise displays inline.

## Tasks

### Backend

- [ ] Add `GET /api/next` endpoint in `internal/web/handlers.go`
  - Accept query params: `limit` (int, default 5), `filter` (repeated, e.g. `?filter=tag%3Dcli&filter=priority%3Dhigh`)
  - Reuse the scoring logic from `internal/cli/next.go` — extract `hasUnmetDependencies`, `isActionable`, `scoreTask`, and the `Recommendation` struct into a shared package (e.g. `internal/next/`) so both CLI and web can import it
  - Return `[]Recommendation` as JSON
- [ ] Add handler tests in `internal/web/handlers_test.go`
- [ ] Register route in `server.go`

### Frontend

- [ ] Add `use-next.ts` hook — SWR fetch from `/api/next?limit=N&filter=...`
- [ ] Add `Recommendation` type to `api/types.ts`
- [ ] Create `pages/NextPage.tsx` — the Next Up tab page
- [ ] Create `components/next/NextView.tsx` — ranked card list
- [ ] Create `components/next/RecommendationCard.tsx` — individual recommendation card with:
  - Rank number (large, bold)
  - Task title and ID
  - Priority badge (reuse existing color scheme)
  - Reason chips (small colored pills)
  - Score indicator (small horizontal bar or number)
  - Downstream count ("unblocks N tasks" with a subtle icon)
  - Critical path indicator (icon or badge when `on_critical_path` is true)
- [ ] Add limit selector (3 / 5 / 10) that re-fetches with updated `limit` param
- [ ] Add optional filter chips that append `filter` query params
- [ ] Register "Next Up" tab in `Shell.tsx` / `App.tsx` routing
- [ ] Integrate with live reload (SSE) so recommendations update when tasks change on disk

## Acceptance Criteria

- `GET /api/next` returns scored recommendations matching the CLI `taskmd next` output
- The "Next Up" tab appears in the web navigation
- Recommendations show rank, title, priority, reasons, and score
- Changing the limit re-fetches and updates the list
- Filters narrow the recommendation set
- Blocked and completed tasks never appear
- Recommendations update live when task files change on disk
- Scoring logic is shared between CLI and web (no duplication)

## Example API Response

```json
[
  {
    "rank": 1,
    "id": "003",
    "title": "Build CLI parser",
    "status": "pending",
    "priority": "critical",
    "effort": "small",
    "score": 60,
    "reasons": ["critical priority", "on critical path", "unblocks 1 task", "quick win"],
    "downstream_count": 1,
    "on_critical_path": true
  }
]
```
