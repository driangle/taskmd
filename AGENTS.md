# Development Guidelines for taskmd

This document provides guidelines and conventions for developing the taskmd project. These instructions are designed to help maintain code quality, consistency, and reliability across the codebase.

## Project Scope Boundary

Before proposing or building a new feature, integration, or scanner, check it against
the **core scope boundary**: [`docs/adr/0001-core-scope-boundary.md`](docs/adr/0001-core-scope-boundary.md).

In short:

- **Core** = reading, writing, and presenting **markdown task files**: parser, scanner,
  validator, the read views (`list`, `get`, `next`, `board`, `stats`, `graph`, …),
  mutations (`add`, `set`, `rm`, `archive`), config, and the MCP/web surfaces over that
  model. Test: *would it still make sense if taskmd only ever touched the local
  filesystem?*
- **Optional / adjacent** = anything reaching outside that line — a foreign task system,
  a network service, or a different source domain.
  - **`sync`** (Jira/Linear/Trello/GitHub) is being **extracted to a separate module**
    behind a stable task-source API. New foreign-system integrations belong there, not
    in core. (Still in-tree and supported until the extraction lands.)
  - **`todos`** (source-code TODO/FIXME scanner) **stays in core but is capped** — an
    on-ramp for turning comments into tasks, not a growth area. New languages/markers/
    analytics must clear the optional-feature bar.

Record non-obvious scope or architecture decisions as a new ADR under `docs/adr/`.

## Prerequisites

Before developing, ensure you have the following tools installed:

