# ADR 0002 — The lite plugin's prose is a derived copy; the CLI is authoritative

- **Status:** Accepted
- **Date:** 2026-08-04
- **Deciders:** taskmd maintainers
- **Related task:** `01kz3c244` — Add conformance check for lite plugin prose logic vs CLI

## Context

`claude-code-plugin-lite` is a CLI-free plugin: it runs taskmd operations using
Claude's native file tools instead of the Go binary. To do that, its 13
`SKILL.md` files **re-express core CLI algorithms as English prose** — ID
generation (sequential / prefixed / random / ULID), the title→slug rule, the
new-task frontmatter template, `.taskmd.yaml` parsing, and the validation enums.

This means several of taskmd's core algorithms now exist twice: once in Go
(`sdk/go/nextid`, `sdk/go/slug`, `sdk/go/validator`, and the `add` command's
frontmatter writer), and once as prose. Unlike the specification — where
`spec_reference_test.go` and `sync-spec` already detect drift — the prose copy
had **no guard**. It could silently diverge from real CLI behavior, and every
CLI behavior change had to be manually re-expressed across the skill files with
nothing to catch a miss. A concrete example found while writing this ADR: the
`add-task` prose claimed random IDs "contain at least one digit," which the CLI
never guarantees.

## Decision

**The Go CLI is the single source of truth for every behavior the lite plugin
reimplements. The prose is a derived copy that must follow the CLI.** When the
two disagree, the CLI wins and the prose is wrong.

This contract is enforced, not just documented:

- `apps/cli/internal/cli/lite_conformance_test.go` derives each duplicated fact
  from **real CLI code** (calls `nextid`/`slug` SDK functions, runs the `add`
  command's frontmatter writer, reads the in-package enum lists) and asserts the
  relevant `SKILL.md` documents that same fact. It covers: the four ID
  strategies and their shapes/charsets, the sequential zero-padding default, the
  slug rules (lowercase / hyphenation / max length), the always-written
  frontmatter fields and the status default, and the status/priority/effort/type
  enum value sets.
- `apps/cli/internal/cli/spec_reference_test.go` (pre-existing) continues to
  guard `SPEC_REFERENCE.md` against the canonical spec document.

Both run in CI through `go test ./...` from `apps/cli`, so drift fails the build.

## Consequences

- A CLI behavior change that the prose doesn't track now **fails CI** with a
  message naming the `SKILL.md` to update. The fix is to update the prose, never
  to weaken the test.
- The conformance test asserts load-bearing *facts* (charsets, defaults, field
  sets, enum sets), not byte-identical wording — the prose stays readable and
  the CLI's minimal output need not match the prose's fuller schema description
  verbatim.
- A small number of facts are not yet machine-derivable from exported CLI code
  (e.g. the default random/ULID `length` is a config literal in `root.go`, not
  an exported constant). These remain documented conventions; promoting them to
  exported constants would let future conformance checks source them directly.
- New foreign behaviors added to the lite prose are expected to come with a
  matching CLI source of truth and, where practical, a conformance assertion.
