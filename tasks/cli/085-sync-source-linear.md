---
id: "085"
title: "Sync source: Linear"
status: pending
priority: low
effort: small
dependencies:
  - "082"
tags:
  - cli
  - go
  - integration
touches:
  - sync/linear
  - sync/core
created: 2026-02-14
---

# Sync Source: Linear

## Objective

Implement a Linear sync source for `taskmd sync`. This provider fetches issues from a Linear team/project and maps them to local taskmd markdown files.

## Tasks

- [ ] Implement the `Source` interface for Linear in `internal/sync/linear/`
- [ ] Authenticate via Linear API key (from config or environment variable)
- [ ] Fetch issues from a configured team or project using the Linear GraphQL API
- [ ] Map Linear fields to taskmd frontmatter (priority, state, labels, assignee, cycle, etc.)
- [ ] Support filtering by team, project, label, or cycle in config
- [ ] Write tests with mocked Linear API responses

## Config Example

```yaml
sources:
  - name: linear
    type: linear
    token_env: LINEAR_API_KEY
    team: ENG
    filters:
      project: "Q1 Roadmap"
      state: ["In Progress", "Todo"]
    field_map:
      priority: priority
      state: status
      labels: tags
      assignee: assignee
```

## Acceptance Criteria

- `taskmd sync --source linear` fetches issues and creates/updates markdown files
- Linear fields are mapped correctly to taskmd frontmatter
- Filtering by team/project/state works as configured
- Authentication works via environment variable or config
- Tests cover the provider with mocked API responses
