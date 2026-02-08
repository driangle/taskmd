---
id: "030"
title: "Static lint enforcement - Code size limits"
status: pending
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

- [ ] Research Go linting tools that support line count limits (revive, golangci-lint, etc.)
- [ ] Configure linter with custom rules:
  - `max-lines-per-file: 200`
  - `max-lines-per-function: 60`
- [ ] Add linter configuration file (`.golangci.yml` or `revive.toml`)
- [ ] Add `lint` target to Makefile
- [ ] Document linting rules in project README or CONTRIBUTING.md
- [ ] Run linter on existing codebase and identify violations
- [ ] Refactor any existing violations to comply with limits
- [ ] Add linter to CI pipeline (GitHub Actions or similar)
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

## Potential Issues

- **snapshot.go violation**: The newly created snapshot.go file is 483 lines, which exceeds the 200-line limit. This will need refactoring into multiple files:
  - `snapshot.go` - Command definition and main logic
  - `snapshot_output.go` - Output formatters (JSON, YAML, MD)
  - `snapshot_analysis.go` - Derived field calculations

## Related Tasks

- Consider adding other quality metrics (cyclomatic complexity, test coverage)
- Add automated code formatting enforcement (gofmt, goimports)
- Implement code complexity limits
