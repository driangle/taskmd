# TUI User Guide

Complete guide to using the taskmd Terminal User Interface (TUI).

## What You'll Learn

- Launching the TUI
- Navigation and keyboard shortcuts
- Search and filtering
- Understanding the interface
- Tips and workflows

## Overview

The taskmd TUI provides an interactive, keyboard-driven interface for browsing and managing tasks directly in your terminal. It's ideal for:

- Quick task browsing without leaving the terminal
- SSH sessions and remote work
- Focused, distraction-free task review
- Keyboard-centric workflows

## Getting Started

### Launching the TUI

**Basic usage:**
```bash
taskmd tui
```

**With options:**
```bash
# Specific directory
taskmd tui ./tasks

# Start with filter
taskmd tui --filter status=pending

# Focus on specific task
taskmd tui --focus 025

# Group by field
taskmd tui --group-by priority

# Read-only mode
taskmd tui --readonly
```

### System Requirements

- Terminal with color support (most modern terminals)
- Works on: Linux, macOS, Windows (WSH, PowerShell, cmd)
- Minimum terminal size: 80x24 characters
- Recommended: 120x40 or larger

## Interface Layout

```
┌────────────────────────────────────────────────┐
│ Header: taskmd  ./tasks                        │ ← Title and directory
├────────────────────────────────────────────────┤
│                                                 │
│ 001    ○ First task [high] (cli, feature)     │ ← Task rows
│ 002    ◐ Second task [medium] (backend)       │   Selected row highlighted
│ 003    ● Completed task                       │
│ ...                                             │
│                                                 │
│ 15 tasks: 8 pending, 3 in-progress, 4 complete │ ← Summary
├────────────────────────────────────────────────┤
│ q: quit  ?: help  /: search  j/k: navigate    │ ← Footer (context help)
└────────────────────────────────────────────────┘
```

### Components

**Header (Top)**
- Shows application name: `taskmd`
- Shows current directory being scanned
- Always visible

**Content (Middle)**
- **Task list**: Scrollable list of tasks
- **Selected row**: Highlighted (different background)
- **Status icons**: Visual indicators
- **Priority badges**: Color-coded
- **Tags**: Shown in parentheses
- **Summary line**: Task counts by status

**Footer (Bottom)**
- Context-sensitive help
- Shows active mode (search, help)
- Displays current keybindings

## Status Icons

The TUI uses icons to indicate task status:

- `○` - **Pending** (yellow) - Not started
- `◐` - **In Progress** (blue) - Currently being worked on
- `●` - **Completed** (green) - Finished
- `✖` - **Blocked** (red) - Cannot proceed

## Priority Colors

Priority levels are color-coded:

- **Critical** - Red `[critical]`
- **High** - Yellow `[high]`
- **Medium** - Blue `[medium]`
- **Low** - Gray `[low]`

## Keyboard Shortcuts

### Essential Shortcuts

**Navigation:**
- `j` or `↓` - Move down one task
- `k` or `↑` - Move up one task
- `g` - Jump to top (first task)
- `G` - Jump to bottom (last task)

**Actions:**
- `/` - Enter search mode
- `Esc` - Clear search / Exit search mode
- `?` - Toggle help display
- `q` or `Ctrl+C` - Quit

### Navigation in Detail

**Moving Up and Down:**
```
j or ↓  : Move selection down one task
k or ↑  : Move selection up one task
```

**Quick Jumps:**
```
g       : Jump to first task (top)
G       : Jump to last task (bottom)
```

The view scrolls automatically to keep the selected task visible.

### Search Mode

**Entering search:**
1. Press `/` to activate search mode
2. Footer changes to: `Search: _`
3. Start typing to filter tasks

**Search behavior:**
- Searches task **ID** and **title**
- Case-insensitive
- Real-time filtering
- Matches partial text

