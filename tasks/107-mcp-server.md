---
id: "107"
title: "Create MCP server wrapping taskmd commands"
status: completed
priority: medium
effort: large
tags:
  - mcp
  - integration
  - ai
  - mvp
created: 2026-02-14
---

# Create MCP Server Wrapping taskmd Commands

## Objective

Build an MCP (Model Context Protocol) server that exposes taskmd functionality as tools, enabling LLM-based tools beyond Claude Code (e.g., Cursor, Windsurf, Copilot agents, custom agents) to interact with taskmd projects. The server wraps existing CLI commands so that any MCP-compatible client can list, get, create, update, and query tasks.

Should be implemented in Go for maximum reuse of taskmd's internal packages.

## Tasks

- [x] Research MCP server SDK options for Go (or decide on implementation language)
- [x] Define the MCP tool schema for each taskmd operation:
  - `list` — list tasks with optional filters (status, tags, group)
  - `get` — get a single task by ID
  - `set` — update task fields (status, priority, tags, etc.)
  - `next` — get the next recommended task
  - `search` — full-text search across tasks (if task 106 is completed)
  - `graph` — get dependency graph
  - `validate` — validate task files
  - `context` — get task context with relevant files (if task 105 is completed)
- [x] Implement MCP server with stdio transport (standard for local MCP servers)
- [x] Wire each MCP tool to invoke the corresponding taskmd logic (reuse internal packages, not shelling out to the CLI binary)
- [x] Add a `taskmd mcp` subcommand that starts the MCP server
- [x] Support `--task-dir` configuration for pointing at the correct task directory
- [x] Write tests for MCP tool handlers
- [x] Add documentation for configuring the MCP server in common clients (Claude Desktop, Cursor, etc.)
- [x] Update the claude-code-plugin to contain the MCP client configuration snippet for connecting to the local MCP server

## Acceptance Criteria

- `taskmd mcp start` starts an MCP server over stdio
- MCP clients can discover and call taskmd tools (list, get, set, next, etc.)
- Tool inputs/outputs follow MCP specification with proper JSON schemas
- Server reuses existing internal packages rather than shelling out
- Works with Claude Desktop and other MCP-compatible clients
- Documentation includes example MCP client configuration snippets
- Tests cover tool invocation, error handling, and edge cases
