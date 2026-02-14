---
id: "083"
title: "Sync source: GitHub Issues"
status: completed
priority: low
effort: small
dependencies:
  - "082"
tags:
  - cli
  - go
  - integration
  - mvp
created: 2026-02-14
---

# Sync Source: GitHub Issues

## Objective

Implement a GitHub Issues sync source for `taskmd sync`. This provider fetches issues from a GitHub repository and maps them to local taskmd markdown files.

## Tasks

- [ ] Implement the `Source` interface for GitHub Issues in `internal/sync/github/`
- [ ] Authenticate via GitHub token (from config or environment variable)
- [ ] Fetch issues from a configured repository using the GitHub API
- [ ] Map GitHub fields to taskmd frontmatter (labels to tags, state to status, milestone to priority, assignee, etc.)
- [ ] Support filtering by labels, milestone, or assignee in config
- [ ] Write tests with mocked GitHub API responses

## Config Example

```yaml
sources:
  - name: github
    type: github-issues
    repo: owner/repo
    token_env: GITHUB_TOKEN
    filters:
      labels: ["task", "feature"]
    field_map:
      labels: tags
      state: status
      assignee: assignee
```

## Acceptance Criteria

- `taskmd sync --source github` fetches issues and creates/updates markdown files
- GitHub-specific fields are mapped correctly to taskmd frontmatter
- Authentication works via environment variable or config
- Tests cover the provider with mocked API responses
