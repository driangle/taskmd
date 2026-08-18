# add-task — improvement ideas

Improvements suggested by the [baseline run](REPORT.md) (2026-08-15, `claude-sonnet-5`,
80 samples). Each one is tied to an observed failure, not a hunch.

## Already applied

**Lite skill: use project templates.** Every `lite-skill` bug report came out with no
reproduction steps, because the skill wrote all tasks from one fixed
Objective/Tasks/Acceptance-Criteria body. It now reads `.taskmd/templates/*.md` and fills the
template's own sections. `add-bug-template`: 0/3 → 5/5. Shipped in `9d24f18`.

## Recommended, highest payoff first

### 1. Make `taskmd validate` warn on unknown frontmatter keys

**Evidence:** all five `bare-project` `add-dependency` samples wrote `depends_on: ["002"]`
instead of `dependencies`. Confirmed by hand: `taskmd validate` reports such a task as
**valid**. The dependency silently does not exist — `next`, `graph`, and blocking logic would
all quietly ignore it.

Warn on frontmatter keys outside the schema, and suggest the nearest real field when one is
close (`depends_on` → `dependencies`). This converts a silent data-loss bug into a visible one
for *every* author — agent or human — rather than only for agents that loaded a skill.

This is the single most valuable item here: it is the one failure mode where the artifact looks
correct and is not.

### 2. Flag placeholder-only tasks

**Evidence:** `no-skill` failed 3/5 `add-basic` samples by running `taskmd add`, then returning
the scaffold with `<!-- Describe the goal … -->` and `- [ ] TODO` intact. The skill's step 4
exists precisely to prevent this, which means the guarantee currently lives in a document the
agent may not have.

Options, cheapest first:
- `taskmd validate --strict` (or a warning) when a task body still contains template
  placeholders — `<!-- … -->`, a bare `TODO`, `1. ...`
- a one-line hint in `taskmd add`'s success output: *"created <path> — fill in the Objective,
  Tasks, and Acceptance Criteria"*

The eval graders already implement this check (`noPlaceholders` in `workspace/.verify/assert.go`)
and it is cheap to port.

### 3. Say "replace the placeholders" in the init scaffold

**Evidence:** the same 3/5 failure. `tasks/CLAUDE.md` documents the file format but never says
the scaffold must be filled in. One sentence there helps every agent in the project, including
those with no skill loaded — the `no-skill` variant reads that file today.

### 4. Consider trimming the full plugin skill

**Evidence:** `plugin-skill` and `lite-skill` are tied on correctness (100% vs 95%, not
significant at n=5), but lite is meaningfully faster: 14.7s vs 19.6s median, at slightly lower
cost. The gap is the CLI round-trips (`taskmd templates list`, `add`, re-`Read`, `validate`).

Not a defect — the full skill is the more robust design and it never failed. But if latency
matters, there is room to cut steps, and the lite skill is evidence that a file-first approach
reaches the same place.

## Not worth doing

- **Group-routing guidance.** `add-group-routing` passes 5/5 for every variant including
  `bare-project`. Directory conventions are legible from the filesystem; no skill text needed.
- **Extra ID-collision tooling.** One `bare-project` sample reused ID `005`, but `taskmd
  validate` already reports duplicate IDs as an error (verified by hand), so the failure is
  loud rather than silent. Any skill writing files directly must scan existing IDs first — the
  lite skill already instructs this.
