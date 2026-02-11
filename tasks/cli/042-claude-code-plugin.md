---
id: "cli-042"
title: "Add Claude Code Plugin with taskmd skill"
status: pending
priority: low
effort: medium
tags:
  - cli
  - integration
  - claude-code
  - dx
  - plugin
  - post-mvp
created: 2026-02-08
---

# Add Claude Code Plugin with Taskmd Skill

## Objective

Create a Claude Code Plugin that provides a `taskmd` skill, allowing users to easily install and use taskmd CLI functionality directly within Claude Code sessions.

## Context

Claude Code supports plugins that can provide custom skills for enhanced workflows. By packaging taskmd as a Claude Code Plugin, users can:

- Install taskmd functionality with `claude code plugins install taskmd`
- Use `/taskmd` slash commands within Claude Code sessions
- Get contextual help and task management capabilities integrated into their Claude workflow
- Avoid manual CLI setup and have Claude automatically understand taskmd conventions

This will significantly improve the DX for users who want to use taskmd with Claude Code.

## Research Phase

- [ ] Review Claude Code plugin documentation and examples
  - Check https://docs.anthropic.com/claude/docs/claude-code-plugins
  - Look at example plugins in the Claude Code ecosystem
- [ ] Understand plugin manifest format (likely JSON/YAML)
- [ ] Determine skill definition format and requirements
- [ ] Identify how skills invoke CLI commands
- [ ] Check if plugins can include binaries or if they reference installed tools

## Implementation Tasks

### Plugin Structure

- [ ] Create `claude-code-plugin/` directory at project root
- [ ] Create plugin manifest file (e.g., `plugin.json` or `manifest.yaml`)
  - Define plugin metadata (name, version, description, author)
  - Declare the `taskmd` skill
  - Specify dependencies (Go binary, or reference to installed CLI)
- [ ] Document plugin file structure

### Skill Definition

- [ ] Define `taskmd` skill with common operations:
  - List tasks in current directory
  - Show next available task
  - Validate task files
  - Show task statistics
  - Visualize dependency graph
- [ ] Create skill prompt/instructions file
  - Explain taskmd format and conventions
  - Provide examples of common usage patterns
  - Include CLI command reference
- [ ] Add parameter handling for:
  - Directory path (default to current)
  - Format options (json, yaml, table)
  - Filter options (status, priority, tags)

### Integration

- [ ] Configure skill to invoke taskmd CLI commands
- [ ] Handle output formatting for Claude context
- [ ] Add error handling and user feedback
- [ ] Test skill invocation from Claude Code

### Documentation

- [ ] Create `claude-code-plugin/README.md` with:
  - Installation instructions
  - Usage examples for the `taskmd` skill
  - Troubleshooting guide
- [ ] Update main project README with plugin information
- [ ] Add plugin to any relevant package registries

### Testing

- [ ] Test plugin installation process
- [ ] Verify skill invocation works correctly
- [ ] Test with various directory structures
- [ ] Ensure error messages are helpful
- [ ] Test integration with existing task workflows

## Acceptance Criteria

- Claude Code Plugin manifest exists with proper metadata
- `taskmd` skill is defined with clear capabilities
- Skill can be invoked with `/taskmd` in Claude Code
- Plugin can access and execute taskmd CLI commands
- Documentation covers installation and usage
- Plugin follows Claude Code plugin best practices
- Users can install with a single command
- Skill provides contextual help within Claude sessions

## Example Skill Usage

```
User: /taskmd list --status pending
Claude: [Invokes taskmd CLI and shows pending tasks]

User: /taskmd next
Claude: [Shows next recommended task with context]

User: /taskmd validate
Claude: [Validates task files and reports issues]
```

## Dependencies

- Requires taskmd CLI to be available (either bundled or installed)
- May need to update CLI for better machine-readable output
- Depends on Claude Code plugin system capabilities

## References

- Claude Code Plugin Documentation
- Existing CLAUDE.md template (task 036)
- Taskmd CLI commands and output formats

## Notes

- Consider whether plugin should bundle the CLI binary or require separate installation
- Skill should gracefully handle cases where CLI is not installed
- Think about cross-platform compatibility (macOS, Linux, Windows)
- Could include a "setup" subcommand to verify/install taskmd CLI
