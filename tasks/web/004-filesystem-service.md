---
id: "004"
title: "Filesystem service (dir scanning, CRUD)"
status: completed
priority: high
effort: medium
dependencies:
  - "003"
tags:
  - core
  - filesystem
created: 2026-02-08
---

# Filesystem Service (Dir Scanning, CRUD)

## Objective

Build the filesystem service that scans directories for `.md` files and performs file-level CRUD operations. This sits between the API routes and the markdown parsing layer.

## Tasks

- [ ] Create `src/lib/filesystem.ts`
- [ ] Implement `scanDirectory(dirPath: string): Promise<Task[]>`
  - Read all `.md` files in the given directory (non-recursive, top-level only)
  - Parse each file using `parseTaskFile()`
  - Return array of `Task` objects sorted by ID
  - Skip files that fail to parse (log a warning)
- [ ] Implement `readTaskFile(filePath: string): Promise<Task>`
  - Read a single `.md` file and return parsed `Task`
- [ ] Implement `writeTaskFile(dirPath: string, task: Task): Promise<Task>`
  - Serialize the task and write to the file at `task.filePath`
  - If creating a new task (no `filePath`), generate the filename and set it
  - Update the `updated` timestamp on write
- [ ] Implement `deleteTaskFile(filePath: string): Promise<void>`
  - Remove the `.md` file from disk
- [ ] Validate that `dirPath` exists and is a directory before operations
- [ ] Handle file permission errors gracefully

## Acceptance Criteria

- `scanDirectory` returns all valid task files from a folder
- Files can be created, read, updated, and deleted
- New tasks get auto-generated filenames based on ID and title
- Invalid or unreadable files are skipped without crashing
- Error messages are descriptive for common issues (dir not found, permission denied)
