# Git Worktree Support

**Status:** Implemented (2026-08-23; owner-aware `next` implemented and reverted — see §6)
**Date:** 2026-08-22
**Related ADRs:** [0004 — Local git metadata is core](../adr/0004-local-git-metadata-is-core.md),
[0005 — Worktrees are facets of one project](../adr/0005-worktrees-are-facets-of-one-project.md)
**Related design docs:** [projects.md](../design/projects.md)
**Shipped behavior:** [CLI guide — Git Worktrees](../../apps/docs/guide/cli.md#git-worktrees),
[configuration reference](../../apps/docs/reference/configuration.md#worktrees-configuration),
[MCP guide](../../apps/docs/guide/mcp.md), [web guide](../../apps/docs/guide/web.md);
code in `apps/cli/internal/gitmeta` and `apps/cli/internal/cli/worktree_*.go`

## Problem

Users (and especially agents) work on multiple git worktrees of the same repository
simultaneously. Each worktree has its own branch-local copy of the task files, so a
task's status diverges between worktrees until branches merge.

taskmd is unaware of worktrees. The consequences:

1. **Double-assignment.** An agent in worktree A claims a task
   (`taskmd set <id> --status in-progress`). The file change exists only in A's
   checkout. Agents in worktrees B and C run `taskmd next` against their own copies,
   where the task is still `pending`, and pick up the same task.
2. **Identity fragmentation.** Each worktree looks like an unrelated project. The
   global registry (`~/.taskmd.yaml`) would need one entry per worktree, each with a
   distinct id, and `--all-projects` would count the same repo N times.
3. **Stale reads.** `list`, `board`, and `stats` in any one worktree show a partial
   view of reality — work completed on a sibling branch is invisible.

There is also a pre-existing coordination gap independent of worktrees: the `next`
recommender treats `in-progress` tasks as actionable and ignores the `owner` field,
so two agents in the *same* checkout can be handed the same task.

Prior art in-repo: task `01kzdpvr1` documented a real bug where an agent `cd`-ed to
the primary checkout to run taskmd writes, flipping status in the wrong worktree. The
"pin the CLI to the primary checkout" approach was rejected there — task state must
travel with the branch that contains the work.

## Goals

1. From any worktree, `taskmd next` never recommends a task that is `in-progress`
   (or further along) in a sibling worktree.
2. Read views (`list`, `board`, `stats`, `get`, `graph`) can present a merged view
   across worktrees, with provenance ("in-progress in worktree `agent-b`").
3. All worktrees of one repository resolve to a **single** project identity — one
   registry entry, one `--project` id, no double-counting in `--all-projects`.
4. Degrade to exactly today's behavior when git is absent, the directory is not a
   repo, or the repo has a single worktree.

## Non-goals

- **Creating or managing worktrees.** taskmd never runs `git worktree add/remove`,
  never touches branches or refs. Orchestration stays in skills
  (e.g. `divide-and-conquer`).
- **Shared mutable state.** No lock files, no claims sidecar, no daemon in v1. The
  overlay is read-only inference from task files that already exist. (A claims
  sidecar is a possible v2 escalation — see Open questions.)
- **Cross-machine coordination.** Worktrees on one filesystem only. Anything that
  travels over the network is out of scope per ADR 0001.
- **Content merging.** The overlay merges *coordination state* (status, owner), not
  task content. Body text, subtasks, and other frontmatter always come from the
  local copy. Git remains the merge tool for content.
- **Write redirection.** Mutations (`set`, `add`, `rm`, `archive`) always write to
  the current worktree's files. Rejected alternative — see ADR 0005.
- **Owner-based coordination.** An earlier revision made `next` owner-aware
  (exclude in-progress tasks claimed by others, plus a `--for <owner>` flag) to
  close the same-checkout multi-agent race. Implemented, then rejected: the
  coordination unit is **one agent per worktree**, where effective status alone
  already prevents double-assignment; `owner` stays pure display/assignment
  metadata, and repurposing it as a claim would silently hide a solo user's own
  in-progress tasks the moment they fill the field in. If same-checkout
  coordination ever becomes real, the claims sidecar (Open question 1) is the
  answer, not `owner`. See §6.

## Terminology

- **Primary worktree** — the original checkout; its `.git` is a directory.
- **Linked worktree** — created by `git worktree add`; its `.git` is a *file*
  containing a `gitdir:` pointer.
- **Repo identity** — the git *common directory* (`git rev-parse --git-common-dir`),
  shared by the primary and all linked worktrees. Two paths with the same common
  directory are the same repository.

## Design

### 1. Repo identity resolution

A small helper (new package `apps/cli/internal/gitmeta`) shells out to git, mirroring
the existing shellout style in `feed.go` / `commit_msg.go`:

```go
// Identity describes the repository containing dir, if any.
type Identity struct {
    CommonDir    string // absolute path to the shared .git directory
    WorktreeRoot string // top level of the worktree containing dir
    IsLinked     bool   // true when dir is inside a linked worktree
}

// Resolve returns (nil, nil) when dir is not inside a git repo or git is not
// installed — callers treat nil as "worktree features inert".
func Resolve(dir string) (*Identity, error)
```

Implemented with a single invocation:
`git -C <dir> rev-parse --path-format=absolute --git-common-dir --git-dir --show-toplevel`.
`IsLinked` is true when the git dir differs from the common dir.

Failure handling: git missing from `PATH`, not a repo, or any git error → `nil`
identity, feature silently off (log under `--debug`). taskmd must keep working in
non-git directories exactly as today.

### 2. Worktree discovery

```go
// Worktree is one checkout of the repository.
type Worktree struct {
    Root     string // worktree top-level path
    Branch   string // e.g. "dnc/01kz.../parser" ("" when detached)
    IsLocal  bool   // true for the worktree the command runs in
}

func ListWorktrees(id *Identity) ([]Worktree, error)
```

Backed by `git worktree list --porcelain`. Entries are filtered:

- **prunable** worktrees (deleted directories) are skipped
- worktrees whose root no longer exists on disk are skipped
- worktrees with no `.taskmd.yaml` at their root are skipped — they have opted out
  of taskmd or predate it; warn under `--verbose`

For each remaining sibling, the tasks dir is resolved by reading *that worktree's*
`.taskmd.yaml` directly (raw `os.ReadFile` + `yaml.Unmarshal` of the `dir` /
`task-dir` keys, relative paths resolved against the worktree root). This
deliberately bypasses viper: the registry loader (`registry.go`) and
`resolveProjectScanDir` (`projects.go`) already established that pattern to avoid
mutating global viper state mid-command.

### 3. The overlay: merged task view

When enabled (see Activation) and the repo has more than one worktree, commands that
scan tasks build a **worktree overlay**:

1. Scan the local worktree's tasks dir as today. This is the **base**: content,
   ordering, and every field come from local files.
2. Scan each sibling worktree's tasks dir with the existing scanner
   (`sdk/go/scanner`), reusing the `newTaskScanner` construction seam in
   `apps/cli/internal/cli/scan.go`.
3. Index sibling tasks by task ID. For each local task that also exists in a
   sibling, compute the **effective status**: the most advanced status across all
   copies, by this ladder:

   ```
   pending < blocked < in-progress < in-review < cancelled < completed
   ```

   Ties across different statuses at the same rank cannot occur (the ladder is
   total); ties for the *winning copy* (same status in two worktrees) are broken by
   file mtime, newest wins. When the winning copy is remote, the task is annotated
   with provenance: the worktree root basename and branch.
4. Tasks that exist **only** in a sibling worktree (created on another branch) are
   appended to read views, annotated the same way. They are visible but not
   addressable by mutations (see §5).

The overlay result is carried on a wrapper, generalizing the existing
`ProjectTask` pattern from `all_projects.go`:

```go
// OverlayTask decorates a task with cross-worktree provenance.
type OverlayTask struct {
    *model.Task                 // base copy (local when one exists)
    EffectiveStatus string      // max status across copies
    EffectiveOwner  string      // owner on the winning copy
    Worktree        string      // "" when the winning copy is local
    Branch          string      // branch of the winning copy
    LocalOnly       bool        // no sibling copy exists
    RemoteOnly      bool        // no local copy exists
}
```

The scanner itself is untouched — the overlay is a merge layer in
`internal/cli`, parallel to (and eventually shared with) the all-projects merge.

Divergence warning: if two copies of a task are in *different terminal states*
(`completed` in one worktree, `cancelled` in another), read views show the winner
per the ladder and print a one-line warning; `taskmd validate` reports it as a
warning when the overlay is active.

### 4. Command behavior

| Command | Behavior with overlay active |
|---------|------------------------------|
| `next` | Recommends against **effective** status: a task `in-progress`/`in-review`/`completed` in any sibling is not actionable. Local `in-progress` tasks keep today's resume semantics. `--explain` names the excluding worktree. |
| `list` | Extra `WORKTREE` column (only rendered when the overlay is active and at least one task is annotated). `--status` filters on effective status. Sibling-only tasks included, marked. |
| `board`, `stats`, `graph`, `report`, `metrics`, `tracks`, `phases` | Operate on effective status. |
| `get` | Shows the local copy, plus a `Worktrees:` section listing each copy's status/owner/branch when copies differ. |
| `set`, `add`, `rm`, `archive` | **Unchanged: local files only.** `set` on an ID that resolves only to a sibling copy fails with: `task 042 exists only in worktree ../agent-b (branch dnc/042/parser); run taskmd there`. This is the guard task `01kzdpvr1` asked for. |
| `validate` | Warns on divergent terminal states across worktrees. Duplicate IDs *within* one worktree remain an error, as today; the same ID across worktrees is the expected case, never a duplicate. |
| `mcp` / web | Serve the merged view for reads; mutations keep local-only semantics. See §9. |

### 5. Activation

```yaml
# .taskmd.yaml
worktrees: auto   # auto | true | false   (default: auto)
```

- `auto` — overlay activates when the directory is inside a git repo with more than
  one worktree. Single-worktree repos and non-repos: zero behavior change, zero
  extra git invocations beyond the one identity probe.
- `true` — always attempt (still inert outside a repo).
- `false` — never; today's behavior.

Per-invocation override: a persistent `--worktrees=<auto|true|false>` global flag,
mirroring how other global flags work in `root.go`. Env: `TASKMD_WORKTREES`.

### 6. Owner-aware `next` — removed

This section previously specified owner-based coordination (in-progress tasks
claimed via `owner` excluded from `next`, plus a `--for <owner>` flag). It was
implemented (2026-08-22), then reverted the same day and moved to Non-goals —
see the rationale there. The section number is retained so cross-references to
§7–§9 stay valid. `EffectiveOwner` in §3 remains: `owner` is still merged and
displayed as provenance, it just carries no exclusion semantics.

### 7. Projects registry integration

A worktree is not a new project — it is another mount of the same project:

- **`projects register`** resolves repo identity before storing. The stored `path`
  is the **primary worktree's** root (derived from the common dir), even when
  registered from a linked worktree. Registering a second worktree of an
  already-registered repo is a friendly no-op ("already registered as `taskmd`").
- **cwd → project matching** (`resolveTaskDir`, `--project` no-op detection)
  compares repo identity, not path prefixes, when the cwd is in a git repo. Running
  `taskmd list --project taskmd` from any worktree of taskmd scopes to *the current
  worktree's* task dir (local base + overlay), not the primary's.
- **`--all-projects`** counts each repo once. The scanned root for a repo is the
  worktree the command runs in when inside that repo, else the registered primary;
  the overlay applies per-repo when active.
- Registry entries are still plain `{id, name, path}` — no schema change to
  `~/.taskmd.yaml`.

### 8. Interaction with the scanner's hidden-dir rule

The scanner skips hidden directories, so `.claude/worktrees/*` nested inside a
checkout is never accidentally double-scanned today — that stays. Sibling worktrees
are reached by their absolute roots from `git worktree list`, not by directory
walking, so the hidden-dir rule and the overlay never conflict. Conversely, a
worktree checked out to a *non-hidden* path inside the scan root (unusual but legal)
would previously double-scan and trigger duplicate-ID warnings; with the overlay
active, copies whose file path lies inside a sibling worktree root are attributed to
that worktree instead of being flagged as duplicates.

### 9. Web and MCP surfaces

Both surfaces sit on cached data providers, so the overlay lands in one place per
surface rather than per endpoint.

**Data layer.** The web server's `DataProvider` (and the MCP server's equivalent)
builds the overlay when active, so every read endpoint — list, board, graph, stats,
task detail — serves effective status and owner. Task payloads gain additive
provenance fields (`effective_status`, `effective_owner`, `worktree`, `branch`,
`remote_only`); existing fields keep their meaning, so clients that ignore the new
fields keep working. For the `taskmd-mcp` plugin this is an additive tool-surface
change — a **minor** bump per ADR 0003.

**Live refresh.** Today the web server's filesystem watcher covers only the local
scan dir; a change there invalidates the provider cache and pushes an SSE broadcast
to the browser. With the overlay active that is insufficient — a claim made in a
sibling worktree changes what the server should display. The watcher must
additionally cover:

- each sibling worktree's tasks dir (same invalidate + broadcast path, same
  debounce), and
