# CLI Tasks Overview

This directory contains tasks for implementing the `taskmd` CLI tool - a comprehensive command-line interface for managing markdown-based tasks.

## Architecture

The CLI uses a subcommand-based architecture with the following structure:

```
taskmd <command> [input] [flags]
```

- Input defaults to `tasks.md` if omitted
- All commands support `--stdin`, `--format`, `--quiet`, `--verbose`, `--config`

## Task Structure

### ✅ Completed Tasks

- **015** - Go CLI scaffolding (taskmd)
- **016** - Task model & markdown parsing

### Foundation (Priority: High)

- **017** - CLI framework setup (cobra) - Sets up subcommand architecture and shared flags
- **018** - `list` command - Quick textual list view
- **019** - `validate` command - Lint and validate tasks

### Core Commands (Priority: High/Medium)

- **020** - `stats` command - Show computed metrics
- **021** - `snapshot` command - Static machine-readable output
- **022** - `graph` command - Export dependency graph
- **023** - `board` command - Grouped kanban-like view
- **024** - `report` command - Comprehensive report generation

### Advanced Features (Priority: Medium/Low)

- **025** - `export` command - Multi-artifact export
- **026** - `tui` command - Interactive terminal UI
- **027** - File watcher & live reload (for TUI)
- **028** - Directory scanner (multi-file support for TUI)

### Polish (Priority: Low)

- **029** - CLI polish & error handling

## Commands Overview

| Command | Purpose | Output Formats | Priority |
|---------|---------|----------------|----------|
| `list` | Quick textual list | table, json, yaml | High |
| `validate` | Lint and validate | text, json | High |
| `stats` | Show metrics | table, json | Medium |
| `snapshot` | Frozen state export | json, yaml, md | High |
| `graph` | Dependency graph | dot, mermaid, ascii, json | High |
| `board` | Kanban-like grouping | md, txt, json | Medium |
| `report` | Comprehensive report | md, html, json | Medium |
| `export` | Multi-artifact output | multiple | Low |
| `tui` | Interactive UI | interactive | High |

## Implementation Order

1. **Foundation** (017-019): CLI framework, list, validate
2. **Analysis** (020-022): stats, snapshot, graph
3. **Reporting** (023-024): board, report
4. **Advanced** (025-028): export, tui, supporting features
5. **Polish** (029): Error handling, completions, UX

## Dependencies

```
015 (scaffolding) ✅
  └─ 016 (parsing) ✅
       ├─ 017 (framework)
       │    ├─ 018 (list)
       │    ├─ 019 (validate)
       │    ├─ 020 (stats)
       │    ├─ 023 (board)
       │    ├─ 027 (watcher)
       │    └─ 028 (scanner)
       │         └─ 026 (tui)
       └─ 019 (validate)
            └─ 021 (snapshot)
                 └─ 022 (graph)
                      └─ 024 (report)
                           └─ 025 (export)
```

## Testing Strategy

Each command should include:
- Unit tests for core logic
- Integration tests for CLI interface
- Test fixtures with sample task files
- Edge case testing (empty files, malformed data, etc.)

## Common Patterns

### Input Resolution
1. Check for `--stdin` flag → read from stdin
2. Check for explicit file argument → use that file
3. Default to `tasks.md` in current directory

### Output Handling
1. Check for `--out` flag → write to file
2. Default to stdout
3. Respect `--quiet` and `--verbose` flags

### Error Handling
- Use appropriate exit codes (0 = success, 1 = error, 2 = validation warning)
- Provide clear, actionable error messages
- Handle edge cases gracefully (missing files, malformed data, etc.)
