---
id: "093"
title: "VSCode extension for task frontmatter validation"
status: pending
priority: medium
effort: large
tags: []
touches:
  - vscode
created: 2026-02-14
---

# VSCode Extension for Task Frontmatter Validation

## Objective

Create a VSCode extension that validates the YAML frontmatter schema in taskmd markdown files. The extension should provide real-time diagnostics (errors, warnings) as users edit task files, ensuring fields like `id`, `title`, `status`, `priority`, `effort`, `dependencies`, and `tags` conform to the taskmd specification.

## Tasks

- [ ] Scaffold a new VSCode extension project (TypeScript)
- [ ] Define the taskmd frontmatter JSON schema based on `docs/taskmd_specification.md`
- [ ] Parse YAML frontmatter from markdown files on open and on change
- [ ] Validate required fields (`id`, `title`) and report errors if missing
- [ ] Validate enum fields (`status`, `priority`, `effort`) against allowed values
- [ ] Validate field types (e.g., `dependencies` is an array of strings, `tags` is an array)
- [ ] Validate date format for `created` field (`YYYY-MM-DD`)
- [ ] Display diagnostics inline in the editor (squiggly underlines, Problems panel)
- [ ] Add autocompletion for enum fields (status, priority, effort values)
- [ ] Configure activation to only trigger on markdown files matching task patterns
- [ ] Write unit tests for the validation logic
- [ ] Package and document installation instructions

## Acceptance Criteria

- The extension activates on markdown files in a `tasks/` directory
- Missing required fields (`id`, `title`) are reported as errors
- Invalid enum values are reported as errors with suggestions
- Invalid field types (e.g., `dependencies` as a string instead of array) are flagged
- Diagnostics appear in real-time as the user types
- The extension does not interfere with non-taskmd markdown files
- Autocompletion works for `status`, `priority`, and `effort` fields
