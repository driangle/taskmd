# get-task — improvement suggestions

Grounded in what [run 1](reports/2026-09-03-3081927.md) actually observed (160 samples). Each
item names the failure it addresses and how many samples it would flip. Nothing here is
speculative polish — if a change is not tied to something the run showed, it is not listed.

**The headline constrains this list.** Both skills scored **40/40**. There is no correctness
failure in either skill to fix, so nothing below claims a correctness flip for a skill edit —
saying otherwise would be inventing a mechanism the data does not support. What the run does
support is one efficiency fix, one suite fix, and one explicit *do not change this*.

| # | Change | Flips | Status |
|---|--------|-------|--------|
| 1 | `--exact` in the plugin skill's lookup step | 0 correctness, 8 wasted round trips | proposed |
| 2 | Suite: 4 of 5 evals are at the ceiling | — (measurement capability) | proposed |
| 3 | Leave both skills' instructions alone | — | recommendation |

---

## 1. The plugin skill should use `--exact` and handle the miss itself — 0 samples, 8 round trips

**Observed:** `get-missing`, `plugin-skill`, 8/8 **passed** — but all 8 got there the hard way.
Every sample ran `taskmd get 042`, and every sample received:

```
No exact match found for "042". Did you mean:

  1. 002: Add full-text search (67% match)

Enter selection (1-1), or 0 to cancel: Error: invalid selection
```

`taskmd get` fuzzy-matches, opens an **interactive selection prompt**, and with no tty exits 1.
Each sample then recovered by running `taskmd list`. Correctness held; the cost did not — that
is one failed command and one extra tool round trip on every miss, in the one command the skill
tells the agent to run.

It needs a *short* query to fire: `taskmd get "SSO login bug"` falls below the 0.6 fuzzy
threshold and returns a clean `task not found`, which is why `get-by-keyword` never hit it. IDs
and single words are exactly the queries this skill is for.

**Fix** — in `claude-code-plugin/skills/get-task/SKILL.md`, steps 1–2:

```markdown
1. Run `taskmd get $ARGUMENTS --exact` to look up the task.
   Without `--exact`, `taskmd get` fuzzy-matches and opens an *interactive* selection
   prompt — which has no tty here, so it exits non-zero with "invalid selection" rather
   than answering. `--exact` fails cleanly with `task not found`.
2. If the task is not found, run `taskmd list` to show the available tasks. Say plainly
   that the requested task does not exist before naming any near match, and never present
   a near match as though it were the task that was asked for.
```

The second sentence of step 2 costs nothing and pins the behavior every variant already got
right, so a future skill edit cannot silently lose it.

**Not applied to the lite skill** — it never runs the CLI, and its `Glob`-then-read path
handled `get-missing` 8/8 with no wasted calls.

## 2. Four of five evals cannot detect a regression — suite change

**Observed:** `get-by-id`, `get-missing`, `get-json-format` and `get-blocked-state` are 8/8 for
**all four variants**, including `bare-project`, which has neither the CLI nor the docs. Only
`get-by-keyword` separates anything. As it stands, a change that broke the plugin skill's ID
lookup or its JSON handling would very likely still show 100%.

This is not an argument to delete those evals — a ceiling result is a real finding, and
`get-json-format` sits at the ceiling here precisely because it caught a regression in
`list-tasks`, so it earns its place as a guard. The gap is that nothing else discriminates.

What the run suggests would discriminate, each grounded in a behavior a variant actually got
wrong or nearly wrong:

- **A genuinely ambiguous keyword.** The fixture's 001/005 auth pair exists for this and is
  currently unused: `get-by-keyword` asks about "the SSO login bug", which only 001 matches.
  An eval asking for *"the auth task"* would grade whether the skill surfaces both and asks,
  rather than silently picking one. This is the fixture's stated purpose
  ([fixtures/README.md](../fixtures/README.md)).
- **A task nested in a group, asked for by title.** `no-skill`'s single `get-blocked-state`
  failure was not a reasoning failure — it globbed the root of `tasks/` and never descended
  into `tasks/cli/` or `tasks/web/`, then reported *"there's no task 003"*. One sample is not
  a result, but it is a reachable failure that no current eval targets.

**Deliberately deferred, not forgotten:** adding evals changes the denominator and breaks
comparability with run 1. Do it in one batch, and treat the next run as a new baseline rather
than a comparison — the same call [`list-tasks` made](../list-tasks/SUGGESTIONS.md).

## 3. Do not "improve" either skill's instructions — recommendation

Both are 40/40. The plugin skill is four steps and is simultaneously the **cheapest**
($0.1115 median vs `no-skill`'s $0.1354) and **fastest** (12.5s vs 17.6s) variant in the suite,
with the leanest tool census (Bash ×81 / Read ×22 / Glob ×8 across 40 samples, against
`no-skill`'s Bash ×106). Brevity is doing work here: it points at one command instead of
prompting exploration.

Item 1 is a targeted two-line correctness-of-mechanism fix. Anything beyond that has no
observed failure behind it, and `list-tasks` run 2 is the cautionary case — an edit made on
reasoning rather than evidence (item 4 there) produced no measurable effect and briefly looked
like a regression.

## What is explicitly *not* actionable here

The suite's largest signal — 16 failures on `get-by-keyword`, all of them the agent answering
from `~/.claude/.../memory/MEMORY.md` instead of the project — occurs **only in the two
variants that have no skill file**. Both skills already eliminate it, 8/8. There is no skill
edit that flips those samples, because there is no skill in them.

It belongs in the report as evidence that these skills earn their cost, not here as a task.
