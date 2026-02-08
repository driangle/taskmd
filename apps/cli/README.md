# taskmd - Markdown Task Tracker CLI

A terminal-based interface for managing markdown task files with automatic file watching and live updates.

## Features

- Scans directories for markdown task files
- Interactive TUI built with Bubble Tea
- Live updates when files change
- Beautiful markdown rendering

## Installation

### Build from source

```bash
make build
```

### Run directly

```bash
make run
```

## Usage

```bash
./taskmd
```

Or from anywhere after installing:

```bash
make install
taskmd
```

## Project Structure

```
apps/cli/
├── cmd/
│   └── taskmd/        # Application entrypoint
│       └── main.go
├── internal/          # Core application logic
├── Makefile          # Build automation
├── go.mod            # Go module definition
└── README.md         # This file
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system watching
- [goldmark](https://github.com/yuin/goldmark) - Markdown parsing

## Development

```bash
# Build
make build

# Run
make run

# Clean
make clean

# Run tests
make test

# Build for multiple platforms
make build-all
```
