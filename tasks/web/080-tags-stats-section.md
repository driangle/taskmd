---
id: "080"
title: "Add tags section to Stats view in web interface"
status: completed
priority: medium
effort: small
tags:
  - mvp
  - web
created: 2026-02-14
---

# Add Tags Section to Stats View in Web Interface

## Objective

Add a "Tags" section nested under the existing Stats view in the web interface. It should display all tags used across task files with their task counts, sorted by count descending (most used first) — mirroring what the CLI `taskmd tags` command provides.

## Tasks

- [ ] Add a tags aggregation endpoint or include tag data in the existing stats API response
- [ ] Create a Tags section component that displays tag names and counts
- [ ] Sort tags by count descending, with alphabetical tie-breaking
- [ ] Integrate the Tags section into the existing Stats view
- [ ] Style consistently with the rest of the Stats view
- [ ] Add tests for the tags aggregation logic and component rendering

## Acceptance Criteria

- The Stats view includes a "Tags" section showing all tags with task counts
- Tags are sorted by count descending (most used first), alphabetical for ties
- Data matches what `taskmd tags` returns for the same task set
- The section handles the empty state gracefully (no tags found)
