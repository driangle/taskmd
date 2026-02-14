# Terminal UI (TUI)

Complete guide to the taskmd interactive terminal interface.

## Overview

The TUI provides a keyboard-driven interface for browsing tasks directly in your terminal. Ideal for:

- Quick task browsing without leaving the terminal
- SSH sessions and remote work
- Focused, distraction-free task review

## Getting Started

```bash
# Launch TUI
taskmd tui

# Specific directory
taskmd tui ./tasks

# Start with filter
taskmd tui --filter status=pending

# Focus on specific task
taskmd tui --focus 025

# Read-only mode
taskmd tui --readonly
```

## Interface Layout

```
┌────────────────────────────────────────────────┐
│ Header: taskmd  ./tasks                        │
├────────────────────────────────────────────────┤
│                                                 │
│ 001    ○ First task [high] (cli, feature)     │
│ 002    ◐ Second task [medium] (backend)       │
│ 003    ● Completed task                       │
│ ...                                             │
│                                                 │
│ 15 tasks: 8 pending, 3 in-progress, 4 complete │
├────────────────────────────────────────────────┤
│ q: quit  ?: help  /: search  j/k: navigate    │
└────────────────────────────────────────────────┘
```

## Status Icons

| Icon | Status | Color |
|------|--------|-------|
| `○` | Pending | Yellow |
| `◐` | In Progress | Blue |
| `●` | Completed | Green |
| `✖` | Blocked | Red |

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `j` or `↓` | Move down one task |
| `k` or `↑` | Move up one task |
| `g` | Jump to first task |
| `G` | Jump to last task |

### Actions

| Key | Action |
|-----|--------|
| `/` | Enter search mode |
| `Esc` | Clear search / exit search mode |
| `?` | Toggle help display |
| `q` or `Ctrl+C` | Quit |

### Search Mode

1. Press `/` to activate search
2. Type to filter tasks by ID and title (case-insensitive)
3. Press `Enter` to keep the filter and exit search mode
4. Press `Esc` to clear the filter and exit search mode

## Common Workflows

### Morning Task Review

```bash
taskmd tui --filter status=pending
# Navigate with j/k, press / to search
```

### Quick Task Lookup

```bash
taskmd tui --focus 025
# Opens with task 025 highlighted
```

### Finding High-Priority Work

```bash
taskmd tui --filter priority=high
```

## Tips

- **Terminal size**: 120x40 or larger recommended for best display
- **Vim users**: `j/k/g/G` navigation will feel natural
- **SSH sessions**: Set `TERM=xterm-256color` for best color support
- **Combine with CLI**: Use TUI for browsing, CLI for actions

### Shell Aliases

```bash
alias tt='taskmd tui'
alias ttp='taskmd tui --filter status=pending'
alias tth='taskmd tui --filter priority=high'
```

## TUI vs. Other Interfaces

| Feature | TUI | CLI | Web |
|---------|-----|-----|-----|
| Interactive browsing | Yes | No | Yes |
| Scriptable | No | Yes | No |
| Works over SSH | Yes | Yes | Via port forwarding |
| Visual graphs | No | ASCII only | Yes (interactive) |
| Server needed | No | No | Yes |
| Startup time | Instant | Instant | 1-2 seconds |
