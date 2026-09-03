---
id: "006"
title: Export reports to CSV
status: pending
priority: high
effort: medium
type: feature
tags: [reports, export]
owner: jordan
phase: polish
created: 2026-03-06
---

# Export reports to CSV

## Objective

Let users download any report table as a CSV file so results can be shared with
people who do not have dashboard access.

## Tasks

- [ ] Add a serializer that turns a report result set into CSV
- [ ] Add a download button to the report toolbar
- [ ] Stream large exports instead of buffering the whole file

## Acceptance Criteria

- A report of any size downloads as a valid CSV with a header row
- Field values containing commas, quotes or newlines are correctly escaped
- Exporting a 100k-row report does not exhaust server memory
