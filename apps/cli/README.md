# taskmd - Markdown Task Tracker CLI

A non-interactive command-line tool (built with [Cobra](https://github.com/spf13/cobra)) for managing tasks stored as markdown files with YAML frontmatter. Tasks live alongside your code in version control; the CLI reads and writes those `.md` files directly.

## Features

- Scans directories for markdown task files
- Subcommands to list, filter, get, add, set, validate, and visualize tasks
- Multiple output formats (table, JSON, YAML, ASCII graphs)
- Dependency graphs, kanban board, and project stats
- Built-in MCP server (`taskmd mcp`) for AI assistants
- Web dashboard export (`taskmd web`)

## Installation

### Build from source

```bash
make build
```

### Install to your PATH

```bash
make install
```

## Usage

Run a subcommand against a directory of task files:

```bash
./taskmd list              # List all tasks
./taskmd next              # Recommend what to work on next
./taskmd add "Task title"  # Create a new task
./taskmd graph --format ascii   # Visualize dependencies
./taskmd validate          # Check task files for errors
```

Run `taskmd --help` to see all available commands, or `taskmd <command> --help` for command-specific flags.

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

- [Cobra](https://github.com/spf13/cobra) - Command-line framework
- [Viper](https://github.com/spf13/viper) - Configuration handling
- [huh](https://github.com/charmbracelet/huh) - Interactive prompts (e.g. `init`)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system watching
- [go-sdk](https://github.com/modelcontextprotocol/go-sdk) - MCP server
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML frontmatter parsing

## Development

```bash
# Build
make build

# Run
make run

# Clean
make clean

# Run unit/integration tests
make test

# Run e2e tests (builds binary, tests full CLI)
make e2e

# Run linter
make lint

# Auto-fix linting issues
make lint-fix

# Build for multiple platforms
make build-all
```

### Code Quality

This project enforces code quality standards using `golangci-lint`:

- **Function length**: Max 60 lines per function
- **Cyclomatic complexity**: Max 15 per function
- **Cognitive complexity**: Max 20 per function
- **Error handling**: All errors must be checked
- **Code formatting**: Enforced via gofmt and goimports

Run `make lint` before committing to ensure your code meets these standards.

**Installation of golangci-lint**:
```bash
# macOS
brew install golangci-lint

# Linux/WSL
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or using Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```
