---
id: "cli-001"
title: "Fix linting and code quality tooling"
status: pending
priority: high
effort: small
tags:
  - infrastructure
  - code-quality
  - tooling
created: 2026-02-08
---

# Fix Linting and Code Quality Tooling

## Current Issues

The linting process fails consistently due to several issues:
1. `golangci-lint` is not installed on the development machine
2. No installation instructions in documentation
3. Missing `make tidy` target in Makefile (user mentioned this fails)
4. `.golangci.yml` configuration exists but cannot be used without the tool

Running `make lint` results in:
```
(eval):1: command not found: golangci-lint
```

## Objective

Fix the linting setup so developers can run code quality checks consistently. Ensure all documented commands in CLAUDE.md work as expected.

## Tasks

### Documentation

- [ ] Add golangci-lint installation instructions to CLAUDE.md
  - macOS: `brew install golangci-lint`
  - Linux: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin`
  - Or: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- [ ] Add section on "Prerequisites" or "Developer Setup" near the top of CLAUDE.md

### Makefile

- [ ] Add `tidy` target to Makefile:
  ```makefile
  # Run go mod tidy to clean up dependencies
  tidy:
      go mod tidy
  ```
- [ ] Consider adding a `check` or `verify` target that runs multiple checks:
  ```makefile
  # Run all checks (test, lint, vet)
  check: test lint
      go vet ./...
  ```

### Linting Configuration

- [ ] Verify `.golangci.yml` configuration is correct
- [ ] Run `golangci-lint run` after installation to check for existing issues
- [ ] Document any intentional lint suppressions needed
- [ ] Consider adding a `.golangci.yml` explanation comment if rules are complex

### CI/CD Consideration

- [ ] If using GitHub Actions or similar CI, ensure golangci-lint runs there
- [ ] Add badge to README.md showing lint status (optional)

## Acceptance Criteria

- `make lint` runs successfully after developer installs golangci-lint
- `make lint-fix` runs successfully and auto-fixes simple issues
- `make tidy` runs `go mod tidy` successfully
- CLAUDE.md includes clear installation instructions
- All existing code passes linting (or has documented exceptions)
- `go vet` and `go test` continue to work as before

## Notes

- Current `.golangci.yml` enforces:
  - Max 60 lines per function
  - Max cyclomatic complexity of 15
  - Max cognitive complexity of 20
  - Standard Go formatting, imports, error checking
- Test files are excluded from function length checks
- The configuration is well-structured and should work once golangci-lint is installed

## Related

- CLAUDE.md lines 80-116 document the linting standards but not installation
- Makefile has `lint` and `lint-fix` targets but no `tidy` target
