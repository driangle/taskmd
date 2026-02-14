---
id: "097"
title: "Add ignore directories option to .taskmd.yaml config"
status: pending
priority: medium
effort: small
dependencies: ["056"]
tags:
  - cli
  - go
  - configuration
  - mvp
created: 2026-02-14
---

# Add Ignore Directories Option to .taskmd.yaml Config

## Objective

Allow users to specify directories to exclude from task scanning via the `.taskmd.yaml` configuration file. This prevents the scanner from traversing irrelevant directories (e.g., `archive/`, `node_modules/`, vendor directories) and keeps task listings clean.

## Context

The scanner (`internal/scanner/scanner.go`) currently walks the entire task directory tree, only skipping hidden directories (those starting with `.`). Users need a way to exclude specific directories from scanning without moving them outside the task root.

## Config Format

```yaml
# .taskmd.yaml
ignore:
  - archive
  - templates
  - drafts
```

## Tasks

- [ ] Add `ignore` field support to config loading (Viper)
- [ ] Pass ignore list from config into the scanner
- [ ] Update `Scanner` to skip directories matching the ignore list during `filepath.WalkDir`
- [ ] Match on directory name (not full path) for simplicity
- [ ] Ensure hidden directory skipping still works alongside ignore list
- [ ] Add tests for ignore behavior (single dir, multiple dirs, nested dirs)
- [ ] Test that config ignore works with both project-level and global config

## Acceptance Criteria

- Directories listed in `ignore` are skipped during task scanning
- Existing hidden-directory skipping behavior is preserved
- Config precedence is respected (project config overrides global)
- Scanner works correctly when no `ignore` config is present (empty/missing = ignore nothing)
- Tests cover ignore matching and edge cases
