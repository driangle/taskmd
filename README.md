# taskmd

[![CI](https://github.com/driangle/taskmd/actions/workflows/ci.yml/badge.svg)](https://github.com/driangle/taskmd/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/driangle/taskmd/branch/main/graph/badge.svg)](https://codecov.io/gh/driangle/taskmd)

> Markdown-based task management designed for both humans and AI coding assistants.

taskmd provides:

1. **[A standardized format for task files](https://driangle.github.io/taskmd/reference/specification)** — Tasks stored as readable `.md` files with YAML frontmatter, version-controlled alongside your code
2. **[A CLI for managing them](https://driangle.github.io/taskmd/guide/cli)** — Create, update, filter, validate, and visualize tasks from the terminal
3. **[A web dashboard for visualization](https://driangle.github.io/taskmd/guide/web)** — Kanban board, dependency graphs, and project metrics in your browser
4. **[A set of AI skills](https://driangle.github.io/taskmd/guide/claude-code-plugin)** — Slash commands for AI assistants to create, update, and work through tasks
5. **[An MCP server exposing task management tools](https://driangle.github.io/taskmd/guide/mcp)** — Direct tool access for AI assistants via the Model Context Protocol

Everything runs locally. Task data stays in your repo and is never shared externally.

## Quick Start

### Installation

**Option 1: Homebrew (macOS and Linux)**
```bash
# Add the tap
brew tap driangle/tap

# Install taskmd
brew install taskmd

# Verify installation
taskmd --version
```

**Option 2: Download Pre-built Binary**
```bash
# Download from the releases page
# https://github.com/driangle/taskmd/releases

# Extract the archive
tar -xzf taskmd-v*.tar.gz  # or unzip for Windows

# Move to PATH
sudo mv taskmd /usr/local/bin/  # macOS/Linux
```

**Option 3: Install with Go**
```bash
go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest
```

**Option 4: Build from Source**
```bash
git clone https://github.com/driangle/taskmd.git
cd taskmd/apps/cli
make build-full
```

**Option 5: Docker**
```bash
# Web dashboard (default)
docker run --rm -p 8080:8080 -v ./tasks:/tasks:ro ghcr.io/driangle/taskmd

# CLI commands
docker run --rm -v ./tasks:/tasks ghcr.io/driangle/taskmd taskmd list
```

### 30-Second Setup

1. **Initialize taskmd in your project**:
   ```bash
   cd my-project
   taskmd init  # Creates tasks/ directory and .taskmd.yaml config
   ```

2. **Create your first task** (`tasks/001-first-task.md`):
   ```markdown
   ---
   id: "001"
   title: "My first task"
   status: pending
   priority: high
   ---

   # My First Task

   ## Objective
   This is my first task using taskmd!

   ## Tasks
   - [ ] Learn taskmd basics
   - [ ] Create more tasks
   ```

3. **List your tasks**:
   ```bash
   taskmd list tasks/
   ```

4. **Launch the web interface**:
   ```bash
   taskmd web start --open
   ```

That's it! You're ready to manage tasks with taskmd.

## Usage

### CLI Commands

```bash
# List tasks
taskmd list tasks/

# Validate task files
taskmd validate tasks/

# View task statistics
taskmd stats tasks/

# Find next task to work on
taskmd next tasks/

# Visualize dependencies
taskmd graph tasks/ --format ascii

# Start web interface
taskmd web start --dir tasks/ --open
```

### Web Interface

Start the web server and open your browser:

```bash
taskmd web start --open --port 8080
```

The web interface provides:
- **Task List**: Sortable, filterable table view
- **Board View**: Kanban-style board with drag-and-drop
- **Graph View**: Interactive dependency visualization
- **Statistics**: Project metrics and progress tracking

## Documentation

**[Read the full documentation →](https://driangle.github.io/taskmd/)**

- **[Quick Start Guide](https://driangle.github.io/taskmd/getting-started/)** - Get productive in 5 minutes
- **[CLI Guide](https://driangle.github.io/taskmd/guide/cli)** - Comprehensive CLI reference
- **[Web Interface](https://driangle.github.io/taskmd/guide/web)** - Web dashboard walkthrough
- **[Task Specification](https://driangle.github.io/taskmd/reference/specification)** - Task file format reference
- **[FAQ](https://driangle.github.io/taskmd/faq)** - Frequently asked questions

## Task Format

Tasks are markdown files with YAML frontmatter:

```markdown
---
id: "001"
title: "Implement feature X"
status: pending
priority: high
effort: medium
dependencies: []
tags:
  - feature
  - backend
created: 2026-02-08
---

# Implement Feature X

## Objective
Build the new feature X that allows users to...

## Tasks
- [ ] Design API endpoints
- [ ] Implement backend logic
- [ ] Write tests
- [ ] Update documentation

## Acceptance Criteria
- All tests pass
- API documentation complete
- Performance meets requirements
```

See the [Task Specification](docs/taskmd_specification.md) for complete format details.

## Configuration

taskmd supports `.taskmd.yaml` configuration files for setting default options:

```yaml
# .taskmd.yaml - Place in project root or home directory
dir: ./tasks                    # Default task directory
web:
  port: 8080                   # Default web server port
  auto_open_browser: true      # Auto-open browser on web start
```

**Config file locations** (in order of precedence):
1. `./.taskmd.yaml` - Project-specific settings
2. `~/.taskmd.yaml` - User-wide defaults
3. Command-line flags always override config values

See [docs/.taskmd.yaml.example](docs/.taskmd.yaml.example) for a complete example with all supported options.

## Project Structure

```
my-project/
├── tasks/              # Task files
│   ├── 001-task.md
│   ├── 002-task.md
│   └── cli/           # Optional subdirectories
│       └── 003-task.md
└── .taskmd.yaml       # Optional project config
```

## Contributing

Contributions are welcome! For development guidelines, see:

- **[CLAUDE.md](CLAUDE.md)** - Development guidelines and testing requirements
- **[Task Specification](docs/taskmd_specification.md)** - Task format conventions

### Development Setup

```bash
# Clone repository
git clone https://github.com/driangle/taskmd.git
cd taskmd

# Build CLI (from apps/cli directory)
cd apps/cli
make build

# Run tests
make test

# Run linter
make lint

# Build with embedded web UI
make build-full
```

### Running Tests

```bash
cd apps/cli
go test ./...
```

All new CLI features must include comprehensive tests. See [CLAUDE.md](CLAUDE.md) for testing requirements.

### Code Coverage

Code coverage is tracked automatically via [Codecov](https://codecov.io/gh/driangle/taskmd). On every push and pull request, the CI generates coverage reports for both the CLI and web app and uploads them to Codecov. The coverage badge at the top of this README reflects the latest coverage on `main`.

To generate coverage reports locally:

```bash
# CLI (Go)
cd apps/cli
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out    # Open in browser

# Web (TypeScript/React)
cd apps/web
pnpm test:coverage                  # Generates coverage/index.html
open coverage/index.html            # Open in browser
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/driangle/taskmd/issues)
- **Documentation**: [driangle.github.io/taskmd](https://driangle.github.io/taskmd/)
- **Specification**: [taskmd_specification.md](docs/taskmd_specification.md)

## Claude Code Plugin

Use taskmd directly inside [Claude Code](https://claude.com/claude-code) with slash commands:

```
/taskmd:next-task              # Find next task to work on
/taskmd:do-task 015            # Pick up and work on a task
/taskmd:list-tasks --status pending  # List pending tasks
/taskmd:add-task Fix login bug       # Create a new task
/taskmd:complete-task 015      # Mark a task done
/taskmd:validate-tasks         # Validate task files
```

Two plugins are available — **taskmd** provides slash command skills for interactive workflows, and **taskmd-mcp** provides an MCP server for direct tool access. You can install either or both.

```bash
# Add the taskmd marketplace
claude plugin marketplace add driangle/taskmd

# Install slash command skills (/taskmd:do-task, /taskmd:next-task, etc.)
claude plugin install taskmd@taskmd-marketplace --scope project

# Optional: install the MCP server for direct tool access
claude plugin install taskmd-mcp@taskmd-marketplace --scope project
```

See [`claude-code-plugin/README.md`](claude-code-plugin/README.md) for full details.

---

**Built with ❤️ for developers who love markdown**
