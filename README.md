# taskmd

A powerful task management system that stores tasks as markdown files with YAML frontmatter. Manage your projects with the flexibility of plain text files and the power of structured metadata.

## Features

- **📝 Markdown-based**: Tasks stored as readable `.md` files with YAML frontmatter
- **🖥️ Dual Interface**: Use the CLI for automation or the web UI for visual management
- **📊 Dependency Tracking**: Visualize task dependencies with interactive graphs
- **✅ Validation**: Built-in linting ensures task files follow conventions
- **🎯 Smart Filtering**: Find tasks by status, priority, tags, and dependencies
- **📈 Analytics**: Track project progress with statistics and metrics
- **🔄 Live Reload**: Web interface updates automatically when task files change
- **🎨 Multiple Views**: List, board (kanban), graph, and interactive TUI modes

## Quick Start

### Installation

**Option 1: Download Pre-built Binary**
```bash
# Download from releases page (coming soon)
# Extract and add to PATH
```

**Option 2: Install with Go**
```bash
go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest
```

**Option 3: Build from Source**
```bash
git clone https://github.com/driangle/taskmd.git
cd taskmd/apps/cli
make build-full
```

### 30-Second Setup

1. **Create a tasks directory**:
   ```bash
   mkdir -p my-project/tasks
   cd my-project
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

# Launch interactive TUI
taskmd tui tasks/

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

- **[Quick Start Guide](docs/guides/quickstart.md)** - Get productive in 5 minutes
- **[CLI User Guide](docs/guides/cli-guide.md)** - Comprehensive CLI reference
- **[TUI User Guide](docs/guides/tui-guide.md)** - Interactive terminal interface guide
- **[Web User Guide](docs/guides/web-guide.md)** - Web interface walkthrough
- **[Task Format Specification](docs/taskmd_specification.md)** - Task file format reference

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

Create `~/.taskmd.yaml` to set default options:

```yaml
# Default task directory
dir: ./tasks

# Default output format (table, json, yaml)
format: table

# Enable verbose logging by default
verbose: false

# Default web server port
web:
  port: 8080
  open: true
```

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

## License

[License information needed - please add LICENSE file]

## Support

- **Issues**: [GitHub Issues](https://github.com/driangle/taskmd/issues)
- **Documentation**: [docs/guides/](docs/guides/)
- **Specification**: [taskmd_specification.md](docs/taskmd_specification.md)

## Roadmap

- [ ] Homebrew installation (see [tasks/045](tasks/045-publish-homebrew.md))
- [ ] GitHub Pages documentation site (see [tasks/046](tasks/046-documentation-site.md))
- [ ] Task templates and scaffolding
- [ ] Git integration features
- [ ] Team collaboration features

---

**Built with ❤️ for developers who love markdown**