**Search controls:**
```
Type        : Add characters to search
Backspace   : Remove last character
Enter       : Exit search mode (keep filter)
Esc         : Exit search mode and clear filter
```

**Example search flow:**
```bash
# Press / to start search
Search: _

# Type "auth"
Search: auth_

# Shows only tasks with "auth" in ID or title
# - 015  ○ User authentication [high]
# - 028  ○ OAuth integration

# Press Enter to keep filter, or Esc to clear
```

### Help Display

Press `?` to toggle the help panel:

```
Key Bindings:
  q / ctrl+c    Quit
  ?             Toggle help
  j / ↓         Move down
  k / ↑         Move up
  g             Go to top
  G             Go to bottom
  /             Search
  Esc           Clear search
```

Press `?` again to hide help.

## Starting Options

### Focus on Specific Task

Start with a particular task selected:

```bash
taskmd tui --focus 025
```

The TUI opens with task 025 highlighted and scrolled into view. Useful for:
- Continuing work on a specific task
- Reviewing a task mentioned in discussion
- Jumping to a task from another tool

### Pre-apply Filter

Start with a search filter active:

```bash
# Show only pending tasks
taskmd tui --filter status=pending

# Show only high-priority tasks
taskmd tui --filter priority=high

# Show specific tag
taskmd tui --filter tag=cli
```

The filter format is: `field=value`

**Supported filter fields:**
- `status` - pending, in-progress, completed, blocked, cancelled
- `priority` - low, medium, high, critical
- `tag` - any tag value

### Group By Field

Group tasks by a field (future feature):

```bash
taskmd tui --group-by status
taskmd tui --group-by priority
taskmd tui --group-by tag
```

**Note:** Grouping is currently a command-line option but full grouping display is planned for future releases.

### Read-Only Mode

Launch in read-only mode (prevents editing):

```bash
taskmd tui --readonly
```

**Note:** Editing features are planned for future releases. This flag is for forward compatibility.

## Task Display

### Task Row Format

Each task row shows:

```
ID     Icon  Title                    [Priority]  (Tags)
001    ○     Set up authentication    [high]      (backend, security)
│      │     │                        │           │
│      │     │                        │           └─ Tags (comma-separated)
│      │     │                        └─ Priority badge (color-coded)
│      │     └─ Task title (truncated if needed)
│      └─ Status icon (color-coded)
└─ Task ID (fixed width)
```

### Selected Row

The currently selected task is highlighted:
- Different background color (inverse/highlighted)
- Stands out from other rows
- Shows which task keyboard commands affect

### Completed Tasks

Completed tasks are rendered with subtle styling:
- Dimmed text
- Still shows status icon (●) in green
- Distinguishes finished work from active tasks

## Summary Line

The summary line shows aggregate statistics:

```
15 tasks: 8 pending, 3 in-progress, 4 completed
```

When filtering, shows both filtered count and total:

```
Showing 5 of 15 tasks: 3 pending, 2 in-progress, 0 completed
```

If any tasks are blocked or cancelled:

```
15 tasks: 8 pending, 3 in-progress, 4 completed, 2 blocked, 1 cancelled
```

## Common Workflows

### Morning Task Review

```bash
# Launch TUI
taskmd tui

# Review pending tasks
# Press / then type "pending"
Search: pending

# Navigate with j/k
# Press Enter when done searching
```

### Finding High-Priority Work

```bash
# Start with high-priority filter
taskmd tui --filter priority=high

# Or search after launching
taskmd tui
# Press /
Search: high
```

### Quick Task Lookup

```bash
# Jump directly to a task
taskmd tui --focus 025

# Review the task
# Use j/k to see surrounding tasks
# Press q to exit
```

### Reviewing Specific Area

```bash
# Filter by tag
taskmd tui --filter tag=cli

# Or search after launch
taskmd tui
# Press /
Search: cli
```

### Status Check

```bash
# Launch TUI
taskmd tui

# Quickly scan the summary line
# See status distribution
# Navigate to specific tasks as needed
```

