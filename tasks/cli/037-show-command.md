---
id: "cli-037"
title: "Add show command with fuzzy matching"
status: pending
priority: high
effort: medium
tags:
  - cli
  - feature
  - fuzzy-matching
  - mvp
created: 2026-02-08
---

# Add Show Command with Fuzzy Matching

## Objective

Implement a `taskmd show <task_id or task_name>` command that allows users to view detailed information about a specific task. The command should support both exact matching and fuzzy matching for task identification.

## Background

Currently, users can list tasks and view summaries, but there's no dedicated command to show a single task in detail. This command will provide:
- Full task metadata
- Complete task description
- Task relationships (dependencies, blockers)
- File location
- Easy-to-read formatting

## Requirements

### Core Functionality

1. **Exact Match Priority**
   - First attempt exact match by task ID (e.g., `cli-037`)
   - Then attempt exact match by task title
   - If exact match found, display task details immediately

2. **Fuzzy Matching Fallback**
   - If no exact match found, use fuzzy matching algorithm
   - Search across both task IDs and titles
   - Present top N matches (e.g., top 5) ranked by similarity
   - **DO NOT auto-select** - require user to choose from options
   - Use interactive selection (numbered list or arrow keys)

3. **No Match Handling**
   - If no fuzzy matches found (or all below threshold), print: "task not found"
   - Optionally suggest similar tasks or common typos

### Command Interface

```bash
# Basic usage
taskmd show <task_id_or_name>

# Examples
taskmd show cli-037
taskmd show "Add show command"
taskmd show sho  # fuzzy match: shows "show command" as option

# Flags
--dir <path>         # Directory to scan (default: current)
--format <format>    # Output format: text, json, yaml (default: text)
--exact              # Disable fuzzy matching, exact only
--threshold <float>  # Fuzzy match threshold 0.0-1.0 (default: 0.6)
```

### Output Format

**Text format** (default):
```
Task: cli-037
Title: Add show command with fuzzy matching
Status: pending
Priority: medium
Effort: medium
Tags: cli, feature, fuzzy-matching
Created: 2026-02-08
File: tasks/cli/037-show-command.md

Description:
─────────────────────────────────────────────────
[Full markdown content here, properly formatted]
─────────────────────────────────────────────────

Dependencies:
  Depends on: [list of tasks]
  Blocks: [list of tasks]
```

**JSON format** (`--format json`):
```json
{
  "id": "cli-037",
  "title": "Add show command with fuzzy matching",
  "status": "pending",
  "priority": "medium",
  "effort": "medium",
  "tags": ["cli", "feature", "fuzzy-matching"],
  "created": "2026-02-08",
  "file_path": "tasks/cli/037-show-command.md",
  "content": "...",
  "dependencies": {
    "depends_on": [],
    "blocks": []
  }
}
```

### Fuzzy Matching Behavior

When fuzzy matching is triggered:

```
No exact match found for "sho". Did you mean:

1. cli-037: Add show command with fuzzy matching (95% match)
2. cli-032: Next command (45% match)
3. web-017: Task detail view (42% match)

Enter selection (1-3), or 0 to cancel: _
```

User must explicitly choose an option. Ctrl+C or entering 0 cancels.

## Implementation Tasks

### Phase 1: Basic Command Structure
- [ ] Create `internal/cli/show.go`
- [ ] Define command with basic flag support
- [ ] Register command with root command
- [ ] Implement exact match by task ID
- [ ] Implement exact match by task title
- [ ] Add basic text output formatter
- [ ] Add JSON/YAML output formatters

### Phase 2: Fuzzy Matching
- [ ] Research and choose fuzzy matching library (e.g., `github.com/sahilm/fuzzy` or `github.com/lithammer/fuzzysearch`)
- [ ] Implement fuzzy search across task IDs and titles
- [ ] Rank matches by similarity score
- [ ] Implement threshold filtering (default 0.6)
- [ ] Create interactive selection UI
- [ ] Handle user input (1-N, 0 for cancel, Ctrl+C)

