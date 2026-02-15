---
id: "112"
title: "MCP server: get, next, search, context tools"
status: pending
priority: high
effort: medium
tags:
  - mcp
  - mvp
touches:
  - cli
parent: "107"
dependencies:
  - "111"
created: 2026-02-15
---

# MCP Server: Get, Next, Search, Context Tools

## Objective

Add read-only MCP tools that wrap existing taskmd query operations: get (single task), next (recommendation), search (full-text), and context (file resolution).

## Tasks

- [ ] Implement `get` tool handler — wraps scanner + parser to return a single task by ID
- [ ] Implement `next` tool handler — wraps the next package for task recommendations
- [ ] Implement `search` tool handler — wraps the search package for full-text search
- [ ] Implement `context` tool handler — wraps the taskcontext package for file resolution
- [ ] Register all tools with the MCP server
- [ ] Write tests for each tool handler

## Acceptance Criteria

- All four tools are discoverable and callable by MCP clients
- `get` returns full task details including body for a given task ID
- `next` returns ranked recommendations with optional filters
- `search` returns matching tasks with snippets
- `context` returns resolved file paths for a task
- Each tool validates its input parameters
- Tests cover happy path and error cases for each tool
