---
id: "057"
title: "Rename show command to get"
status: completed
priority: low
effort: small
dependencies: ["037"]
tags:
  - cli
  - go
  - refactor
  - commands
  - mvp
created: 2026-02-12
---

# Rename Show Command to Get

## Objective

Rename the `show` command to `get` to follow more common CLI conventions (similar to kubectl, git, etc.).

## Tasks

- [ ] Rename `internal/cli/show.go` to `internal/cli/get.go` (if separate file exists)
- [ ] Update command definition from `show` to `get`
- [ ] Update all command use/short/long descriptions
- [ ] Update help text and examples
- [ ] Update any references in other files
- [ ] Update or rename test file from `show_test.go` to `get_test.go`
- [ ] Update all test function names and test cases
- [ ] Update documentation and README with new command name
- [ ] Consider adding `show` as a deprecated alias that warns users

## Acceptance Criteria

- `taskmd get <task-id>` displays task details
- `taskmd show <task-id>` either doesn't exist or shows deprecation warning
- `taskmd get --help` shows correct command documentation
- All tests pass with new command name
- Documentation reflects the new `get` command name
- No broken references remain in codebase

## Implementation Notes

The `get` command name is more consistent with common CLI tools:
- `kubectl get pods`
- `git show` vs `git log` (show is for commits, not retrieval)
- `docker ps` vs `docker inspect` (inspect is like get)
- `gh pr view` (view/get are similar)

Common CLI patterns use:
- `get` - retrieve and display a resource
- `list` - list multiple resources
- `show` - less common, sometimes used for formatted display

## Migration Options

1. **Hard rename**: Remove `show` entirely (breaking change)
2. **Alias with warning**: Keep `show` as hidden alias with deprecation message
3. **Both commands**: Support both temporarily

Recommend option 2: Keep `show` as a hidden deprecated alias for one version.

## Examples

```bash
# New command
taskmd get 042

# With format options
taskmd get 042 --format json
taskmd get 042 --format yaml

# Old command (if aliased)
taskmd show 042
# Warning: 'show' is deprecated, use 'get' instead
```
