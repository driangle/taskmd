# Contributing

Guidelines and conventions for developing the taskmd project.

## Prerequisites

- **Go** (1.22+): [go.dev/dl](https://go.dev/dl/)
- **pnpm**: `npm install -g pnpm` (for web frontend)
- **golangci-lint**: `brew install golangci-lint` or [install docs](https://golangci-lint.run/welcome/install/)

## Development Setup

```bash
# Clone repository
git clone https://github.com/driangle/taskmd.git
cd taskmd

# Build CLI
cd apps/cli
make build

# Run tests
make test

# Run linter
make lint

# Run all checks (test, lint, vet)
make check

# Build with embedded web UI
make build-full
```

### Development Binary

Install as `taskmd-dev` to keep your stable installation separate:

```bash
cd apps/cli
make install-dev    # Installs to ~/bin/taskmd-dev

# Test your changes
taskmd-dev list
taskmd-dev next
```

### Build Options

| Command | Output | Use Case |
|---------|--------|----------|
| `make build` | `apps/cli/taskmd` | Quick local testing |
| `make install-dev` | `~/bin/taskmd-dev` | Development (recommended) |
| `make install` | `$GOPATH/bin/taskmd` | Replace system binary |
| `make build-full` | `apps/cli/taskmd` (with web) | Full build with embedded web assets |

## Testing Requirements

**All new CLI features must include comprehensive tests.**

### Required Coverage

- Happy path tests - verify the feature works correctly
- Format tests - test all output formats (JSON, YAML, ASCII, etc.)
- Flag tests - test all command-line flags and combinations
- Error handling - test invalid inputs and edge cases
- Integration tests - test with real temporary files when applicable

### Test Naming Convention

```go
func TestCommandName_FeatureDescription(t *testing.T)
```

### Running Tests

```bash
cd apps/cli

# Run all tests
go test ./...

# Run specific test
go test ./internal/cli -run TestGraphCommand

# Run with verbose output
go test -v ./internal/cli -run TestGraphCommand

# Run with coverage
go test -cover ./...
```

### Coverage Goals

- **CLI commands**: Minimum 80%
- **Core packages** (graph, parser, validator): Minimum 90%
- **Critical paths**: 100%

## Code Quality

### Linting

```bash
# Run linter
make lint

# Auto-fix issues
make lint-fix

# Run go mod tidy
make tidy
```

**Enforced metrics:**
- Function length: max 60 lines
- Cyclomatic complexity: max 15
- Cognitive complexity: max 20

### Go Conventions

**Error handling:**
```go
// Good
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Bad - don't ignore errors
_ = someOperation()
```

**Common rules:**
- Use `any` instead of `interface{}`
- No unused imports or variables
- Consistent naming conventions
- No magic numbers - use named constants

## Adding a CLI Command

1. Create `internal/cli/<command>.go`
2. Define flags as package-level variables
3. Register in `init()` with `rootCmd.AddCommand()`
4. Create `RunE` function: `func(cmd *cobra.Command, args []string) error`
5. Use `GetGlobalFlags()` for common flags
6. Add tests in `internal/cli/<command>_test.go`

### Template

```go
package cli

import "github.com/spf13/cobra"

var myFlag string

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

## Git Workflow

### Commit Messages

Follow conventional commit format:

```
type(scope): brief description

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

Types: `feat`, `fix`, `test`, `docs`, `refactor`, `chore`

### Before Committing

1. Run tests: `make test`
2. Run linter: `make lint`
3. Build successfully: `make build`
4. Test with dev binary: `make install-dev && taskmd-dev list`

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
