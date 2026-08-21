# ADR 0003 — Plugin versioning: independent semver per plugin

- **Status:** Accepted
- **Date:** 2026-08-21
- **Deciders:** taskmd maintainers
- **Related task:** `01kz3c247` — Unify version numbers across the three plugins

## Context

The repo ships three Claude Code plugins from one marketplace
(`.claude-plugin/marketplace.json`):

| Plugin | Directory | Version |
|--------|-----------|---------|
| `taskmd` | `claude-code-plugin/` | `0.4.1` |
| `taskmd-lite` | `claude-code-plugin-lite/` | `0.1.7` |
| `taskmd-mcp` | `claude-code-plugin-mcp/` | `1.0.0` |

Those three numbers were not a decision. `scripts/release.sh` silently rewrote
`claude-code-plugin`'s manifest to the **repo** version on every release, so that one
plugin was in lockstep by accident of tooling. The other two were bumped by hand,
whenever someone remembered. Nothing documented what `taskmd-mcp` at `1.0.0` was
promising, or why `taskmd-lite` was six minor versions behind its sibling.

Because the marketplace presents the three as one product family, the natural instinct is
to force them into lockstep on the repo version. That instinct is wrong here: the three
plugins have genuinely different surfaces, different audiences, and change at genuinely
different rates. `taskmd-lite` is untouched by a CLI release it does not use.
`taskmd-mcp` is a tool contract that other MCP clients — Claude Desktop, Cursor,
Windsurf — code against, and that contract's stability is exactly what its version number
should be reporting. Lockstep would bump all three on every CLI patch, which tells a
consumer nothing about whether *their* plugin changed.

## Decision

### 1. Each plugin versions independently

Every plugin has its own semver line. A plugin's version moves when **that plugin's**
directory changes, and not otherwise. A taskmd CLI release does not bump any plugin.

The version lines and what each guarantees:

**`taskmd` — `0.x`, pre-1.0.** Slash-command skills that orchestrate the `taskmd` CLI.
Tracks CLI capability: new skills and new flags on existing skills are a patch bump;
renaming or removing a skill, or changing the arguments one accepts, is a minor bump.
Pre-1.0 this surface is still moving, so skill names are not yet a stability promise.

**`taskmd-lite` — `0.x`, pre-1.0.** The CLI-free plugin, whose skills re-express core
taskmd algorithms as prose. It is bound to the **spec**, not to the binary, so it moves
when the spec or the prose moves — see
[ADR 0002](0002-lite-plugin-prose-conformance.md). Same bump rules as `taskmd`.

**`taskmd-mcp` — `1.x`, stable.** The MCP tool surface — tool names, argument schemas,
and result shapes — is a compatibility contract that non-Claude-Code clients code
against. That contract is stable, which is what `1.0.0` says and why it is not being
renumbered downward. Adding a tool or an optional argument is a minor bump; a bug fix in
an existing tool is a patch bump; removing a tool, renaming one, or changing an existing
argument or result shape is a **major** bump.

The `taskmd` plugin currently sitting at `0.4.1` — the same number as the repo — is now a
coincidence of history, not a rule. It will drift from the repo version at the next
release that touches either one alone, and that is expected.

### 2. `plugin.json` is the single source of truth

Each plugin's version lives in exactly one place: its
`<plugin-dir>/.claude-plugin/plugin.json`.

`.claude-plugin/marketplace.json` deliberately carries **no** `version` fields on its
plugin entries. A version mirrored into the marketplace manifest is a second copy with no
mechanism keeping it honest, and duplicated version numbers drift — that is how the pin
in `apps/cli/go.mod` went stale twice. `scripts/release.sh` enforces this: it fails the
release if any marketplace entry grows a `version` key.

### 3. The release script enforces the policy

`scripts/release.sh` no longer propagates the repo version into any plugin manifest.
Instead it mirrors the design already used for the independently-versioned `sdk/go`
module:

- For each plugin, it checks whether the plugin's directory changed since the last repo
  tag.
- **Unchanged** → nothing to do; a version flag passed anyway is warned about and ignored.
- **Changed, no flag** → the release **fails**, showing the diffstat and asking for the
  bump. A changed plugin shipping under its old version is the failure mode this exists
  to prevent.
- **Changed, with `--plugin-taskmd-version` / `--plugin-lite-version` /
  `--plugin-mcp-version`** → the manifest is rewritten, after checking the value is valid
  semver and strictly greater than the current one.

The bump size is a human judgment call against the guarantees above; a script cannot
detect that a tool's result shape changed meaning. The release skill
(`.agents/skills/release/SKILL.md`) prompts for that judgment before invoking the script.

## Consequences

- **A release can now fail on a plugin change.** That is the point: the alternative is
  shipping a modified plugin under a version consumers already have cached.
- **Three numbers, three stories.** A consumer reading `taskmd-mcp 1.2.0` learns that the
  MCP tool surface gained something and broke nothing. Under lockstep that number would
  have carried no information at all.
- **`taskmd-mcp` stays at `1.x` while the CLI is pre-1.0.** This looks inconsistent at a
  glance and is intentional: the MCP tool surface is a narrower, older, and genuinely
  more stable contract than the CLI as a whole.
- **Version bumps are no longer automatic.** Each plugin README documents its own line so
  the guarantee is discoverable from the plugin, not just from this ADR.
