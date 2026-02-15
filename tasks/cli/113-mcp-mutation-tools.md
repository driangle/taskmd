---
id: "113"
title: "MCP server: set, validate, graph tools"
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

# MCP Server: Set, Validate, Graph Tools

## Objective

Add MCP tools for task mutation (`set`), validation (`validate`), and dependency graph queries (`graph`).

## Tasks

- [ ] Implement `set` tool handler — wraps taskfile package to update task fields (status, priority, tags, etc.)
- [ ] Implement `validate` tool handler — wraps validator package to lint task files
- [ ] Implement `graph` tool handler — wraps graph package for dependency visualization (JSON format)
- [ ] Register all tools with the MCP server
- [ ] Write tests for each tool handler

## Acceptance Criteria

- `set` tool can update status, priority, effort, owner, and tags on a task
- `set` tool validates enum values before applying changes
- `validate` tool returns validation issues with severity levels
- `graph` tool returns dependency graph data in JSON format
- Each tool validates its input parameters
- Tests cover happy path, validation errors, and edge cases
