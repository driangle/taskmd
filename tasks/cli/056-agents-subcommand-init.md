---
id: "056"
title: "Move init command under agents subcommand"
status: pending
priority: medium
effort: medium
dependencies: ["048"]
tags:
  - cli
  - go
  - refactor
  - commands
created: 2026-02-12
---

# Move Init Command Under Agents Subcommand

## Objective

Restructure the CLI command hierarchy to group agent-related commands under an `agents` subcommand. Move the existing `init` command to be invoked as `taskmd agents init` with agent-specific flags.

## Tasks

- [ ] Create `agents` parent command in `internal/cli/agents.go`
- [ ] Move `init` command logic to be a subcommand of `agents`
- [ ] Update init command to use agent-specific flags:
  - `--claude` - Initialize for Claude Code (default if no flags specified)
  - `--gemini` - Initialize for Gemini
  - `--codex` - Initialize for Codex
  - Allow multiple flags to generate configs for multiple agents
  - Default to Claude if no agent flags are provided
- [ ] Update command registration in root command
- [ ] Update help text and command descriptions
- [ ] Update any existing tests for init command
- [ ] Add tests for the new `agents` command structure
- [ ] Update documentation and README with new command syntax
- [ ] Consider adding `agents list` subcommand to show available agents
- [ ] Maintain backward compatibility or add deprecation warning for old `taskmd init`

## Acceptance Criteria

- `taskmd agents init` generates Claude Code configuration (default)
- `taskmd agents init --claude` explicitly generates Claude Code configuration
- `taskmd agents init --gemini` generates Gemini configuration
- `taskmd agents init --codex` generates Codex configuration
- `taskmd agents init --claude --gemini` generates configs for both agents
- `taskmd agents --help` shows available subcommands (init, and potentially others)
- `taskmd agents init --help` shows agent-specific flags and indicates Claude is default
- All tests pass with new command structure
- Documentation reflects new command syntax and default behavior

## Implementation Notes

The `agents` command should be designed as a parent command that can be extended with additional agent-related subcommands in the future, such as:
- `taskmd agents list` - List available agents
- `taskmd agents validate` - Validate agent configurations
- `taskmd agents update` - Update agent configurations

Command structure:
```
taskmd/
  agents/         (parent command)
    init/         (subcommand - current init logic)
    list/         (future: list agents)
    validate/     (future: validate configs)
```

Consider creating a shared `internal/agents/` package for agent-related logic if it doesn't exist.

## Migration Path

Options for handling the old `taskmd init` command:
1. Remove it entirely (breaking change)
2. Keep it with a deprecation warning pointing to `taskmd agents init`
3. Alias it to `taskmd agents init` transparently

Recommend option 2 or 3 for better user experience.

## Examples

```bash
# Initialize for Claude Code (default)
taskmd agents init

# Explicitly initialize for Claude Code
taskmd agents init --claude

# Initialize for Gemini
taskmd agents init --gemini

# Initialize for multiple agents
taskmd agents init --claude --gemini --codex

# Show help for agents commands
taskmd agents --help

# Show help for init subcommand
taskmd agents init --help
```
