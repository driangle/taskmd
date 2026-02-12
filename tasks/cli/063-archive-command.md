---
id: "063"
title: "Implement archive command to hide completed/cancelled tasks"
status: pending
priority: medium
effort: large
dependencies: ["060"]
tags:
  - cli
  - go
  - commands
  - feature
created: 2026-02-12
---

# Implement Archive Command

## Objective

Add an `archive` command that allows users to hide tasks from regular CLI operations by moving them to an `archive/` subdirectory. This keeps the main task list clean while preserving completed or cancelled tasks for historical reference. Users can archive tasks by ID, by status (completed/cancelled), or by tag filter. An optional flag allows permanent deletion instead of archiving.

## Context

As projects mature, the number of completed and cancelled tasks grows, cluttering the task list and slowing down operations. Users need a way to "hide" these tasks from regular views without losing the historical record. Moving tasks to an `archive/` subdirectory is a simple, filesystem-native solution that:

- Keeps files organized and accessible
- Makes tasks invisible to the scanner by default
- Preserves directory structure for archived tasks
- Allows easy restoration (just move files back)
- Works naturally with version control

## Tasks

### Command Implementation

- [ ] Create `archive` command with cobra
  - Command signature: `taskmd archive [flags]`
  - Support multiple selection criteria (ID, status, tags)
  - Dry-run mode to preview what will be archived
  - Confirmation prompt before archiving (unless `--yes` flag)
  - Option to delete files instead of moving to archive
- [ ] Implement task selection logic
  - Select by task ID (single or multiple IDs)
  - Select by status (`--status completed`, `--status cancelled`)
  - Select by tag filter (`--tag deprecated`)
  - Combine multiple filters with AND logic
  - Select all completed: `--all-completed`
  - Select all cancelled: `--all-cancelled`
- [ ] Implement archive operation
  - Create `archive/` subdirectory if it doesn't exist
  - Preserve relative directory structure in archive
  - Move task files to archive location
  - Handle file conflicts (if archived file already exists)
  - Verify move succeeded before confirming
- [ ] Implement delete operation
  - Add `--delete` flag to permanently delete instead of archive
  - Require explicit confirmation for delete
  - Add `--force` flag to skip confirmation
  - Delete files permanently (no archive)
  - Log deleted files for safety

### Archive Directory Structure

- [ ] Define archive directory structure
  - Default location: `<task-root>/archive/`
  - Preserve subdirectory structure: `tasks/cli/001.md` → `tasks/archive/cli/001.md`
  - Alternative: flat structure with prefixes
- [ ] Update scanner to exclude archive directory by default
  - Skip `archive/` directories during scan
  - Add `--include-archived` flag to include archived tasks
  - Add `--archived-only` flag to show only archived tasks
- [ ] Add restore capability (future enhancement placeholder)
  - Document manual restoration process
  - Placeholder for `taskmd restore` command

### Output & Reporting

- [ ] Implement dry-run mode
  - `--dry-run` flag to preview without making changes
  - Show list of tasks that would be archived/deleted
  - Show source and destination paths
  - Show total count
- [ ] Implement progress reporting
  - Show progress during archive operation
  - List each file being moved/deleted
  - Summary at the end (X tasks archived, Y tasks deleted)
  - Handle errors gracefully with clear messages
- [ ] Add verbose output
  - Show detailed file paths
  - Show archive destination
  - Show any warnings or conflicts

### Flags and Options

- [ ] Define command flags
  - `--id <task-id>` - Archive specific task by ID (repeatable)
  - `--status <status>` - Archive tasks with specific status
  - `--tag <tag>` - Archive tasks with specific tag
  - `--all-completed` - Archive all completed tasks
  - `--all-cancelled` - Archive all cancelled tasks
  - `--delete` - Delete files instead of archiving
  - `--dry-run` - Preview without making changes
  - `--yes` / `-y` - Skip confirmation prompt
  - `--force` / `-f` - Force delete without confirmation
  - `--include-archived` - Include already archived tasks (for other commands)
  - `--archived-only` - Show only archived tasks (for other commands)

