---
id: "079"
title: "Add readonly flag to web interface"
status: completed
priority: medium
effort: medium
tags:
  - web
  - cli
  - mvp
created: 2026-02-14
---

# Add Readonly Flag to Web Interface

## Objective

Allow the CLI `web` command to start the web UI in readonly mode. When this flag is active, all editing features (inline editing, task creation, status changes, etc.) are hidden from the interface, making it a view-only dashboard.

## Tasks

- [x] Add `--readonly` flag to the CLI `web` command
- [x] Pass the readonly flag to the web server (e.g. via API endpoint or embedded config)
- [x] Expose a `/api/config` or similar endpoint that returns `{ readonly: true/false }`
- [x] Add a React context or hook (`useReadonly`) to access the flag client-side
- [x] Hide inline editing controls when readonly is active
- [x] Hide task creation dialog/button when readonly is active
- [x] Hide status change controls when readonly is active
- [x] Hide any delete or destructive action buttons when readonly is active
- [x] Show a subtle "Read Only" badge in the UI when the flag is on
- [x] Ensure API write endpoints return 403 when readonly is enabled

## Acceptance Criteria

- `taskmd web --readonly` starts the web UI with all editing features hidden
- `taskmd web` (without flag) behaves as before with full editing capabilities
- No editing controls are visible or accessible in readonly mode
- API write endpoints are also protected (not just hidden in the UI)
- A visual indicator shows the user that readonly mode is active
