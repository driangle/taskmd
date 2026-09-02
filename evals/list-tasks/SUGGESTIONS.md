# list-tasks — improvement suggestions

Grounded in the failures observed in [run 1](reports/2026-09-02-565f740.md) (160 samples). Each
item names the failure it fixes and how many samples it would have flipped. Nothing here is
speculative polish — if a change is not tied to an observed failure, it is not listed.

**Status after applying items 1–4 in `7dc4f3f` and re-running
([run 2](reports/2026-09-02-95e136c.md)):**

| Item | Applied | Verified by run 2 |
|---|---|---|
| 1 — preserve requested output format | yes | **Confirmed.** 0/8 → 8/8 on the JSON eval; `plugin-skill` 77.5% → 100%, zero failures |
| 2 — drop the `core-cli` literal | yes | **Confirmed.** No sample copied a literal example; `list-scope-filter` 7/8 → 8/8 |
| 3 — widen `description:` | yes | **Not verifiable here** — skills are injected as a system prompt, so `description:` never decides loading |
| 4 — status/priority are enums | yes | **No measurable effect.** `lite-skill` was already 8/8 on the status eval; it dipped elsewhere, and an A/B showed the dip was *not* caused by this edit |
| 5 — lite skill shelling out | no, by recommendation | — |
| 6 — suite gaps | no | Deliberately deferred: adding evals changes the denominator and breaks comparability with run 1 |

---

## 1. The plugin skill must preserve the requested output format — 8 samples

**Failure:** `list-json-format`, `plugin-skill`, **0/8**. The agent runs
`taskmd list --priority high --format json`, receives correct JSON, and then re-renders it as a
markdown table, closing with *"Raw JSON is above in the tool output if you need to consume it
directly"* — which is false, because the user does not see tool output.

**Cause:** the skill's whole instruction for this is step 2, *"Present the output to the
user."* Presenting is exactly what the agent does — it reformats for readability, which is the
right default for a table and the wrong one for a machine-readable format the user asked for by
name.

**Fix** — in `claude-code-plugin/skills/list-tasks/SKILL.md`, replace step 2:

```markdown
2. Present the output to the user
   - **If the user asked for a specific format (`json`, `yaml`, csv, "raw"), reproduce the
     command's output verbatim in a fenced code block.** Do not re-render it as a table and do
     not summarize it — the format *is* the request. Tool output is not visible to the user, so
     "the JSON is above" is never true.
   - Otherwise, present the table as-is.
```

This is the highest-value change available: it alone takes `plugin-skill` from 77.5% to a
projected ~97%, and moves it from *worse than no skill* to *best-in-class on correctness at a
third of the latency*.

**Worth checking beyond this suite:** every skill that shells out to a formatting-capable
command has the same shape. `get-task`, `next-task` and `validate-tasks` all say some version
of "present the output" and all accept `--format`. This suite only measured `list-tasks`.

## 2. Stop shipping a repo-specific phase ID as a literal example — 4 samples, wasted work

**Observation (not a failure — all four still passed):** `list-scope-filter`, `plugin-skill`.
Asked *"what's on the plate for the CLI?"*, 4 of 8 samples opened with a command lifted verbatim
from the skill file — `taskmd list --phase core-cli` (3) or `taskmd list --scope cli` (1).
Neither exists in the fixture: `core-cli` is a phase from **taskmd's own repo**, and the fixture
configures no `scopes:`. All four got an empty result, fell back to `taskmd list`, and answered
correctly. The cost is a wasted round-trip, not a wrong answer.

**Cause:** line 19 of `claude-code-plugin/skills/list-tasks/SKILL.md`:

```markdown
   - Phase filtering: `taskmd list --phase core-cli` or `taskmd list --filter phase=core-cli`
```

`core-cli` reads as a suggested value rather than as one project's arbitrary ID, and the "cli"
in it collides with any request mentioning a CLI. The evidence is narrow and specific: it fired
on 4/8 of the one eval whose prompt contains "CLI" and **0/32 everywhere else**, including
`list-phase-filter`, where the prompt names the phase outright.

**Fix** — use an obvious placeholder, and point at the discovery command that already exists:

```markdown
   - Phase filtering: `taskmd list --phase <phase-id>` (or `--filter phase=<phase-id>`).
     Phase IDs are per-project — run `taskmd phases` to see this project's before filtering.
     Never guess an ID.
```

Cheap, and pure upside: `taskmd phases` is already documented on the next line, and 5 of the 8
samples ran it unprompted.

**Separately, the one `list-scope-filter` failure has a different cause** and this fix would not
have prevented it: sample 7 ran `taskmd list --status pending` and reported the pending tasks —
it read "what's on the plate?" as a status question and never considered the group. The skill
lists the available flags but says nothing about mapping a request onto them, so this is the
weakest-evidenced item here (n=1) and not worth acting on alone. Watch whether it recurs.

## 3. Bind the skill to the project, not to the assistant's memory — 14 samples

