---
id: "086"
title: "Sync source: Trello"
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
  - sync/trello
  - sync/core
created: 2026-02-14
---

# Sync Source: Trello

## Objective

Implement a Trello sync source for `taskmd sync`. This provider fetches cards from a Trello board and maps them to local taskmd markdown files.

## Tasks

- [ ] Implement the `Source` interface for Trello in `internal/sync/trello/`
- [ ] Authenticate via Trello API key and token (from config or environment variables)
- [ ] Fetch cards from a configured board using the Trello REST API
- [ ] Map Trello fields to taskmd frontmatter (list name to status, labels to tags, members to assignee, due date, etc.)
- [ ] Support filtering by list, label, or member in config
- [ ] Write tests with mocked Trello API responses

## Config Example

```yaml
sources:
  - name: trello
    type: trello
    api_key_env: TRELLO_API_KEY
    token_env: TRELLO_TOKEN
    board_id: abc123
    filters:
      lists: ["To Do", "In Progress"]
    field_map:
      list: status
      labels: tags
      members: assignee
      due: due_date
```

## Acceptance Criteria

- `taskmd sync --source trello` fetches cards and creates/updates markdown files
- Trello fields are mapped correctly to taskmd frontmatter
- Filtering by list/label works as configured
- Authentication works via environment variables or config
- Tests cover the provider with mocked API responses
