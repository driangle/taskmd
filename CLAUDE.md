# Development Guidelines for taskmd

This document provides guidelines and conventions for developing the taskmd project. These instructions are designed to help maintain code quality, consistency, and reliability across the codebase.

## Prerequisites

Before developing, ensure you have the following tools installed:

- **Go** (1.22+): [https://go.dev/dl/](https://go.dev/dl/)
- **pnpm**: `npm install -g pnpm` (for web frontend)
- **golangci-lint**: Required for `make lint` and `make lint-fix`
  - macOS: `brew install golangci-lint`
  - go install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
  - Other: See [golangci-lint install docs](https://golangci-lint.run/welcome/install/)

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
# Run all tests
cd apps/cli && go test ./...

# Run specific test
go test ./internal/cli -run TestGraphCommand

# Run with verbose output
go test -v ./internal/cli -run TestGraphCommand

# Run with coverage
go test -cover ./...
```

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

**Note on file length**: Unlike some languages, Go doesn't enforce file-length limits. Instead, the focus is on keeping functions small and focused. If functions stay under 60 lines, files naturally remain manageable.

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

Assume `taskmd` CLI is installed globally. You can run with `taskmd --help` for usage instructions.
### Task File Conventions

When working on tasks:

1. **Check task status** before starting:
   ```bash
   taskmd list tasks/cli
   ```

2. **Update task status** as you work:
   - Mark as `in-progress` when starting
   - Mark as `completed` when done
   - Check off subtasks `- [x]` as you complete them

3. **Reference the task specification** document:
   - See `docs/TASKMD_SPEC.md` for task format conventions
   - Follow the defined frontmatter schema

### Task Dependencies

- Always check task dependencies before starting work
- Ensure dependent tasks are completed first
- Use the graph command to visualize dependencies:
  ```bash
  taskmd graph --format ascii --exclude-status completed
  ```

## Building and Deployment

### Building the CLI

```bash
cd apps/cli
make build-full
```

### Build Flags

For release builds, use version information:

```bash
go build -ldflags="-X 'main.Version=1.0.0' -X 'main.GitCommit=$(git rev-parse HEAD)' -X 'main.BuildDate=$(date)'" -o bin/taskmd ./cmd/taskmd
```

## Documentation

### When to Update Documentation

Update documentation when:
- Adding new CLI commands or flags
- Changing existing behavior
- Adding new task file conventions
- Implementing new output formats

### Documentation Locations

- **CLI commands**: Help text in the command definition
- **Task format**: `docs/TASKMD_SPEC.md`
- **Development**: This file (`CLAUDE.md`)
- **Project overview**: `PLAN.md`

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

1. Run tests: `make test` or `go test ./...`
2. Run linter: `make lint` or `golangci-lint run`
3. Build successfully: `make build` or `go build ./...`
4. Update relevant documentation

All three checks (test, lint, build) should pass before committing.

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