- **Go** (1.25+): [https://go.dev/dl/](https://go.dev/dl/)
- **pnpm**: `npm install -g pnpm` (for web frontend)
- **golangci-lint**: Required for `make lint` and `make lint-fix`
  - macOS: `brew install golangci-lint`
  - go install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
  - Other: See [golangci-lint install docs](https://golangci-lint.run/welcome/install/)
- **Git hooks**: Run `git config core.hooksPath .githooks` to enable project git hooks (see [Git Hooks Setup](#git-hooks-setup))

## Testing Requirements

### CLI Testing Policy

**IMPORTANT: All new CLI features MUST include comprehensive tests.**

When implementing new CLI commands or features:

1. **Create test files** in the same package with `_test.go` suffix
   - Example: `internal/cli/graph.go` → `internal/cli/graph_test.go`

2. **Required test coverage** includes:
   - ✅ **Happy path tests** - Verify the feature works correctly with valid inputs
   - ✅ **Format tests** - Test all output formats (JSON, YAML, ASCII, etc.)
   - ✅ **Flag tests** - Test all command-line flags and their combinations
   - ✅ **Error handling** - Test invalid inputs and edge cases
   - ✅ **Integration tests** - Test with real temporary files when applicable

3. **Test naming convention**:
   ```go
   func TestCommandName_FeatureDescription(t *testing.T)
   ```
   Example: `TestGraphCommand_ExcludeStatus_BugFix`

4. **Test structure**:
   ```go
   func TestMyFeature(t *testing.T) {
       // Setup
       tmpDir := createTestFiles(t)

       // Reset flags to known state
       flagVar = defaultValue

       // Execute
       err := runCommand(cmd, args)

       // Verify
       if err != nil {
           t.Fatalf("unexpected error: %v", err)
       }
       // Add specific assertions
   }
   ```

5. **Use test helpers**:
   - Create helper functions for common setup (e.g., `createTestTaskFiles`)
   - Use `t.TempDir()` for temporary directories
   - Use `t.Helper()` in helper functions

6. **Examples of good test coverage**:
   - See `internal/cli/graph_test.go` for a comprehensive example
   - See `internal/graph/graph_test.go` for package-level tests

### Running Tests

```bash
# Run unit/integration tests
cd apps/cli && go test ./...

# Run specific test
go test ./internal/cli -run TestGraphCommand

# Run with verbose output
go test -v ./internal/cli -run TestGraphCommand

# Run with coverage
go test -cover ./...
```

### Running E2E Tests

E2E tests build the real binary and invoke it as a subprocess, testing the full CLI including config loading, argument parsing, and output formatting.

```bash
# Run all e2e tests
cd apps/cli && make e2e

# Run a specific e2e test group
go test -tags e2e -count=1 -run TestConfig ./internal/e2e/...

# Run with verbose output
go test -tags e2e -count=1 -v ./internal/e2e/...
```

**Note:** `make test` does **not** include e2e tests. Run `make e2e` separately.

### Test Coverage Goals

- **CLI commands**: Minimum 80% coverage
- **Core packages** (graph, parser, validator): Minimum 90% coverage
- **Critical paths**: 100% coverage

## Code Quality Standards

### Linting

The project uses `golangci-lint` for code quality checks, following Go conventions:

```bash
# Run linter (from apps/cli directory)
make lint

# Or run directly
golangci-lint run

# Auto-fix issues where possible
make lint-fix

# Run go mod tidy
make tidy

# Run all checks (test, lint, vet)
make check
```

**Key quality metrics enforced**:
- **Function length**: Max 60 lines per function (excluding comments)
  - Promotes single responsibility principle
  - Improves testability and readability
- **Cyclomatic complexity**: Max 15 complexity per function
  - Catches overly complex branching logic
- **Cognitive complexity**: Max 20 cognitive complexity
  - Measures how hard code is to understand
- **Code quality**: errcheck, staticcheck, govet, unused code detection
- **Formatting**: gofmt, goimports

**Note on file length**: golangci-lint has no built-in file-length rule, so the primary focus is on keeping functions small and focused — if functions stay under 60 lines, files naturally remain manageable. As a soft guardrail, `scripts/check-go-file-length.sh` warns about non-test `.go` files over 300 lines. It runs as part of `make check-lite` (and thus the pre-commit hook) but is **non-blocking** — it prints warnings and always exits 0. Treat warnings as a nudge to split a file, not a hard gate. Run it directly with `make check-file-length`.

**Common lint rules**:
- Use `any` instead of `interface{}`
- No unused imports or variables
- Proper error handling (don't ignore errors)
- Consistent naming conventions
- No magic numbers - use named constants

### Go Conventions

1. **Error handling**:
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("operation failed: %w", err)
   }

   // Bad - don't ignore errors
   _ = someOperation()
   ```

2. **Use context for cleanup**:
   ```go
   func TestSomething(t *testing.T) {
       tmpDir := t.TempDir() // Auto-cleanup
       // ... test code
   }
   ```

3. **Consistent formatting**:
   - Use `gofmt` (automatically done by most editors)
   - Use `goimports` for import organization

## CLI Command Development

### Adding a New Command

1. **Create the command file**: `internal/cli/<command>.go`
2. **Define flags** as package-level variables
3. **Register command** in `init()` function with `rootCmd.AddCommand()`
4. **Create RunE function** with signature `func(cmd *cobra.Command, args []string) error`
5. **Use GetGlobalFlags()** for common flags (verbose, format, etc.)
6. **Add comprehensive tests** in `internal/cli/<command>_test.go`

### Command Structure Template

```go
package cli

import (
    "github.com/spf13/cobra"
)

var (
    // Command-specific flags
    myFlag string
)

var myCmd = &cobra.Command{
    Use:   "mycommand [args]",
    Short: "Brief description",
    Long:  `Detailed description with examples`,
    Args:  cobra.MaximumNArgs(1),
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)

    myCmd.Flags().StringVar(&myFlag, "flag", "default", "description")
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    flags := GetGlobalFlags()

    // Implementation

    return nil
}
```

## Task Management

When working on tasks, use the CLI to manage task status and dependencies.

**During development:** Use `taskmd-dev` to test your changes (see Building and Deployment section).
**For stable operations:** Use `taskmd` (the Homebrew-installed version).

### Task File Conventions

When working on tasks:

1. **Check task status** before starting:
   ```bash
   taskmd list tasks/cli
   # Or during development:
   taskmd-dev list tasks/cli
   ```

2. **Update task status** as you work:
   - Mark as `in-progress` when starting
   - Mark as `completed` when done
   - Check off subtasks `- [x]` as you complete them

   ```bash
   taskmd set 042 --status in-progress
   # Or during development:
   taskmd-dev set 042 --status in-progress
   ```

3. **Maintain a worklog** as you work (unless `worklogs: false` in `.taskmd.yaml`):
   - Create/append to `tasks/<group>/.worklogs/<ID>.md`
   - Add timestamped entries when starting, making decisions, hitting blockers, or finishing
   - See the agent template (e.g., `CLAUDE.md`) for format details and examples

4. **Reference the task specification** document:
   - See `docs/taskmd_specification.md` for task format conventions
   - Follow the defined frontmatter schema

### Task Dependencies

- Always check task dependencies before starting work
- Ensure dependent tasks are completed first
- Use the graph command to visualize dependencies:
  ```bash
  taskmd graph --format ascii --exclude-status completed
  # Or during development:
  taskmd-dev graph --format ascii --exclude-status completed
  ```

## Building and Deployment

### Development Workflow

When working on the CLI, use the development build to test your changes while keeping the Homebrew-installed stable version available.

#### Install Development Binary

```bash
cd apps/cli
make install-dev
```

This installs the binary as `taskmd-dev` in `~/bin/`, keeping your stable `taskmd` (from Homebrew) available.

**Note:** Ensure `~/bin` is in your PATH. It should already be configured in `~/.zshrc`.

#### Testing Your Changes

**ALWAYS use `taskmd-dev` when testing changes you've made to the source code:**

```bash
# After making code changes
cd apps/cli
make install-dev

# Test your changes
taskmd-dev list
taskmd-dev next
taskmd-dev set 042 --status completed

# Compare with stable version if needed
taskmd list  # Uses Homebrew version
```

#### Quick Build for Testing

For rapid iteration without installing:

```bash
cd apps/cli
make build

# Run directly
./taskmd list
```

#### Build Options

| Command | Output | Use Case |
|---------|--------|----------|
| `make build` | `apps/cli/taskmd` | Quick local testing |
| `make install-dev` | `~/bin/taskmd-dev` | Development (recommended) |
| `make install` | `$GOPATH/bin/taskmd` | Replace system binary |
| `make build-full` | `apps/cli/taskmd` (with web) | Full build with embedded web assets |

### Production Builds

For release builds with version information:

```bash
cd apps/cli
go build -ldflags="-X 'main.Version=1.0.0' -X 'main.GitCommit=$(git rev-parse HEAD)' -X 'main.BuildDate=$(date)'" -o bin/taskmd ./cmd/taskmd
```

## Documentation

### When to Update Documentation

Update documentation when:
- Adding new CLI commands or flags
- Changing existing behavior
- Adding new task file conventions
- Implementing new output formats

### Specification Sync

The taskmd specification lives in `docs/taskmd_specification.md` (the canonical source). Three other artifacts derive from it and must stay in sync:

- `apps/cli/internal/cli/templates/TASKMD_SPEC.md` (embedded in the CLI binary) — **byte-identical copy**
- `apps/docs/reference/specification.md` (docs site) — **byte-identical copy**
- `claude-code-plugin-lite/SPEC_REFERENCE.md` (embedded in the lite plugin) — **condensed subset**, not a verbatim copy

**After editing `docs/taskmd_specification.md`, always run:**

```bash
cd apps/cli && make sync-spec
```

`make sync-spec` copies the canonical spec to the two byte-identical locations, and `TestSpecTemplate_MatchesCanonicalSpec` fails if either drifts.

`SPEC_REFERENCE.md` is a hand-written condensed subset (it restructures and omits content), so it is not copied by `sync-spec`. Instead, it is **drift-checked, not freely hand-edited**: `TestSpecReference_EnumsMatchCanonical` and `TestSpecReference_RequiredFieldsMatchCanonical` fail if its enum value sets (`status`, `priority`, `effort`, `type`) or required fields (`id`, `title`) contradict the canonical spec. When you change those facts in the canonical spec, update `SPEC_REFERENCE.md` in the same commit.

### Documentation Locations

- **CLI commands**: Help text in the command definition
- **Task format**: `docs/taskmd_specification.md` (canonical spec — see Specification Sync above)
- **Development**: This file (`CLAUDE.md`)
- **Project overview**: `README.md` (monorepo: Go CLI in `apps/cli`, Vite + React 19 web app in `apps/web`, Go SDK in `sdk/go`, docs in `apps/docs`)

## Common Patterns

### Scanner Usage

```go
taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
result, err := taskScanner.Scan()
if err != nil {
    return fmt.Errorf("scan failed: %w", err)
}

tasks := result.Tasks
```

### Output Formatting

Support multiple output formats consistently:

```go
switch flags.Format {
case "json":
    return outputJSON(data, outFile)
case "yaml":
    return outputYAML(data, outFile)
case "table":
    return outputTable(data, outFile)
default:
    return fmt.Errorf("unsupported format: %s", flags.Format)
}
```

### File Writing

```go
var outFile *os.File
if outputPath != "" {
    f, err := os.Create(outputPath)
    if err != nil {
        return fmt.Errorf("failed to create output file: %w", err)
    }
    defer f.Close()
    outFile = f
} else {
    outFile = os.Stdout
}
```

## Git Workflow

### Git Hooks Setup

This project uses a `.githooks/` directory for version-controlled git hooks. After cloning, configure Git to use them:

```bash
git config core.hooksPath .githooks
```

**Active hooks:**
- **pre-commit**: Runs `taskmd validate` to check task files, `make check-lite` (compile + lint), and `scripts/check-sdk-pin.sh` (see below). The commit is blocked if any of them fail.

### The sdk/go pin

`apps/cli` and `sdk/go` are **separate Go modules**. `go.work` makes in-repo builds
resolve `sdk/go` locally, so a stale `sdk/go` pin in `apps/cli/go.mod` is invisible
during development. External consumers have no workspace:

```bash
go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest
```

reads `apps/cli/go.mod`, so if the CLI uses SDK symbols added after the pinned
commit, that install fails to compile. This shipped once already — see
[issue #8](https://github.com/driangle/taskmd/issues/8) and
[PR #9](https://github.com/driangle/taskmd/pull/9). There are no `apps/cli/vX.Y.Z`
tags, so `@latest` resolves to **main HEAD**: drift breaks users as soon as it lands
on main, not at release time.

Three guards keep the pin honest:

| Guard | When | Behavior |
|-------|------|----------|
| `scripts/check-sdk-pin.sh --staged` | pre-commit hook | Warns on the commit that *introduces* SDK changes (the pin can't reference an unpushed commit), then **blocks** any later commit that leaves the pin stale |
| `make check-sdk-pin` / CI `sdk-pin` job | pushes to main | Any drift is an error; also builds with `GOWORK=off` to exercise the real `go install` path |
| `scripts/release.sh` | every release | Tags and pushes `sdk/go` when it changed, repoints the pin at that version, and verifies the `GOWORK=off` build before tagging the CLI |

### Versioning: two independent modules, two tags

`apps/cli` and `sdk/go` version **independently**, because the SDK's version numbers
have to mean something to people importing the library:

| Tag | Module | When |
|-----|--------|------|
| `vX.Y.Z` | the CLI / repo release | every release |
| `sdk/go/vX.Y.Z` | `github.com/driangle/taskmd/sdk/go` | only when `sdk/go` changed |

The `sdk/go/` prefix is not decoration — Go requires a module in a subdirectory to be
tagged with its directory path, or the tag is invisible to `go get`.

`apps/cli/go.mod` pins a **released** SDK version (`v0.4.0`), not a pseudo-version
(`v0.0.0-20260811122305-775ccf445961`). Both work, but a pseudo-version is an opaque
commit pointer that no reviewer can evaluate — which is how the pin went stale twice.

The SDK is pre-1.0, so under semver a breaking API change is a **minor** bump
(`v0.4.0` → `v0.5.0`), not a major one. Going to `v1.0.0` would promise stability;
going to `v2.0.0` or beyond would additionally require the major version in the import
path (`github.com/driangle/taskmd/sdk/go/v2`) and a repo-wide import rewrite.

**Module versions are immutable.** Once a tag is pushed and the Go module proxy has
fetched it, that version is fixed forever — you cannot retag it. Pick the next number
instead.

### Workflow when you change `sdk/go`

During development, `go.work` means you change `sdk/go` and `apps/cli` together and
everything just builds — no pin bump needed per commit. The pre-commit hook will remind
you that the pin is behind, and block unrelated commits until it is resolved.

The pin is normally repointed **at release time** by `scripts/release.sh`:

```bash
./scripts/release.sh 0.3.1 --sdk-version 0.4.1 --notes-file notes.md
```

If `sdk/go` changed and you omit `--sdk-version`, the script stops and tells you.
If `sdk/go` did not change, omit it and no SDK tag is created.

To repoint the pin by hand between releases, the SDK tag must already be pushed:

```bash
git tag -a sdk/go/v0.4.1 -m "sdk/go v0.4.1" && git push origin sdk/go/v0.4.1
cd apps/cli && go get github.com/driangle/taskmd/sdk/go@v0.4.1 && go mod tidy
GOWORK=off go build ./cmd/taskmd    # verify
```

### Commit Messages

Follow conventional commit format:

```
type(scope): brief description

Longer description if needed

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `test`: Adding tests
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

### Before Committing

1. **Run tests:** `make test` or `go test ./...`
2. **Run e2e tests:** `make e2e`
3. **Run linter:** `make lint` or `golangci-lint run`
4. **Build successfully:** `make build` or `go build ./...`
5. **Test with development binary:** Install and test your changes
   ```bash
   make install-dev
   taskmd-dev list    # Test basic functionality
   taskmd-dev next    # Test your specific changes
   ```
6. **Update relevant documentation**

All checks (test, lint, build, manual testing) should pass before committing.

## Troubleshooting

### Tests Failing

1. Check if flags are properly reset between tests
2. Verify temporary directories are being used
3. Look for race conditions or shared state

### Build Failures

1. Check for missing dependencies: `go mod tidy`
2. Verify Go version compatibility
3. Check for syntax errors: `go vet ./...`

### Linting Issues

1. Run auto-fix: `golangci-lint run --fix`
2. Check for unused imports
3. Verify error handling patterns

## Performance Considerations

### Large Repositories

When working with large task repositories:
- Use streaming where possible
- Avoid loading all tasks into memory at once
- Consider pagination for output
- Use efficient data structures (maps for lookups)

### Testing Performance

For performance-critical code:
- Add benchmark tests: `func BenchmarkMyFunction(b *testing.B)`
- Run benchmarks: `go test -bench=.`
- Profile if needed: `go test -cpuprofile=cpu.prof`

## Security

### Input Validation

Always validate:
- File paths (prevent directory traversal)
- Task IDs (prevent injection)
- User input in flags

### Safe File Operations

```go
// Good - validate paths
if !strings.HasPrefix(filepath.Clean(userPath), basePath) {
    return fmt.Errorf("invalid path")
}

// Good - proper permissions
os.WriteFile(path, data, 0644)
```

## Questions or Issues?

- Check existing tasks for similar implementations
- Review test files for usage examples
- Refer to `docs/taskmd_specification.md` for task format questions
- Use the graph command to understand dependencies

---

**Last Updated**: 2026-02-08
**Maintained By**: taskmd contributors
