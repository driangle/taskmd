# taskmd-mcp — MCP Server Plugin

A [Claude Code](https://claude.com/claude-code) plugin that exposes taskmd operations as
MCP tools, so Claude can call them directly instead of shelling out to the CLI. Best for
autonomous task operations; if you prefer human-driven slash commands, install
[`taskmd`](../claude-code-plugin/README.md) instead.

> **Choose one, not both.** `taskmd` and `taskmd-mcp` overlap in functionality —
> installing both clutters your environment with redundant capabilities.

## Prerequisites

The `taskmd` CLI must be installed and on your `PATH`. This plugin runs `taskmd mcp` as
its server (see [`.mcp.json`](./.mcp.json)).

```bash
# Homebrew (macOS and Linux)
brew tap driangle/tap
brew install taskmd

# Or install with Go
go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest

taskmd --version
```

## Installation

```bash
claude plugin marketplace add driangle/taskmd
claude plugin install taskmd-mcp@taskmd-marketplace --scope project
```

## Available Tools

| Tool | Description |
|------|-------------|
| `list` | List and filter tasks in a taskmd project |
| `get` | Get full details of a single task by ID, including body content and dependency information |
| `status` | Get lightweight metadata for a task (no body content, no resolved dependencies) |
| `next` | Get ranked task recommendations based on priority, dependencies, and critical path analysis |
| `search` | Full-text search across task titles and bodies, returning matches with snippets |
| `context` | Resolve relevant file paths for a task based on its touches (scopes) and explicit context fields |
| `set` | Update fields on a task (status, priority, effort, owner, tags) |
| `validate` | Validate task files for correctness, checking required fields, enum values, dependencies, and cycles |
| `graph` | Get the task dependency graph as JSON with nodes, edges, and cycle detection |

## Other MCP Clients

The same server works outside Claude Code. See the
[MCP Server Guide](https://driangle.github.io/taskmd/guide/mcp) for configuration
snippets (Claude Desktop, Cursor, Windsurf, etc.).

## Versioning

This plugin is on its **own `1.x` semver line**, independent of the taskmd CLI release
number and of the other marketplace plugins. See
[ADR 0003](https://github.com/driangle/taskmd/blob/main/docs/adr/0003-plugin-versioning-policy.md).

The tool surface above — tool names, argument schemas, and result shapes — is a
compatibility contract that other MCP clients code against, and it is **stable**:

- **Patch** — bug fixes in an existing tool's behavior.
- **Minor** — a new tool, or a new optional argument on an existing one.
- **Major** — a tool removed or renamed, or an existing argument or result shape changed.

The authoritative version is the `version` field in
[`.claude-plugin/plugin.json`](./.claude-plugin/plugin.json). The marketplace manifest
deliberately does not repeat it.

## Learn More

- [taskmd documentation](https://driangle.github.io/taskmd/)
- [Task file specification](https://github.com/driangle/taskmd/blob/main/docs/taskmd_specification.md)
- [GitHub repository](https://github.com/driangle/taskmd)
