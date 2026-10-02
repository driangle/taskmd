## 2026-10-02T10:40:00Z

Started. Strict phase tiers live in `sdk/go/next.assignPhaseTiers`; every `next`
surface (CLI, `--all-projects`, MCP, web `/api/next`, static export) builds
`next.Options` itself, so the setting has to be threaded through all five.

**Key name:** kept `next.unphased`. Nested sections (`web:`, `id:`, `todos:`) are
the existing convention, so a `next:` section fits; `next` is added to the
validator's known top-level keys.

## 2026-10-02T11:18:40Z

Done.

**SDK** (`sdk/go/next`): new `UnphasedPlacement` type (`current`, `last`) and
`Options.UnphasedPlacement`; the zero value keeps today's behaviour, so this is an
additive patch bump. `last` puts unphased tasks at tier `len(phases)` and unknown
phases at `len(phases)+1`. Pulled the current-tier computation into
`currentPhaseTier`.

**CLI** (`internal/cli/next_config.go`): `resolveUnphasedPlacement` (viper) and
`loadProjectUnphasedPlacement` (per-project yaml for `--all-projects`). Unlike
`effort`, an invalid value is a hard error — follows the `worktree_scope`
precedent, since silently ranking differently from the configured intent is
worse than refusing. `next`, `mcp`, `web start` and `web export` fail on it;
`--all-projects` skips the bad project with a warning (existing behaviour for
per-project errors); `validate` reports it as a config error.

**Web:** `handleNext` now takes the server `Config` instead of a growing
positional list.

**Tests:** SDK tier tests, CLI tests (both values, default, invalid, no-phases,
`--strict-phases=false`, per-project `--all-projects`, invalid project skipped),
MCP/web/export tests, and e2e through a real `.taskmd.yaml` (incl. `validate`).

**Docs:** config reference (table row + "Unphased tasks" section), cli guide
"Phase ordering", mcp guide, best-practices, and `next --help`.

**Open item:** the web UI's per-project switch resolves phases per project but
uses the server's startup `next.unphased` (the same gap that `effort` already has).

## 2026-10-02T11:32:00Z

Completed. Backlog reconciled: filed 01m3yn6f8 (resolve efforts and
next.unphased per project in the web project switcher) for the open item above.
