---
id: "01kz3c24p"
title: "Reduce web test coupling to CSS and markup structure"
status: completed
priority: low
effort: medium
type: improvement
phase: critical-feedback
dependencies: []
tags: [web, testing, tech-debt]
created_at: 2026-08-03
completed_at: 2026-08-04
---

# Reduce web test coupling to CSS and markup structure

## Objective

The web app's logic/util tests are good, but ~105 assertions across 29 test files couple tests
to Tailwind class names and DOM structure via `querySelector` — e.g.
`src/components/shared/LoadingState.test.tsx` asserts on `.py-12`, `.animate-pulse`,
`.grid-cols-2`, and exact child counts (6 rows, 4 columns) of a purely presentational skeleton.
These tests break on any restyle and validate structure rather than behavior — the clearest
low-value test cluster in the repo. Rework them toward behavior/role-based queries.

## Tasks

- [x] Inventory the ~105 CSS-class / `querySelector` assertions across the 29 files
- [x] For presentational components, replace class/structure assertions with role/text/a11y queries
      (Testing Library `getByRole`, `getByText`, etc.)
- [x] Delete assertions that only check styling with no behavioral meaning
- [x] Keep (or add) tests that assert user-visible behavior and state transitions
- [x] Confirm web coverage stays reasonable after removing brittle assertions

## Acceptance Criteria

- Presentational-only class/structure assertions are removed or replaced with behavior queries
- Remaining web tests do not break on a pure Tailwind restyle
- Meaningful behavior coverage is preserved (no net loss of real assertions)
