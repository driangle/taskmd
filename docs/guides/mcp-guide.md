# MCP Server Guide

Use the taskmd MCP server to give LLM-based tools direct access to your tasks. Any client that supports the [Model Context Protocol](https://modelcontextprotocol.io) (Claude Code, Claude Desktop, Cursor, Windsurf, etc.) can list, query, update, and analyze tasks without running CLI commands.

## Starting the Server

```bash
taskmd mcp
```

This starts an MCP server over stdio. The server exposes all task operations as tools that MCP clients can discover and call.

## Client Configuration

### Claude Code

If you have the [taskmd plugin](../claude-code-plugin/) installed, the MCP server is bundled and starts automatically — no extra configuration needed.

Otherwise, add to your project's `.mcp.json` or run `claude mcp add --transport stdio taskmd -- taskmd mcp`:

```json
{
  "mcpServers": {
    "taskmd": {
      "command": "taskmd",
      "args": ["mcp"]
    }
  }
}
```

### Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "taskmd": {
      "command": "taskmd",
      "args": ["mcp"]
    }
  }
}
```

### Cursor

Add to your project's `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "taskmd": {
      "command": "taskmd",
      "args": ["mcp"]
    }
  }
}
```

### Windsurf

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "taskmd": {
      "command": "taskmd",
      "args": ["mcp"]
    }
  }
}
```

### Pointing to a Specific Task Directory

If your tasks live outside the current directory, pass the `task_dir` parameter to individual tool calls, or start the server from the project root where your `.taskmd.yaml` is located.

## Available Tools

The MCP server exposes 8 tools. All tools accept an optional `task_dir` parameter (defaults to the current directory).

---

### list

List and filter tasks in a taskmd project.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_dir` | string | no | Directory to scan (default: `.`) |
| `filters` | string[] | no | Filter expressions, e.g. `["status=pending", "priority=high"]` |
| `sort` | string | no | Sort field: `id`, `title`, `status`, `priority`, `effort`, `created` |

**Returns:** JSON array of task objects.

**Example:**
```json
{
  "filters": ["status=pending", "priority=high"],
  "sort": "priority"
}
```

---

### get

Get full details of a single task by ID, including body content and dependency information.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_dir` | string | no | Directory to scan (default: `.`) |
| `task_id` | string | **yes** | Task ID to retrieve |

**Returns:** JSON object with task metadata, full markdown `content`, `depends_on` (upstream dependencies with titles), `blocks` (downstream dependents), and `children` (subtasks).

**Example:**
```json
{
  "task_id": "042"
}
```

---

### next

Get ranked task recommendations based on priority, dependencies, and critical path analysis.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_dir` | string | no | Directory to scan (default: `.`) |
| `limit` | integer | no | Max recommendations (default: 5) |
| `filters` | string[] | no | Filter expressions, e.g. `["priority=high", "tag=mvp"]` |
| `quick_wins` | boolean | no | Only show small-effort tasks |
| `critical` | boolean | no | Only show tasks on the critical path |

**Returns:** JSON array of ranked task recommendations with scores.

**Example:**
```json
{
  "limit": 3,
  "quick_wins": true
}
```

---

### search

Full-text search across task titles and bodies, returning matches with snippets.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_dir` | string | no | Directory to scan (default: `.`) |
| `query` | string | **yes** | Search query |

**Returns:** JSON array of matching tasks with search result snippets.

**Example:**
```json
{
  "query": "authentication"
}
```

---

### context

Resolve relevant file paths for a task based on its `touches` scopes and explicit context fields.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_dir` | string | no | Directory to scan (default: `.`) |
| `task_id` | string | **yes** | Task ID to resolve context for |
| `scopes` | object | no | Scope definitions mapping scope names to file path arrays |
| `project_root` | string | no | Project root for resolving paths (default: `.`) |
| `resolve` | boolean | no | Expand directory paths to individual files |
| `include_content` | boolean | no | Inline file contents and task body |
| `max_files` | integer | no | Cap number of files returned (0 = unlimited) |

**Returns:** JSON object with resolved file paths and optional content.

**Example:**
```json
{
  "task_id": "042",
  "resolve": true,
  "include_content": true
}
```

---

### set

Update fields on a task (status, priority, effort, owner, tags).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_dir` | string | no | Directory to scan (default: `.`) |
| `task_id` | string | **yes** | Task ID to update |
| `status` | string | no | `pending`, `in-progress`, `completed`, `blocked`, `cancelled` |
| `priority` | string | no | `low`, `medium`, `high`, `critical` |
| `effort` | string | no | `small`, `medium`, `large` |
| `owner` | string | no | Owner/assignee |
| `tags` | string[] | no | Replace all tags |
| `add_tags` | string[] | no | Tags to add |
| `rem_tags` | string[] | no | Tags to remove |

**Returns:** JSON object with `task_id`, `file_path`, and a map of `updated` fields.

**Example:**
```json
{
  "task_id": "042",
  "status": "in-progress",
  "add_tags": ["sprint-3"]
}
```

---

### validate

Validate task files for correctness, checking required fields, enum values, dependencies, and cycles.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_dir` | string | no | Directory to scan (default: `.`) |
| `strict` | boolean | no | Enable strict mode for additional warnings |

**Returns:** JSON object with `valid` (boolean), `errors` and `warnings` counts, `task_count`, and an `issues` array with `level`, `task_id`, `file_path`, and `message`.

**Example:**
```json
{
  "strict": true
}
```

---

### graph

Get the task dependency graph as JSON with nodes, edges, and cycle detection.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_dir` | string | no | Directory to scan (default: `.`) |
| `root_task_id` | string | no | Focus on a specific task and its upstream/downstream dependencies |
| `exclude_status` | string[] | no | Exclude tasks with these statuses |
| `filters` | string[] | no | Filter expressions, e.g. `["status=pending"]` |

**Returns:** JSON graph object with nodes (tasks), edges (dependency relationships), and cycle detection information.

**Example:**
```json
{
  "exclude_status": ["completed", "cancelled"],
  "root_task_id": "042"
}
```

## Troubleshooting

**"taskmd: command not found"**
The MCP client needs `taskmd` in its PATH. Install with Homebrew (`brew install driangle/tap/taskmd`) or `go install`.

**Tools return empty results**
Make sure the MCP server is started from your project root (where `tasks/` lives), or pass `task_dir` to each tool call.

**Client doesn't see the tools**
Verify your MCP configuration file is in the right location and restart the client. Check the client's MCP logs for connection errors.

## Learn More

- [CLI User Guide](cli-guide.md) - Full CLI command reference
- [Task Specification](../taskmd_specification.md) - Task file format
- [Quick Start](quickstart.md) - Getting started with taskmd