### Daily Planning

```bash
# See pending tasks
taskmd tui --filter status=pending

# Navigate with j/k
# Take notes on what to work on
# Press q to exit and start work
```

## Tips and Best Practices

### Efficient Navigation

**Use quick jumps:**
- `g` to jump to top when you're far down
- `G` to jump to bottom to see latest tasks
- `j/k` for fine control

**Search strategically:**
- Search by task ID for exact matches: `025`
- Search by keyword for topic: `auth`, `database`
- Search by tag: `cli`, `backend`

**Keep search active:**
- Press Enter (not Esc) to keep filter
- Continue navigating filtered list
- Press Esc only when you want all tasks back

### Terminal Tips

**Optimal terminal size:**
- Width: 120+ characters (shows full task info)
- Height: 40+ lines (shows more tasks at once)
- Font: Monospace with good Unicode support

**Color schemes:**
- Works best with dark terminals
- Light terminals also supported
- Make sure your terminal supports colors

**SSH sessions:**
```bash
# Set TERM for best results
export TERM=xterm-256color

# Launch TUI
taskmd tui
```

### Keyboard Efficiency

**Vim-style navigation:**
If you use Vim, the `j/k/g/G` keys will feel natural.

**Arrow keys work too:**
Use `↓/↑` if you prefer arrow keys over `j/k`.

**Search workflow:**
1. `/` to start search
2. Type a few characters
3. `Enter` to lock filter
4. Navigate filtered results
5. `Esc` when done to see all tasks

### Integrating with Workflow

**Quick checks:**
```bash
# Add to your shell aliases
alias tt='taskmd tui'
alias ttp='taskmd tui --filter status=pending'
alias tth='taskmd tui --filter priority=high'
```

**Morning routine:**
```bash
#!/bin/bash
# morning-check.sh

echo "Today's tasks:"
taskmd tui --filter status=pending
```

**Task lookup:**
```bash
# Function to jump to task
tt() {
    taskmd tui --focus "$1"
}

# Usage: tt 025
```

## Limitations and Future Features

### Current Limitations

**Read-only:**
- Cannot edit tasks in TUI (yet)
- Use text editor to modify task files
- TUI is for browsing and planning

**No multi-select:**
- Cannot select multiple tasks
- Single-task focus only

**No inline details:**
- Cannot expand task to show full markdown
- Shows summary information only
- Use task detail view in web or CLI for full content

### Planned Features

Future enhancements planned:

- [ ] **Task detail view** - Press Enter to see full task
- [ ] **Inline editing** - Edit task fields in TUI
- [ ] **Status updates** - Change status with keyboard
- [ ] **Multi-column layout** - Show more info at once
- [ ] **Grouping display** - Visual grouping by field
- [ ] **Dependency view** - Show dependencies inline
- [ ] **Color themes** - Configurable color schemes
- [ ] **Custom keybindings** - Rebind keys via config
- [ ] **Mouse support** - Click to select tasks

## Troubleshooting

### Display Issues

**Colors not showing:**
```bash
# Check terminal color support
echo $TERM

# Try setting explicitly
export TERM=xterm-256color
taskmd tui
```

**Layout broken:**
```bash
# Ensure terminal is big enough
# Minimum 80x24, recommended 120x40

# Resize terminal and relaunch
```

**Unicode icons not showing:**
- Ensure your terminal font supports Unicode
- Try a modern monospace font (Fira Code, JetBrains Mono)
- Some minimal terminals may not render icons correctly

### Navigation Issues

**Can't select tasks:**
- Check if any tasks are loaded (look for task count)
- Try searching to verify tasks exist
- Check verbose output: `taskmd tui --verbose`

**Scroll not working:**
- Use `j/k` instead of scrolling
- Terminal scroll may interfere
- Use `g/G` to jump to ends

### Search Issues