### Safety & Validation

- [ ] Implement safety checks
  - Confirm before archiving (unless `--yes`)
  - Extra confirmation for delete (unless `--force`)
  - Prevent archiving tasks with incomplete dependencies
  - Warn if archiving in-progress tasks
  - Check for file write permissions
- [ ] Implement validation
  - Validate task IDs exist before archiving
  - Validate status values
  - Validate tag values
  - Error if no tasks match criteria
  - Error if archive destination is invalid
- [ ] Handle edge cases
  - Task file already in archive
  - File conflicts in archive directory
  - Symlinks and special files
  - Read-only files
  - Large number of files

### Testing

- [ ] Unit tests for archive logic
  - Test task selection by ID
  - Test task selection by status
  - Test task selection by tag
  - Test combined filters
- [ ] Integration tests for file operations
  - Test moving files to archive
  - Test preserving directory structure
  - Test file deletion
  - Test error handling
- [ ] E2E tests for full workflow
  - Archive completed tasks
  - Archive cancelled tasks
  - Archive by tag
  - Delete tasks
  - Dry-run mode
  - Confirmation prompts
- [ ] Test scanner exclusion
  - Verify archived tasks don't appear in list
  - Verify --include-archived works
  - Verify --archived-only works

### Documentation

- [ ] Update CLI help text
- [ ] Add archive command to user guide
- [ ] Document archive directory structure
- [ ] Add examples to README
- [ ] Document restoration process

## Acceptance Criteria

- ✅ Users can archive tasks by ID: `taskmd archive --id 001 --id 002`
- ✅ Users can archive all completed tasks: `taskmd archive --all-completed`
- ✅ Users can archive all cancelled tasks: `taskmd archive --all-cancelled`
- ✅ Users can archive by status: `taskmd archive --status completed`
- ✅ Users can archive by tag: `taskmd archive --tag deprecated`
- ✅ Users can combine filters: `taskmd archive --status completed --tag old`
- ✅ Archived tasks move to `archive/` subdirectory with preserved structure
- ✅ Archived tasks no longer appear in `taskmd list` by default
- ✅ Archived tasks no longer appear in `taskmd board` by default
- ✅ Archived tasks no longer appear in `taskmd stats` by default
- ✅ `--include-archived` flag makes archived tasks visible
- ✅ `--archived-only` flag shows only archived tasks
- ✅ `--dry-run` shows what would be archived without making changes
- ✅ Confirmation prompt shown before archiving (unless `--yes`)
- ✅ `--delete` flag permanently deletes files instead of archiving
- ✅ Extra confirmation required for `--delete` (unless `--force`)
- ✅ Archive operation preserves directory structure
- ✅ Clear error messages for invalid operations
- ✅ Summary shown after archive operation
- ✅ All existing tests pass
- ✅ New tests cover archive functionality

## Implementation Notes

### Archive Directory Structure

**Approach 1: Preserve full structure (Recommended)**
```
tasks/
├── cli/
│   ├── 001-setup.md          (active)
│   └── 002-auth.md           (active)
├── web/
│   └── 003-ui.md             (active)
└── archive/
    ├── cli/
    │   ├── 010-old-task.md   (archived)
    │   └── 011-deprecated.md (archived)
    └── web/
        └── 012-old-ui.md     (archived)
```

**Approach 2: Flat archive**
```
tasks/
├── cli/
│   ├── 001-setup.md          (active)
│   └── 002-auth.md           (active)
└── archive/
    ├── cli-010-old-task.md   (archived, prefixed)
    ├── cli-011-deprecated.md (archived, prefixed)
    └── web-012-old-ui.md     (archived, prefixed)
```

**Recommended:** Approach 1 - preserves structure and makes restoration easier.

### Scanner Updates

Update the scanner to skip `archive/` directories by default:

