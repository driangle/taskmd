---
id: "111"
title: "MCP server: scaffolding and list tool"
status: completed
priority: high
effort: medium
tags:
  - mcp
  - mvp
touches:
  - cli
parent: "107"
created: 2026-02-15
---

# MCP Server: Scaffolding and List Tool

## Objective

Set up the MCP server scaffolding and implement the first tool (`list`) as a proof of concept. This establishes the foundation that all subsequent MCP tools will build on.

## Tasks

- [x] Add `github.com/modelcontextprotocol/go-sdk` dependency to go.mod
- [x] Create `apps/cli/internal/mcp/` package directory
- [x] Implement `server.go` — creates `mcp.Server` with taskmd implementation info
- [x] Implement `list` tool handler — wraps scanner + filter packages to list tasks
- [x] Add `taskmd mcp` subcommand in `apps/cli/internal/cli/mcp.go` that starts the MCP server over stdio
- [x] Support `--task-dir` flag (already a global flag) for pointing at the correct task directory
- [x] Write tests for the list tool handler using in-memory transport

## Acceptance Criteria

- `taskmd mcp` starts an MCP server over stdio
- The `list` tool is discoverable and callable by MCP clients
- The list tool supports filter parameters (status, tags, group, etc.)
- Tests cover happy path, filtering, and error cases
- Code follows existing project conventions (function length, error handling, etc.)
