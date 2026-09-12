---
id: "197"
title: "Log debug message when directories are skipped via ignore config"
status: pending
priority: low
effort: small
type: improvement
tags: [cli, scanner, logging]
created: 2026-02-22
---

# Log debug message when directories are skipped via ignore config

## Objective

When `--verbose` is active and the scanner skips a directory because it matches a user-configured `ignore` entry, emit a debug message to stderr so users can confirm their ignore rules are taking effect.

## Tasks

- [ ] In `Scanner.shouldSkipDirectory`, log a message (to stderr) when a directory is skipped due to a user-configured ignore entry (not the built-in defaults)
- [ ] Ensure the message only appears in verbose mode
- [ ] Add a test verifying the debug output is emitted when verbose is true and an ignore dir is hit

## Acceptance Criteria

- Running any command with `--verbose` and an `ignore` entry in `.taskmd.yaml` prints a message like `Skipping ignored directory: drafts` to stderr when that directory is encountered
- No output is produced when `--verbose` is not set
- Existing tests continue to pass
