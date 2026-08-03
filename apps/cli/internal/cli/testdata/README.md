# CLI test fixtures

Shared, canonical task-repo fixtures for the `cli` package tests. Each
subdirectory is one named **fixture set** — a tree of task `.md` files that the
test harness copies wholesale into an isolated temp repo.

Seed a repo from a set with the harness helpers in `harness_test.go`:

```go
repo := newTaskRepoFromFixture(t, "dependency-chain")
res := repo.Run("get", "001", "--format", "json")
```

`repo.SeedFixture(name)` overlays another set onto an existing repo, and
`repo.Write(name, content)` still adds one-off inline files on top.

## Available sets

| Set | Shape | Notes |
|-----|-------|-------|
| `dependency-chain` | 001-setup → 002-auth → 003-ui, a 3-task linear dependency chain with mixed statuses/priorities/tags | The most duplicated shape across the suite (get/status/set). |
| `parent-children` | 010 parent with two children (011 pending, 012 completed) | Exercises parent/child rendering. |
| `phases` | one phased task (p01, `phase: beta`) and one un-phased (p02) | Exercises phase display/omission. |
| `subdir-projects` | tasks in `cli/` and `backend/` subdirs, including an intentionally ambiguous `055-api.md` in both | Exercises file-path matching and ambiguity resolution. |

## Conventions

- One task per file, named `<id>-<slug>.md`, matching the real repo layout.
- Prefer adding a **new set** over mutating an existing one — a set is shared, so
  changing it can ripple across unrelated tests.
- Keep genuinely one-off, test-specific fixtures **inline** in the test (e.g.
  malformed-frontmatter, duplicate-ID, or verify-block permutations). A shared
  fixture is only worth it when the same shape recurs and sharing does not
  obscure what a given test is checking.
