---
id: "058"
title: "Rename update command to set"
status: pending
priority: low
effort: small
dependencies: ["049"]
tags:
  - cli
  - go
  - refactor
  - commands
created: 2026-02-12
---

# Rename Update Command to Set

## Objective

Rename the `update` command to `set` to follow common CLI conventions and provide clearer semantics for modifying task properties.

## Tasks

- [ ] Rename `internal/cli/update.go` to `internal/cli/set.go` (if separate file exists)
- [ ] Update command definition from `update` to `set`
- [ ] Update all command use/short/long descriptions
- [ ] Update help text and examples
- [ ] Update any references in other files
- [ ] Update or rename test file from `update_test.go` to `set_test.go`
- [ ] Update all test function names and test cases
- [ ] Update documentation and README with new command name
- [ ] Consider adding `update` as a deprecated alias that warns users

## Acceptance Criteria

- `taskmd set <task-id> --status completed` updates task status
- `taskmd set <task-id> --tags tag1,tag2` updates task tags
- `taskmd update` either doesn't exist or shows deprecation warning
- `taskmd set --help` shows correct command documentation
- All tests pass with new command name
- Documentation reflects the new `set` command name
- No broken references remain in codebase

## Implementation Notes

The `set` command name is more consistent with common CLI tools:
- `kubectl set image deployment/nginx nginx=nginx:1.16`
- `git config --set user.name "Name"`
- `docker update` (sets resource limits, but less common)
- `gh pr edit` (edit is similar to set)

Common CLI patterns use:
- `set` - set specific properties or values
- `update` - often implies downloading/syncing from remote
- `edit` - interactive editing (usually opens editor)
- `patch` - partial updates (REST convention)

The `set` command is clearer because:
- It explicitly indicates you're setting a value
- It avoids confusion with "update from remote"
- It's shorter and more direct

## Migration Options

1. **Hard rename**: Remove `update` entirely (breaking change)
2. **Alias with warning**: Keep `update` as hidden alias with deprecation message
3. **Both commands**: Support both temporarily

Recommend option 2: Keep `update` as a hidden deprecated alias for one version.

## Examples

```bash
# Set task status
taskmd set 042 --status completed

# Set multiple properties
taskmd set 042 --status in-progress --priority high

# Set tags
taskmd set 042 --tags cli,bug,urgent

# Old command (if aliased)
taskmd update 042 --status completed
# Warning: 'update' is deprecated, use 'set' instead
```
