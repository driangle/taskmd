---
id: "01kz3c247"
title: "Unify version numbers across the three plugins"
status: completed
priority: low
effort: small
type: chore
phase: critical-feedback
dependencies: []
tags: [plugins, release, cleanup]
created_at: 2026-08-03
completed_at: 2026-08-21
---

# Unify version numbers across the three plugins

## Objective

The three plugins are versioned incoherently: `claude-code-plugin` is at `0.2.7` (tracks the
repo), `claude-code-plugin-lite` at `0.1.6`, and `claude-code-plugin-mcp` at `1.0.0`. The
marketplace presents them as one product family, so three unrelated version numbers give no
coherent story about compatibility or maturity. Decide on a versioning policy and align them.

## Tasks

- [x] Decide the versioning policy: lockstep with the repo, or independent semver per plugin
      — **independent semver**, recorded in `docs/adr/0003-plugin-versioning-policy.md`
- [x] ~~If lockstep: align all three `plugin.json` versions to the repo version~~ — not
      taken; `release.sh` no longer propagates the repo version into any plugin manifest
- [x] If independent: document each plugin's version line and stability guarantees in its README
      (`taskmd` `0.x`, `taskmd-lite` `0.x`, `taskmd-mcp` `1.x`; the mcp plugin had no README,
      so one was written)
- [x] Reconcile `.claude-plugin/marketplace.json` so listed versions match the plugin manifests
      — it carries no versions by design; `plugin.json` is the sole source of truth and
      `release.sh` fails the release if a `version` key appears in the marketplace manifest
- [x] Add the version bump(s) to the release checklist / `release` skill — `--plugin-taskmd-version`,
      `--plugin-lite-version`, `--plugin-mcp-version`, required when that plugin's directory
      changed since the last release tag

## Acceptance Criteria

- The three plugins follow one documented versioning policy
- Marketplace metadata and plugin manifests agree on versions
- The release process keeps plugin versions consistent going forward
