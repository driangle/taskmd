---
id: "01kzrmtj4"
title: "Serve the configured effort vocabulary to the web UI"
status: pending
priority: low
effort: small
dependencies: ["01kzqwfns"]
tags: ["web", "effort", "config"]
created_at: 2026-08-11
---

# Serve the configured effort vocabulary to the web UI

## Objective

Task `01kzqwfns` made the `effort` vocabulary configurable via `effort: [...]` in
`.taskmd.yaml`, and threaded it through the CLI, the MCP tools, and the web
*server*. The web *frontend* still hard-codes it.

`apps/web/src/components/tasks/TaskTable/constants.ts` exports a fixed
`EFFORTS` constant, used by the filter checkboxes in `TaskTable.tsx` and the
effort dropdown in `TaskEditFormFields.tsx`. Under a custom vocabulary those two
pickers list values the project does not use, and omit the ones it does.

Custom values still round-trip correctly — tasks display and save fine; the gap
is limited to the two pickers.

The server already resolves the vocabulary (`web.Config.Efforts`, populated from
`resolveEffortScale()` in `apps/cli/internal/cli/web.go`), so the work is to
expose it over the API and consume it in the frontend.

## Tasks

- [ ] Add the effort vocabulary to the `/api/config` response
      (`configResponse` in `apps/cli/internal/web/handlers.go`, which already
      carries `phases`), and to the static export's `config.json`
- [ ] Replace the `EFFORTS` constant with a value read from the config API,
      keeping the current values as the fallback when the key is absent
- [ ] Update `TaskTable.tsx` (filter checkboxes, `selectAll` reset, URL sync) and
      `TaskEditFormFields.tsx` (dropdown) to use it
- [ ] Update the web test fixtures in `apps/web/src/test-utils/fixtures.ts`
- [ ] Add a frontend test covering a project with a custom vocabulary

## Acceptance Criteria

- With no `effort` config, the web UI behaves exactly as today
- With `effort: [xs, s, m, l, xl]`, the filter checkboxes and the edit-form
  dropdown offer exactly those five values
- The static export (`taskmd web export`) carries the vocabulary too, so an
  exported site matches the live server
- `cd apps/web && pnpm test` and `pnpm build` pass
