---
id: "003"
title: "Markdown parsing and serialization service"
status: pending
priority: high
effort: medium
dependencies:
  - "002"
tags:
  - core
  - parsing
created: 2026-02-08
---

# Markdown Parsing and Serialization Service

## Objective

Build the service layer that converts between `.md` files (with YAML frontmatter) and `Task` objects. This must handle round-trip parsing — reading a file and writing it back should preserve the original content as closely as possible.

## Tasks

- [ ] Create `src/lib/markdown.ts`
- [ ] Implement `parseTaskFile(filePath: string, content: string): Task`
  - Use `gray-matter` to extract frontmatter and body
  - Map frontmatter fields to `TaskFrontmatter` type
  - Handle missing/optional fields with sensible defaults
  - Include `filePath` in the returned `Task` object
- [ ] Implement `serializeTask(task: Task): string`
  - Use `gray-matter`'s `stringify()` to combine frontmatter + body
  - Preserve field ordering in frontmatter where possible
- [ ] Implement `generateTaskId(): string`
  - Generate zero-padded numeric IDs (e.g., "001", "015")
  - Accept an optional list of existing IDs to find the next available
- [ ] Implement `generateFileName(id: string, title: string): string`
  - Produce slug-style filenames like `001-my-task-title.md`
- [ ] Handle edge cases: files with no frontmatter, empty body, invalid YAML

## Acceptance Criteria

- `parseTaskFile` correctly extracts all frontmatter fields and the markdown body
- `serializeTask` produces valid markdown with YAML frontmatter
- Round-trip: `serializeTask(parseTaskFile(path, content))` preserves content
- Invalid files don't crash — they return sensible defaults or are skipped with warnings
