---
id: "084"
title: "Sync source: Jira"
status: pending
priority: low
effort: small
dependencies:
  - "082"
tags:
  - cli
  - go
  - integration
created: 2026-02-14
---

# Sync Source: Jira

## Objective

Implement a Jira sync source for `taskmd sync`. This provider fetches issues from a Jira project and maps them to local taskmd markdown files.

## Tasks

- [ ] Implement the `Source` interface for Jira in `internal/sync/jira/`
- [ ] Authenticate via Jira API token and base URL (from config or environment variables)
- [ ] Fetch issues from a configured project using the Jira REST API
- [ ] Support JQL filters in config for fine-grained issue selection
- [ ] Map Jira fields to taskmd frontmatter (priority, status, labels, assignee, sprint, etc.)
- [ ] Handle Jira's rich text description by converting to markdown
- [ ] Write tests with mocked Jira API responses

## Config Example

```yaml
sources:
  - name: jira
    type: jira
    base_url: https://myorg.atlassian.net
    project: PROJ
    token_env: JIRA_TOKEN
    email_env: JIRA_EMAIL
    jql: "status != Done AND sprint in openSprints()"
    field_map:
      priority: priority
      status: status
      labels: tags
      assignee: assignee
```

## Acceptance Criteria

- `taskmd sync --source jira` fetches issues and creates/updates markdown files
- Jira fields are mapped correctly to taskmd frontmatter
- JQL filtering works as configured
- Authentication works via environment variables or config
- Tests cover the provider with mocked API responses
