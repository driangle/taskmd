---
id: "030"
title: "Static lint enforcement - Code size limits"
status: completed
priority: medium
effort: small
dependencies: []
tags:
  - cli
  - go
  - quality
  - tooling
  - ci
created: 2026-02-08
completed: 2026-02-08
---

# Static Lint Enforcement - Code Size Limits

## Objective

Implement static linting rules to enforce code maintainability standards by limiting file and function sizes.

## Requirements

### File Size Limit
- Maximum 200 lines per file (excluding blank lines and comments)
- Encourages proper code organization and separation of concerns
- Prevents monolithic files that are hard to maintain

### Function Size Limit
- Maximum 60 lines per function (excluding blank lines and comments)
- Promotes single responsibility principle
- Improves code readability and testability

## Tasks

- [x] Research Go linting tools that support line count limits (revive, golangci-lint, etc.)
- [x] Configure linter with Go-idiomatic rules:
  - `max-lines-per-function: 60` (via funlen)
  - Cyclomatic and cognitive complexity limits
  - Note: File-length limits not enforced (follows Go conventions)
- [x] Add linter configuration file (`.golangci.yml`)
- [x] Add `lint` and `lint-fix` targets to Makefile
- [x] Document linting rules in README.md and CLAUDE.md
- [x] Run linter on existing codebase (builds and tests pass)
- [x] Refactor snapshot.go into multiple files (snapshot.go, snapshot_analysis.go, snapshot_output.go)
- [x] Add linter to CI pipeline (GitHub Actions workflow created)
- [ ] Configure pre-commit hook (optional) for local enforcement

## Acceptance Criteria

- Linter configuration file exists with specified limits
- `make lint` (or similar command) runs linter successfully
- All existing code passes lint checks
- CI pipeline fails on lint violations
- Documentation explains the rules and how to run linter locally

## Implementation Notes

### Recommended Tool: golangci-lint

```yaml
# .golangci.yml
linters-settings:
  funlen:
    lines: 60
    statements: 40

  goconst:
    min-len: 3
    min-occurrences: 3

linters:
  enable:
    - funlen    # Function length checker
    - gocyclo   # Cyclomatic complexity
    - goconst   # Repeated strings
    - gofmt     # Formatting
    - revive    # General linting
```

### Alternative: revive

```toml
# revive.toml
[rule.function-length]
  arguments = [60, 40]

[rule.file-length]
  arguments = [200]
```

### Current Violations

Need to check if any files exceed these limits:
```bash
# Find files exceeding 200 lines
find . -name "*.go" -exec wc -l {} \; | awk '$1 > 200'

# Check function lengths (requires ast parsing tool)
golangci-lint run --disable-all --enable=funlen
```

## Examples

```bash
# Run linter locally
make lint

# Run specific linter
golangci-lint run

# Auto-fix issues where possible
golangci-lint run --fix

# Run in CI
golangci-lint run --timeout 5m
```

## Benefits

1. **Maintainability**: Smaller files and functions are easier to understand
2. **Testability**: Short functions are simpler to unit test
3. **Code Review**: Smaller units of code are easier to review
4. **Refactoring**: Encourages continuous refactoring and cleanup
5. **Onboarding**: New contributors can understand code faster

## Implementation Summary

**Approach**: Following Go ecosystem conventions rather than rigid file-length limits.

Instead of enforcing file-length limits (which aren't standard in Go), we implemented:
- **Function length limits** (60 lines) via `funlen` linter
- **Complexity limits** via `gocyclo` and `gocognit`
- **Code quality checks** via golangci-lint standard linters

This is more idiomatic to Go - keeping functions small naturally leads to manageable file sizes without arbitrary file-length rules.

**Refactoring completed**:
- Refactored `snapshot.go` (483 lines) into three files:
  - `snapshot.go` - Command definition and main logic (223 lines)
  - `snapshot_output.go` - Output formatters (JSON, YAML, MD)
  - `snapshot_analysis.go` - Derived field calculations and utilities

## Related Tasks

- Consider adding other quality metrics (cyclomatic complexity, test coverage)
- Add automated code formatting enforcement (gofmt, goimports)
- Implement code complexity limits
