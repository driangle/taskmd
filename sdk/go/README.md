# taskmd Go SDK

```
import "github.com/driangle/taskmd/sdk/go/<package>"
```

The task model behind the [taskmd](https://github.com/driangle/taskmd) CLI as a
Go module: parse task markdown files, scan directories, validate, build the
dependency graph, rank what to work on next, and write task files back. The CLI,
MCP server, and web dashboard are all built on these packages; nothing here
prints, reads flags, or touches the network.

- **Module:** `github.com/driangle/taskmd/sdk/go` (a separate module from the CLI; tags are `sdk/go/vX.Y.Z`)
- **Go:** 1.24 or later
- **Status:** pre-1.0. Any `v0.x` to `v0.(x+1)` bump may break the API; patch bumps are additive or fixes. Pin a version.

```bash
go get github.com/driangle/taskmd/sdk/go@latest
```

## Packages

| Package | Purpose |
|---------|---------|
| `model` | Core types: `Task`, statuses, priorities, phases, flexible dates |
| `parser` | Parse one task file or in-memory content (YAML frontmatter plus markdown body) |
| `scanner` | Walk a directory tree, parse every task file, report per-file errors |
| `taskfile` | Read and update task files on disk, preserving unknown fields and body |
| `validator` | Structural and semantic checks on a task set and on `.taskmd.yaml` |
| `graph` | Dependency graph: cycles, critical path, depth, upstream and downstream |
| `next` | Score and rank actionable tasks, with phase and priority ordering |
| `tracks` | Group actionable tasks into parallel tracks by scope overlap |
| `filter` | Apply `field=value` filter expressions to a task collection |
| `search` | Full-text search across titles and bodies |
| `board` | Group tasks into columns by a field |
| `metrics` | Aggregate statistics over a task set |
| `feed` | Recent task activity reconstructed from git history of the task files (git access is injected, not called directly) |
| `effort` | The configurable effort vocabulary and its ordering |
| `nextid` | Generate the next ID for the sequential, prefixed, random, and ULID strategies |
| `slug` | Title to filename slug |
| `verify` | Run the `verify` steps declared in task frontmatter |
| `worklog` | Timestamped progress entries alongside a task |

## Example

Scan a directory, validate, and ask what to work on next:

```go
package main

import (
	"fmt"
	"log"

	"github.com/driangle/taskmd/sdk/go/next"
	"github.com/driangle/taskmd/sdk/go/scanner"
	"github.com/driangle/taskmd/sdk/go/validator"
)

func main() {
	result, err := scanner.NewScanner("./tasks", false, nil).Scan()
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range result.Errors {
		log.Printf("skipped %s: %v", e.FilePath, e.Error)
	}

	report := validator.NewValidator(false).Validate(result.Tasks)
	if !report.IsValid() {
		log.Fatalf("%d validation error(s)", report.Errors)
	}

	recs, err := next.Recommend(result.Tasks, next.Options{Limit: 3})
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range recs {
		fmt.Printf("%d. [%s] %s (score %d)\n", r.Rank, r.ID, r.Title, r.Score)
	}
}
```

## Where things go

The SDK is the *pure task model layer*. Code belongs here if it parses,
validates, scans, scores, searches, or writes task files and would make sense in
a program that only ever touched the local filesystem. Code that formats output,
knows how it was invoked, or reaches outside the task files (git, network,
watching) belongs in the CLI. The full rule is
[ADR 0006](../../docs/adr/0006-sdk-is-the-pure-task-model-layer.md).

## Versioning and the CLI pin

`apps/cli/go.mod` pins a released SDK version. During development `go.work`
makes the CLI build against the in-tree SDK, and CI repoints the pin after SDK
changes land on `main`. A commit that breaks the SDK API must carry an
`sdk-bump: minor` trailer so the next tag is a minor bump. Details are in the
"Versioning" and "The sdk/go pin" sections of [`AGENTS.md`](../../AGENTS.md).

## Conformance fixtures

[`tests/conformance`](../../tests/conformance) holds fixture task files and the
outputs the CLI produces for them, for checking that another implementation of
the [specification](../../docs/taskmd_specification.md) agrees.
