---
id: "029"
title: "CLI polish & error handling"
status: completed
priority: low
effort: medium
dependencies: ["018", "019", "020", "021", "022", "023", "024", "025", "026"]
tags:
  - cli
  - go
  - polish
  - quality
  - mvp
created: 2026-02-08
---

# CLI Polish & Error Handling

## Objective

Polish the CLI with comprehensive error handling, helpful messages, shell completions, and overall UX improvements.

## Tasks

- [ ] Implement consistent error messages across all commands
- [ ] Add colorized output (using lipgloss) with `--no-color` flag
- [ ] Generate shell completions (bash, zsh, fish) using Cobra
- [ ] Add progress indicators for long-running operations
- [ ] Implement `--dry-run` flag where applicable
- [ ] Add examples to all command help text
- [ ] Improve validation error messages with suggestions
- [ ] Add `--debug` flag for troubleshooting
- [ ] Create man pages for commands
- [ ] Test edge cases:
  - Empty input files
  - Malformed task files
  - Missing files
  - Permission errors
  - Invalid flag combinations
- [ ] Add integration tests for all commands
- [ ] Document all exit codes

## Acceptance Criteria

- Error messages are clear and actionable
- Shell completions work for all commands and flags
- `--debug` provides useful troubleshooting info
- Progress indicators show for long operations
- Help text includes practical examples
- All edge cases are handled gracefully
- Exit codes are consistent and documented
- Integration tests pass

## Examples

```bash
taskmd completion bash > /etc/bash_completion.d/taskmd
taskmd list --debug
taskmd validate --dry-run
```

## Notes

This is the final polish task - should be done after all core commands are implemented.
