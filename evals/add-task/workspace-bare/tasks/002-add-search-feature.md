---
id: "002"
title: Add full-text search
status: pending
priority: medium
effort: large
type: feature
tags: [search, backend, frontend]
created: 2026-03-02
---

# Add full-text search

## Objective

Implement full-text search across all user content. Users should be able to search by keyword and get ranked results.

## Tasks

- [ ] Design the search index schema
- [ ] Implement backend search API endpoint
- [ ] Add search UI with autocomplete
- [ ] Index existing content
- [ ] Add result ranking and highlighting
- [ ] Write API tests
- [ ] Add frontend component tests
- [ ] Performance test with 10k+ documents

## Acceptance Criteria

- Search returns relevant results within 200ms
- Results are ranked by relevance
- Search supports partial matches and typo tolerance
- UI shows highlighted matching terms