**Failure:** 14 of 31 failures — every one in `no-skill` and `bare-project` — answered a task
question by running `cat ~/.claude/projects/<workdir>/memory/MEMORY.md`, finding nothing, and
asking the user where their tasks live. None of them ever ran `ls`.

**Nothing needs fixing in either skill** — both already score 0 on this mode, and this is the
measured value of shipping a skill at all. The actionable part is **keeping** it, and the risk
lives in the `description:` field, because that is what decides whether the skill loads in a
real session. Both plugins currently use:

> `List tasks with optional filters. Use when the user wants to see their tasks.`

The observed misfires were on *"which of my tasks are still pending?"* and *"list my high
priority tasks"* — phrasings a description built around the word "tasks" may not reliably win
against Claude's own memory/todo behavior. Widen it to the vocabulary that actually failed:

```yaml
description: >
  List tasks from the project's taskmd files, with optional filters (status, priority,
  phase, group, owner). Use whenever the user asks what is pending, in progress, done,
  outstanding, assigned, high priority, or "on the plate" — including phrasings like
  "my tasks", "my todos" or "what's left", which refer to the project's task files and
  not to conversation memory.
```

This suite cannot verify that change — skills are injected as a system prompt here, so
`description:` is never exercised. It needs a triggering eval, which skival does not currently
model. Flagging it rather than claiming it fixed.

## 4. State that status and priority are enums, not adjectives — 9 samples

**Failure:** modes C and D, all in `bare-project` — it counted `in-progress` as pending (3) and
`critical` as high priority (5). (The fourth mode-D sample is `plugin-skill`'s `list-scope-filter`
miss, which is a wrong-filter slip rather than an enum misreading; see item 2.)

The plugin skill gets this right for free by delegating to the CLI. The **lite skill parses
frontmatter itself** and is one prompt away from the same mistake — it scored 8/8 on
`list-status-filter` here, but nothing in its instructions says these are closed value sets. Add
to step 4 of `claude-code-plugin-lite/skills/list-tasks/SKILL.md`:

```markdown
   Status and priority are **exact enum values, not descriptions**. `--status pending` means
   `status: pending` only — it does not include `in-progress`. `--priority high` does not
   include `critical`. When the user's wording is broader than the enum ("what's not done",
   "what's urgent"), say which statuses or priorities you included.
```

That last clause matters: several correct answers in this run explained their filter, and being
explicit is what makes an over-broad reading visible instead of silent.

## 5. Consider whether the lite skill should shell out when the CLI exists — cost, not correctness

**Observation, not a failure:** `lite-skill` is 100% correct and the most expensive way to get
there. Across 40 samples it made 241 `Read` and 57 `Glob` calls, against 49 `Bash` calls total
for `plugin-skill` — median 1,134 tokens versus 283, $0.154 versus $0.102, 17.2s versus 9.4s.
It costs the same as having no skill at all.

Being CLI-free is the lite plugin's whole premise, so this is a deliberate trade and possibly
the right one. But the premise is *"works without the CLI"*, which does not have to mean
*"never uses the CLI"*. A one-line preamble — "if `taskmd` is on `PATH`, run
`taskmd list $ARGUMENTS` and present it; otherwise scan the files as below" — would keep the
guarantee and recover the cost, at the price of two behaviors to maintain instead of one.

**Recommendation: leave it alone unless cost becomes a concern.** 100% correctness is worth
more than $0.05 a call, and a second code path is a real maintenance cost. Recorded so the
trade is visible rather than accidental.

## 6. Suite gaps to close next

- **An owner filter eval.** The fixture carries `owner` on four tasks (`alex` ×2, `sam`,
  `jordan`) and nothing exercises it. "what is alex working on?" is a natural request and would
  probably surface the same generic-vocabulary weakness as mode A.
- **A skill-triggering eval**, if skival grows support for it. Item 2 is the biggest measured
  effect in this run and the one thing this suite structurally cannot verify, because skills are
  injected as a system prompt rather than selected by `description:`.
- **A sort/limit eval.** `--sort` and `--limit` are documented in the plugin skill and untested
  here; "what are my top 3 priorities" is a common phrasing with a checkable answer.
- **Rephrase `list-scope-filter`, or accept two answers.** Run 2 surfaced samples that read
  "what's on the plate for the CLI?" as *outstanding* CLI work and so omitted the completed task
  `005` — a defensible answer graded wrong. The eval currently measures prompt interpretation as
  much as group filtering. Either name the group explicitly or make the expected set tolerant.
- **Baselines need more samples than skill variants.** `no-skill` moved 17.5 points between two
  identical runs, driven by the bimodal "check my own memory" behavior. Any future comparison
  resting on the baseline should run it twice or raise `samples` for it.
- **Cross-check item 1 against the other read skills.** If "present the output" loses `--format`
  in `list-tasks`, `get-task` and `next-task` are likely identical — a cheap single-eval probe
  each would confirm before rewriting them.