**Search not finding tasks:**
- Search is case-insensitive
- Searches only ID and title (not tags yet)
- Try broader search terms

**Can't exit search:**
- Press `Esc` to clear search and exit
- Press `Enter` to keep filter but exit search mode
- Press `q` to quit entirely

### Performance Issues

**Slow with many tasks:**
- Search/filter to reduce displayed tasks
- Consider organizing tasks in subdirectories
- TUI handles 100s of tasks well, 1000s may be slow

**Terminal lag:**
- Check terminal performance (try a different terminal)
- Disable shell prompt enhancements while in TUI
- Close other terminal tabs/windows

## Advanced Usage

### Combining with CLI

Use TUI for browsing, CLI for actions:

```bash
# Find task in TUI
taskmd tui

# Note task ID (e.g., 025)
# Press q to exit

# Use CLI for detailed operations
taskmd graph --root 025 --format ascii
taskmd list --filter tag=related-to-025
```

### Scripting Integration

Launch TUI from scripts:

```bash
#!/bin/bash
# review-pending.sh

# Show pending high-priority tasks
taskmd tui --filter priority=high --filter status=pending

# Continue with other work after TUI exits
echo "Time to get to work!"
```

### Remote Sessions

TUI works great over SSH:

```bash
# SSH to remote machine
ssh user@remote-host

# Launch TUI in project directory
cd ~/projects/my-app
taskmd tui

# Navigate, review, then exit
# Continue working remotely
```

### Tmux/Screen Integration

Run TUI in a persistent terminal session:

```bash
# Start tmux
tmux

# Launch TUI in one pane
taskmd tui

# Split pane (Ctrl+B then %)
# Work in other pane
# Switch back to TUI as needed (Ctrl+B then arrow keys)
```

## Configuration

### Config File (Coming Soon)

Configuration file support is **planned but not yet implemented**. See [task 056](../../tasks/056-implement-taskmd-yaml-config.md).

### Current Options

Use command-line flags:

```bash
# Specify directory
taskmd tui --dir ./tasks

# Or create a shell alias
alias tmtui='taskmd tui --dir ./tasks'
```

### Environment Variables

```bash
# Default directory
export TASKMD_DIR=./tasks

# Terminal settings
export TERM=xterm-256color
```

## Comparison with Other Interfaces

### TUI vs. CLI

**TUI advantages:**
- Interactive browsing
- Visual overview
- Keyboard-driven
- Real-time filtering

**CLI advantages:**
- Scriptable
- Pipe-able output
- Faster for one-off queries
- More output formats

**When to use TUI:**
- Exploring tasks
- Quick reviews
- SSH sessions
- Keyboard-focused work

**When to use CLI:**
- Automation
- Scripting
- CI/CD pipelines
- Generating reports

### TUI vs. Web

**TUI advantages:**
- No server needed
- Terminal-only environment
- Instant startup
- Lower resource usage
- Works over SSH

**Web advantages:**
- Mouse interaction
- Visual graphs
- Multiple views
- Easier for non-technical users
- Board/kanban layout

**When to use TUI:**
- Terminal workflow
- Remote sessions
- Quick checks
- Minimal setup

**When to use Web:**
- Team collaboration
- Visual planning
- Stakeholder demos
- Extended review sessions

## Getting Help

### Built-in Help

Press `?` in the TUI to show keyboard shortcuts.

### Documentation

- **[Quick Start Guide](quickstart.md)** - Get started fast
- **[CLI Guide](cli-guide.md)** - Command-line reference
- **[Web Guide](web-guide.md)** - Web interface guide
- **[Task Specification](../taskmd_specification.md)** - Task format

### Command-Line Help

```bash
taskmd tui --help
```

### Support

- **GitHub Issues**: Report bugs and request features
- **Documentation**: Check other guides for related info

---

**Next:** Explore the [Web interface](web-guide.md) for visual task management, or check the [CLI guide](cli-guide.md) for automation capabilities.
