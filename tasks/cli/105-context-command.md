---
id: "105"
title: "Add context command for AI agent task context"
status: pending
priority: medium
effort: medium
tags:
  - ai
  - dx
  - mvp
touches:
  - cli
created: 2026-02-14
---

# Add Context Command for AI Agent Task Context

## Objective

Add a new `taskmd context --task-id <ID>` command that gathers all relevant context for a task into a single output, optimized for LLM consumption. The command reads the task file, resolves its `touches` scopes from `.taskmd.yaml`, and outputs the task description along with referenced files and directories in a structured format an AI agent can use to understand the full context of a task.

## Tasks

- [ ] Create `internal/cli/context.go` with the `context` command and flags
- [ ] Implement task lookup by `--task-id` flag
- [ ] Read `.taskmd.yaml` to resolve `touches` scopes to file/directory paths
- [ ] Implement three detail modes controlled by a `--detail` flag:
  - `names` (default) — output only filenames and directory names
  - `files` — include file contents for directly referenced files in scopes, but only directory names for directory scopes
  - `full` — include ALL file contents, expanding directories to include their files
- [ ] For directory scopes in `files` mode: if the directory contains very few files with fewer than 100 total lines of code, inline their contents instead of just listing the directory
- [ ] Format output with clear headings/separators suitable for LLM context (task description section, then files section with path headers)
- [ ] Add comprehensive tests in `internal/cli/context_test.go`
- [ ] Register command with `rootCmd`

## Acceptance Criteria

- `taskmd context --task-id 042` outputs the task description and lists touched scope paths
- `--detail names` (default) outputs only file/directory names from scopes
- `--detail files` outputs file contents for individual file paths and directory names for directory scopes (with small-directory auto-expansion)
- `--detail full` outputs all file contents, expanding directories recursively
- Output is clearly structured with headings and separators for LLM readability
- Works with `--task-dir` flag for custom task directories
- Gracefully handles missing scopes, missing files, or tasks with no `touches`
- Tests cover all three detail modes, edge cases, and error handling
