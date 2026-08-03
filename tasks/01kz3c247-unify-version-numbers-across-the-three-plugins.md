---
id: "01kz3c247"
title: "Unify version numbers across the three plugins"
status: pending
priority: low
effort: small
type: chore
phase: critical-feedback
dependencies: []
tags: [plugins, release, cleanup]
created_at: 2026-08-03
---

# Unify version numbers across the three plugins

## Objective

The three plugins are versioned incoherently: `claude-code-plugin` is at `0.2.7` (tracks the
repo), `claude-code-plugin-lite` at `0.1.6`, and `claude-code-plugin-mcp` at `1.0.0`. The
marketplace presents them as one product family, so three unrelated version numbers give no
coherent story about compatibility or maturity. Decide on a versioning policy and align them.

## Tasks

- [ ] Decide the versioning policy: lockstep with the repo, or independent semver per plugin
- [ ] If lockstep: align all three `plugin.json` versions to the repo version and add a release
      step that bumps them together
- [ ] If independent: document each plugin's version line and stability guarantees in its README
- [ ] Reconcile `.claude-plugin/marketplace.json` so listed versions match the plugin manifests
- [ ] Add the version bump(s) to the release checklist / `release` skill

## Acceptance Criteria

- The three plugins follow one documented versioning policy
- Marketplace metadata and plugin manifests agree on versions
- The release process keeps plugin versions consistent going forward
