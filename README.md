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

**Homebrew (macOS and Linux)**
```bash
brew tap driangle/tap
brew install taskmd
taskmd --version
```

**Docker** (CLI and web dashboard, nothing to install)
```bash
# Web dashboard
docker run --rm -p 8080:8080 -v ./tasks:/tasks:ro ghcr.io/driangle/taskmd

# CLI commands
docker run --rm -v ./tasks:/tasks ghcr.io/driangle/taskmd taskmd list
```

**Go**
```bash
go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest
```

> `go install` builds the CLI only; the web dashboard is not included because the
> built frontend is not part of the Go module.

Pre-built binaries for Linux, macOS, and Windows, and building from source, are covered on the [installation page](https://driangle.github.io/taskmd/getting-started/installation).

### 30-Second Setup

1. **Initialize taskmd in your project**:
   ```bash
   cd my-project
   taskmd init  # Creates tasks/, .taskmd.yaml, task templates, the spec, and an agent config file
   ```

2. **Create your first task**:
   ```bash
   taskmd add "My first task" --priority high
   ```

3. **List your tasks**:
   ```bash
   taskmd list
   ```

4. **Launch the web interface**:
   ```bash
   taskmd web start --open
   ```

Tasks are plain markdown files under `tasks/`, so you can also create and edit them in your editor. The [tutorial](https://driangle.github.io/taskmd/getting-started/tutorial) walks through dependencies, recommendations, and validation from here.

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
taskmd web start --task-dir tasks/ --open
```

### Web Interface

```bash
taskmd web start --open
```

- **Tasks**: sortable, filterable table
- **Board**: kanban grouped by status, priority, effort, type, group, tag, or phase, with drag-and-drop
- **Graph**: interactive dependency graph
- **Stats, Next, Tracks, Phases, Feed, Validate**: one page each

The server binds `127.0.0.1` and the API is unauthenticated; before exposing it with `--host 0.0.0.0`, read [Network Access](https://driangle.github.io/taskmd/guide/web#network-access) in the web guide.

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

Three plugins are available. **taskmd** provides slash command skills that drive the CLI. **taskmd-mcp** provides an MCP server so Claude can call task operations as tools, and can be installed alongside **taskmd**. **taskmd-lite** re-implements the skills with Claude's file tools for environments where the binary cannot be installed. See the [plugin guide](https://driangle.github.io/taskmd/guide/claude-code-plugin) for which to pick.

```bash
# Add the taskmd marketplace
claude plugin marketplace add driangle/taskmd

# Install slash command skills (/taskmd:do-task, /taskmd:next-task, etc.)
claude plugin install taskmd@taskmd-marketplace --scope project

# Optional: add the MCP server for direct tool access
claude plugin install taskmd-mcp@taskmd-marketplace --scope project

# Or, with no taskmd binary available:
claude plugin install taskmd-lite@taskmd-marketplace --scope project
```

See [`claude-code-plugin/README.md`](claude-code-plugin/README.md) for full details.

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
created_at: 2026-02-08
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

taskmd loads **one** config file: the nearest `.taskmd.yaml` walking up from the current directory to the repository root, else `~/.taskmd.yaml`. The two are not merged, so a project file means the home file is ignored. Command-line flags and `TASKMD_*` environment variables override whatever the file says.

See [docs/.taskmd.yaml.example](docs/.taskmd.yaml.example) for a complete example with all supported options.

## Task Layout

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

- **[CONTRIBUTING.md](CONTRIBUTING.md)**: what belongs in core, how to extend taskmd, and the PR process
- **[AGENTS.md](AGENTS.md)**: local setup, build and test commands, lint rules, and versioning (`CLAUDE.md` is a symlink to it)
- **[Task Specification](docs/taskmd_specification.md)**: task format conventions

```bash
cd apps/cli
make build        # CLI only; make build-full embeds the web UI
make check        # tests, lint, vet
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/driangle/taskmd/issues)
- **Documentation**: [driangle.github.io/taskmd](https://driangle.github.io/taskmd/)
- **Specification**: [taskmd_specification.md](docs/taskmd_specification.md)
