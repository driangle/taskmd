---
id: "110"
title: "Add full-text search to web interface"
status: completed
priority: medium
effort: medium
tags:
  - search
  - dx
  - mvp
touches:
  - web
created: 2026-02-15
---

# Add Full-Text Search to Web Interface

## Objective

Add a search feature to the web interface that performs full-text search across all task titles and markdown bodies, similar to the CLI `taskmd search` command (task 106). Users should be able to quickly find tasks by keyword without scrolling through lists or manually browsing.

## Tasks

- [x] Add a search input component to the web layout (e.g., header or sidebar)
- [x] Implement a search API endpoint that accepts a query string
- [x] Perform case-insensitive matching across task titles and body content
- [x] Return matching tasks with ID, title, status, match location, and a context snippet
- [x] Display search results in a results list with highlighted match terms
- [x] Show match location indicator (title, body, or both)
- [x] Show a context snippet around the match in the body
- [x] Handle empty results with a clear "no results" message
- [x] Support keyboard shortcut to focus the search input (e.g., Cmd+K or /)
- [x] Add tests for the search API endpoint
- [x] Add tests for the search UI component

## Acceptance Criteria

- Typing a query returns all tasks mentioning that term in title or body
- Search is case-insensitive
- Results show task ID, title, status, and a context snippet around the match
- Matched terms are visually highlighted in the results
- Empty results display a helpful message
- Search is responsive and works with the existing task data pipeline