### Phase 3: Polish
- [ ] Add color coding for better readability (use existing color helpers)
- [ ] Format markdown content nicely in terminal
- [ ] Add syntax highlighting for code blocks in description (optional)
- [ ] Implement `--exact` flag to disable fuzzy matching
- [ ] Add helpful error messages
- [ ] Consider pagination for very long task descriptions

### Phase 4: Testing
- [ ] Test exact match by ID
- [ ] Test exact match by title
- [ ] Test fuzzy matching with typos
- [ ] Test fuzzy matching with partial matches
- [ ] Test threshold filtering
- [ ] Test "task not found" scenario
- [ ] Test all output formats
- [ ] Test with special characters in task names
- [ ] Test user interaction (selection, cancellation)
- [ ] Add test file `internal/cli/show_test.go` with comprehensive coverage

### Phase 5: Documentation
- [ ] Add command to CLI help text
- [ ] Add examples to CLAUDE.md
- [ ] Update README if applicable
- [ ] Add inline code documentation

## Technical Considerations

### Fuzzy Matching Libraries

**Option 1: `github.com/sahilm/fuzzy`**
- Pros: Lightweight, simple API, good for substring matching
- Cons: Less sophisticated scoring

**Option 2: `github.com/lithammer/fuzzysearch`**
- Pros: Levenshtein distance, handles typos well
- Cons: May be slower for large task sets

**Option 3: Custom implementation**
- Use combination of:
  - Prefix matching (highest priority)
  - Substring matching
  - Levenshtein distance for typos
  - Token-based matching for multi-word titles

### Interactive Selection

Use standard input/output for selection:
```go
fmt.Fprintf(os.Stderr, "No exact match found for %q. Did you mean:\n\n", query)
for i, match := range matches {
    fmt.Fprintf(os.Stderr, "%d. %s: %s (%.0f%% match)\n",
        i+1, match.ID, match.Title, match.Score*100)
}
fmt.Fprintf(os.Stderr, "\nEnter selection (1-%d), or 0 to cancel: ", len(matches))

var choice int
fmt.Fscanln(os.Stdin, &choice)
```

### Performance

- Fuzzy matching on large repos (1000+ tasks) should complete in < 500ms
- Consider caching task index if performance is an issue
- For very large repos, might need to limit fuzzy search scope

## Edge Cases

- Task with special characters in ID or title
- Multiple tasks with very similar names
- Very long task descriptions (pagination?)
- Task files with parsing errors
- Concurrent modifications to task files
- Non-ASCII characters in fuzzy matching
- Empty task directory
- Invalid task ID format

## Acceptance Criteria

- ✅ Exact match by task ID works correctly
- ✅ Exact match by task title works correctly
- ✅ Fuzzy matching presents options when no exact match
- ✅ User can select from fuzzy match options
- ✅ User can cancel fuzzy match selection
- ✅ "task not found" message displays when no matches
- ✅ All output formats (text, JSON, YAML) work correctly
- ✅ `--exact` flag disables fuzzy matching
- ✅ `--threshold` flag controls match sensitivity
- ✅ Comprehensive tests cover all scenarios
- ✅ Command help text is clear and accurate
- ✅ Performance is acceptable (< 500ms for 1000 tasks)

## Related Tasks

- `cli-018`: List command (shows how to scan and display tasks)
- `cli-022`: Graph command (shows dependency resolution)
- `cli-032`: Next command (similar single-task display logic)
- `web-017`: Task detail view (web equivalent)

## Future Enhancements

- Add `--edit` flag to open task file in editor after showing
- Add `--open` flag to open task file in default markdown viewer
- Show task history/changelog if tracked
- Show related tasks (similar tags, same directory)
- Add `--preview` mode with truncated description
- Support showing multiple tasks at once