```go
// In scanner.go
func (s *Scanner) shouldSkipDir(path string) bool {
    base := filepath.Base(path)

    // Skip hidden directories
    if strings.HasPrefix(base, ".") {
        return true
    }

    // Skip archive directory (unless explicitly included)
    if base == "archive" && !s.includeArchived {
        return true
    }

    return false
}
```

### Command Signature

```bash
# Archive by ID
taskmd archive --id 042
taskmd archive --id 042 --id 043 --id 044

# Archive by status
taskmd archive --status completed
taskmd archive --status cancelled
taskmd archive --all-completed
taskmd archive --all-cancelled

# Archive by tag
taskmd archive --tag deprecated
taskmd archive --tag post-mvp

# Combine filters (AND logic)
taskmd archive --status completed --tag old

# Dry run
taskmd archive --all-completed --dry-run

# Delete instead of archive
taskmd archive --id 042 --delete
taskmd archive --all-completed --delete --force

# Skip confirmation
taskmd archive --all-completed --yes
```

### Archive Operation Logic

```go
func archiveTask(task *model.Task, archiveRoot string) error {
    // Determine source path (current task file location)
    srcPath := task.FilePath

    // Determine destination path in archive
    // Preserve relative structure from task root
    relPath, _ := filepath.Rel(taskRoot, srcPath)
    destPath := filepath.Join(archiveRoot, relPath)

    // Create destination directory
    destDir := filepath.Dir(destPath)
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return fmt.Errorf("failed to create archive directory: %w", err)
    }

    // Check for conflicts
    if _, err := os.Stat(destPath); err == nil {
        return fmt.Errorf("archived file already exists: %s", destPath)
    }

    // Move file to archive
    if err := os.Rename(srcPath, destPath); err != nil {
        return fmt.Errorf("failed to move file to archive: %w", err)
    }

    return nil
}
```

### Delete Operation Logic

```go
func deleteTask(task *model.Task) error {
    // Extra confirmation for delete
    fmt.Printf("⚠️  WARNING: Permanently deleting task %s: %s\n", task.ID, task.Title)
    fmt.Print("Type 'DELETE' to confirm: ")

    var confirm string
    fmt.Scanln(&confirm)

    if confirm != "DELETE" {
        return fmt.Errorf("delete cancelled")
    }

    // Delete file
    if err := os.Remove(task.FilePath); err != nil {
        return fmt.Errorf("failed to delete file: %w", err)
    }

    return nil
}
```

### Task Selection

```go
func selectTasksToArchive(tasks []*model.Task, criteria ArchiveCriteria) []*model.Task {
    var selected []*model.Task

    for _, task := range tasks {
        // Check ID filter
        if len(criteria.IDs) > 0 {
            found := false
            for _, id := range criteria.IDs {
                if task.ID == id {
                    found = true
                    break
                }
            }
            if !found {
                continue
            }
        }

        // Check status filter
        if criteria.Status != "" && task.Status != criteria.Status {
            continue
        }

        // Check tag filter
        if criteria.Tag != "" {
            hasTag := false
            for _, tag := range task.Tags {
                if tag == criteria.Tag {
                    hasTag = true
                    break
                }
            }
            if !hasTag {
                continue
            }
        }

        // Task matches all criteria
        selected = append(selected, task)
    }

    return selected
}
```

### Dry Run Output

```bash
$ taskmd archive --all-completed --dry-run

📋 Tasks to be archived (dry run):

  001  Setup project                    tasks/cli/001-setup.md
       → tasks/archive/cli/001-setup.md

  015  Implement authentication         tasks/cli/015-auth.md
       → tasks/archive/cli/015-auth.md

  023  Build UI components             tasks/web/023-ui.md
       → tasks/archive/web/023-ui.md

Total: 3 tasks would be archived

Run without --dry-run to proceed.
```

### Confirmation Prompt

