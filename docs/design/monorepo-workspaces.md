# Monorepo Workspaces

Support multiple independent task folders in a monorepo, where each workspace has its own disjoint set of tasks.

## Problem

Today taskmd has **one task directory per config** (`dir: ./tasks`). In a monorepo you'd have to either:

- Dump all tasks into one folder and use groups/scopes to organize
- Run `taskmd --task-dir packages/foo/tasks` manually each time

Neither is great for truly disjoint task sets.

## Design Options

### Option A: Multiple `dir` entries in one config

```yaml
# .taskmd.yaml at monorepo root
dirs:
  - path: ./packages/api/tasks
    name: api
  - path: ./packages/web/tasks
    name: web
  - path: ./packages/shared/tasks
    name: shared
```

- Single config file, multiple scan roots
- Each becomes a top-level namespace
- Commands default to scanning all, but accept `--workspace api` to scope
- Tasks across workspaces are **disconnected** — no cross-workspace dependencies

**Pros:** Simple mental model, one config to rule them all
**Cons:** Couples unrelated teams/packages, config becomes a bottleneck

### Option B: Per-package `.taskmd.yaml` (distributed config)

Each package has its own `.taskmd.yaml`:

```
monorepo/
├── .taskmd.yaml          # optional root config (workspace discovery)
├── packages/
│   ├── api/
│   │   ├── .taskmd.yaml  # dir: ./tasks, own phases, own scopes
│   │   └── tasks/
│   ├── web/
│   │   ├── .taskmd.yaml
│   │   └── tasks/
```

- `taskmd list` in `packages/api/` only sees API tasks (already works today via config walk-up)
- Root `.taskmd.yaml` could optionally declare workspaces for aggregate views:

```yaml
# root .taskmd.yaml
workspaces:
  - packages/api
  - packages/web
  - packages/shared
```

- `taskmd list --all` from root scans all workspaces
- Each workspace is fully independent (own ID strategy, phases, scopes)

**Pros:** Each team/package is autonomous, scales naturally, works with existing config walk-up
**Cons:** Aggregate views need explicit opt-in, harder to get a unified dashboard

### Option C: Workspace globs (like pnpm/npm workspaces)

```yaml
# .taskmd.yaml
workspaces:
  - "packages/*/tasks"
  - "apps/*/tasks"
```

Auto-discovers task directories by glob. Each discovered directory becomes its own workspace, inheriting defaults from root but overridable with a local `.taskmd.yaml`.

**Pros:** Zero-config for new packages, familiar pattern from JS ecosystem
**Cons:** Magic discovery can be surprising, harder to reason about config inheritance

## Open Design Questions

1. **Cross-workspace dependencies** — should they be allowed? If workspaces are "totally disjoint," probably not — but then `taskmd graph` across the whole monorepo is just N separate graphs.

2. **ID uniqueness scope** — are IDs unique per-workspace or globally? Per-workspace is simpler but means you can't reference `task-042` unambiguously from the root. Could namespace as `api:042`.

3. **Config inheritance** — should a per-package config inherit from the root? (e.g., root sets `id.strategy: ulid`, packages inherit unless overridden)

4. **CLI ergonomics** — what does `taskmd list` show when you're at the monorepo root? All tasks? Nothing? Require `--workspace`?

5. **Web dashboard** — does it show all workspaces in one view with a workspace switcher, or is each workspace a separate instance?

## Recommendation

**Option B (distributed config)** feels most natural because:

- It already half-works today (config walks up to nearest `.taskmd.yaml`)
- It respects package autonomy — each team owns their task config
- An optional root `workspaces:` field enables aggregate views without coupling
- Minimal new concepts: a "workspace" is just "a directory with a `.taskmd.yaml`"
- Matches how tools like turborepo, nx, and pnpm handle monorepos

### Implementation sketch

1. Add optional `workspaces` field to root config
2. Teach the scanner to iterate over workspace roots
3. Namespace task references in aggregate mode (`api:042`)
4. Keep everything else (phases, scopes, IDs) per-workspace
