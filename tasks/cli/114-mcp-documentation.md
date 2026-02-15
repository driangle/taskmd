---
id: "114"
title: "MCP server: documentation and claude-code-plugin config"
status: completed
priority: medium
effort: small
tags:
  - mcp
  - mvp
  - docs
touches:
  - cli
parent: "107"
dependencies:
  - "111"
  - "112"
  - "113"
created: 2026-02-15
---

# MCP Server: Documentation and Claude Code Plugin Config

## Objective

Add documentation for configuring the taskmd MCP server with common clients and update the claude-code-plugin with MCP configuration.

## Tasks

- [x] Add MCP server documentation to `docs/` covering setup and usage
- [x] Include example configuration snippets for Claude Desktop, Cursor, and other MCP clients
- [x] Update the claude-code-plugin to include MCP client configuration snippet
- [x] Document available tools with their input/output schemas

## Acceptance Criteria

- Documentation explains how to start the MCP server (`taskmd mcp`)
- Configuration snippets work for Claude Desktop and Cursor
- claude-code-plugin contains MCP configuration for connecting to the local server
- Tool reference documents all available tools with parameters and examples