```bash
$ taskmd archive --all-completed

📋 Tasks to be archived: 3

  001  Setup project (completed)
  015  Implement authentication (completed)
  023  Build UI components (completed)

Archive these tasks? [y/N]: y

Archiving tasks...
  ✓ Archived 001: tasks/cli/001-setup.md → tasks/archive/cli/001-setup.md
  ✓ Archived 015: tasks/cli/015-auth.md → tasks/archive/cli/015-auth.md
  ✓ Archived 023: tasks/web/023-ui.md → tasks/archive/web/023-ui.md

✨ Successfully archived 3 tasks
```

### Scanner Integration

Add flags to scanner for archive handling:

```go
type ScanOptions struct {
    IncludeArchived bool  // Include archived tasks in scan
    ArchivedOnly    bool  // Only scan archived tasks
}

func (s *Scanner) Scan() (*ScanResult, error) {
    // ... existing code ...

    // Determine archive behavior
    if s.opts.ArchivedOnly {
        // Only scan archive directory
        return s.scanDirectory(filepath.Join(s.rootDir, "archive"))
    }

    // Regular scan (with or without archives)
    result := s.scanDirectory(s.rootDir)

    return result, nil
}
```

### Global Flags

Add these flags to all relevant commands (list, board, stats, graph, next):

```go
// In root command or shared flags
rootCmd.PersistentFlags().BoolVar(&includeArchived, "include-archived", false, "include archived tasks")
rootCmd.PersistentFlags().BoolVar(&archivedOnly, "archived-only", false, "show only archived tasks")
```

## Examples

### Archive Completed Tasks

```bash
# Archive all completed tasks
taskmd archive --all-completed

# Preview what would be archived
taskmd archive --all-completed --dry-run

# Archive without confirmation
taskmd archive --all-completed --yes
```

### Archive Cancelled Tasks

```bash
# Archive all cancelled tasks
taskmd archive --all-cancelled

# Archive completed AND cancelled
taskmd archive --status completed
taskmd archive --status cancelled
```

### Archive by Tag

```bash
# Archive deprecated tasks
taskmd archive --tag deprecated

# Archive completed post-mvp tasks
taskmd archive --status completed --tag post-mvp
```

### Delete Tasks

```bash
# Delete specific task
taskmd archive --id 042 --delete

# Delete all cancelled tasks (with confirmation)
taskmd archive --all-cancelled --delete

# Delete with force (skip confirmation)
taskmd archive --all-cancelled --delete --force
```

### View Archived Tasks

```bash
# List including archived
taskmd list --include-archived

# List only archived tasks
taskmd list --archived-only

# Stats including archived
taskmd stats --include-archived
```

### Manual Restoration

To restore an archived task manually:

```bash
# Move from archive back to original location
mv tasks/archive/cli/042-task.md tasks/cli/042-task.md

# Or restore entire directory
mv tasks/archive/cli/* tasks/cli/
```

## References

- Scanner implementation: `apps/cli/internal/scanner/scanner.go`
- File operations: Go `os` and `filepath` packages
- Similar commands: `update`, `validate`
- Archive pattern: existing `tasks/cli/archive/` directory

## Related Tasks

- Future: `taskmd restore` command to unarchive tasks
- Future: `taskmd clean` command to delete old archives
- Future: Archive with compression (tar.gz)
- Future: Archive metadata file tracking what was archived when

## Security Considerations

- **File permissions** - Ensure user has write permissions
- **Path traversal** - Validate paths to prevent escaping task root
- **Confirmation** - Always confirm destructive operations
- **Backup** - Recommend users commit to git before archiving
- **Atomic operations** - Use atomic file moves where possible

## Performance Considerations

- **Large archives** - Handle large numbers of files efficiently
- **Progress reporting** - Show progress for long operations
- **Parallel operations** - Consider parallelizing file moves
- **Directory creation** - Create directories in batch

## Notes

- Archiving is reversible by manually moving files back
- Deleting is permanent - use with caution
- Consider integrating with git for version control
- Archive directory should be in `.gitignore` or committed based on team preference
- Future enhancement: automated archiving based on age or criteria