- worktree membership itself: watch `<common-dir>/worktrees/` so `git worktree
  add`/`remove` triggers re-enumeration; as a fallback, re-list worktrees on any
  invalidation.

Watching paths outside the current worktree is still read-only local-filesystem
access, within ADR 0004's limits.

**Mutations.** Web edits and MCP mutation tools keep strictly-local write semantics.
A mutation targeting a sibling-only task returns the §4 guard error, and the web UI
must surface it as a visible failure — never a silent no-op.

**Frontend.** The React app renders provenance: a worktree badge on list rows and
board cards, a per-worktree copies section on the task detail page (mirroring
`get`), and a header indicator when the overlay is active ("worktree `agent-b` —
3 siblings"). TypeScript API types are extended with the new fields. No frontend
change is visible in single-worktree repos.

**Static export.** The exported site is a snapshot, so it bakes in effective status
and provenance at export time. Exports of single-worktree repos are byte-identical
to today's.

## Performance

Overlay cost is one `git rev-parse` + one `git worktree list` + one extra directory
scan per sibling worktree. Task dirs are small (hundreds of files); the scans are
the same cost as `--all-projects` with N projects, which is already accepted. No
caching in v1; if it ever matters, cache sibling scans keyed by dir mtime.

## Testing

Per the CLI testing policy (CLAUDE.md):

- **Unit tests** (`internal/cli`): overlay merge rules (status ladder, mtime
  tie-break, remote-only/local-only, divergent terminal states), owner-aware
  recommender rules in `sdk/go/next`, activation matrix (`auto`/`true`/`false` ×
  repo/no-repo/one-worktree/many). Worktree discovery is injected so merge logic
  tests need no git.
- **E2E tests** (`internal/e2e`, `-tags e2e`): build the binary, create a real repo
  with `git worktree add`, set a task `in-progress` in worktree A, assert `next` in
  worktree B skips it and `list` shows provenance; assert `set` on a sibling-only ID
  fails with the guard message; assert graceful no-git behavior by scrubbing `PATH`.
- **Registry tests**: register from a linked worktree stores the primary path;
  re-register from another worktree no-ops.
- **Web live-refresh tests**: editing a task file in a sibling worktree invalidates
  the provider cache, fires an SSE broadcast, and the next payload reflects the new
  effective status; `git worktree add`/`remove` triggers re-enumeration.

## Migration

None. The feature is additive and defaults to `auto`, which only changes behavior
in multi-worktree repos — where current behavior is the bug being fixed. Users who
depend on per-worktree isolation set `worktrees: false`.

## Implementation order

1. **`gitmeta` package** — identity resolve + worktree list, with e2e coverage.
2. **Overlay merge layer** — `OverlayTask`, effective status, provenance; wire into
   `next` and `list` first.
3. **Remaining read views** — board/stats/graph/get/validate annotations; MCP/web.
4. **Registry identity** — common-dir resolution in register / cwd-matching /
   `--all-projects` dedupe.
5. **Docs** — spec-sync surfaces if any wording in `docs/taskmd_specification.md`
   changes (run `make sync-spec`).

## Open questions

1. **Claims sidecar escalation.** The overlay leaves a small race: two agents
   running `next` in the same second, before either writes `in-progress`. If this
   matters in practice, v2 adds `taskmd claim <id>` writing a lease
   (`{task, owner, worktree, timestamp}`) to `<git-common-dir>/taskmd/claims.yaml`
   — shared across worktrees, never versioned, `flock`-guarded, with lease expiry.
   Deliberately deferred: it introduces a second source of truth and stale-lease
   failure modes. See ADR 0005.
2. **Effective status vs. `in-review`.** Under the `pr-review` workflow a task can
   be `in-review` in a merged-away branch's worktree after the primary already
   marked it `completed`. The ladder handles this (completed wins), but should
   `next --explain` call out excluded `in-review` tasks as "awaiting review"
   rather than silently hiding them? Leaning yes, via `--explain` text only.
3. **Provenance naming.** Worktree root basename is usually meaningful
   (`agent-b`, `taskmd-2`) but not guaranteed unique. Branch name is the fallback
   discriminator. Good enough for v1?
4. **`worklogs` across worktrees.** Worklogs are per-worktree files under
   `.worklogs/`. Should `get` aggregate sibling worklogs read-only? Deferred —
   nothing coordinates on worklogs.
