# Skill evals

Skill benchmarks built on [skival](https://github.com/driangle/skival). This replaces the
bash + LLM-graded harness in `benchmark/`: same question ("does the skill actually help?"),
but with deterministic grading, per-sample isolation, and cost/latency/variance reporting.

One directory per skill, each with its own `suite.yaml`. Each suite keeps one eval per
behavior — when a single eval asserts two things, a failure in one masks the other (see
`add-bug-template` vs `add-group-routing`).

| Suite | Skill | Status |
|-------|-------|--------|
| [`add-task/`](add-task) | `add-task` | trial suite — [report](add-task/REPORT.md) |
| [`list-tasks/`](list-tasks) | `list-tasks` | first read-only suite — [report](list-tasks/REPORT.md) |
| [`get-task/`](get-task) | `get-task` | read-only, single-task lookup — [report](get-task/REPORT.md) |

## Running

```bash
cd evals/add-task
skival validate suite.yaml                       # structure check, no cost
skival run suite.yaml                            # full suite
skival run suite.yaml --samples 1                # cheap smoke run
skival run suite.yaml --evals add-basic --variants plugin-skill --samples 1
skival run suite.yaml --results-dir ./results    # save for `skival report` / `skival compare`
skival run suite.yaml --max-cost 40              # abort if the run overruns its budget
```

**Always `--samples 1` first.** A smoke run is ~5% of the cost of a full one and is where
harness and grader bugs surface; both suites' graders were wrong in ways only a real agent
answer exposed. Read the failures before paying for the full matrix.

Sample counts are per suite: `add-task` uses 5 (3 lands anywhere between 0/3 and 3/3 on a
bimodal variant), `list-tasks` uses 8 (its variants cluster near the ceiling, so separating
them needs narrower intervals).

Requires `taskmd` and `go` on `PATH`. The `taskmd` used is whatever is installed, *not*
`taskmd-dev`, so evals measure the skill against a released CLI.

## How a suite is put together

**`workspace/`** is the fixture project every sample starts from: `taskmd init` output plus
baseline tasks (mixed statuses/priorities/types) and `src/app.go` with TODO comments.

New suites fork [`fixtures/`](fixtures) — the shared base, six tasks carrying `phase`, `owner` and
a real dependency edge, with its verified ground truth recorded in
[`fixtures/README.md`](fixtures/README.md). **`add-task/workspace/` is frozen and stays on its own
five-task copy**: its graders hardcode `baselineIDs` as `001`–`005`, and its committed
[REPORT.md](add-task/REPORT.md) baseline was measured against that exact fixture, so adding a task
there would both break the graders and invalidate the numbers.

The fixture tasks are **grouped** — `tasks/cli/`, `tasks/web/`, and two cross-cutting tasks at
the root. That is deliberate: an eval that asserts "route this to `cli`" against a flat project
measures whether the skill hardcodes taskmd's own taxonomy, not whether it reads the project.
With groups present, the convention is discoverable and both CLI-driven and file-writing skills
can satisfy it.

Each eval sets `isolate: true`, so skival copies `workspace/` to a temp dir per sample. Samples
never see each other's writes and the checked-in fixture is never mutated — no reset hook needed.

**Variants** hold everything constant except the skill file:

| Variant | Skill | Project |
|---------|-------|---------|
| `no-skill` | none | full taskmd project (the ranking baseline) |
| `plugin-skill` | `claude-code-plugin/skills/<skill>/SKILL.md` | full taskmd project |
| `lite-skill` | `claude-code-plugin-lite/skills/<skill>/SKILL.md` | full taskmd project |
| `bare-project` | none | `workspace-bare/` — no `.taskmd.yaml`, no `.taskmd/`, no
  `tasks/CLAUDE.md`, no `TASKMD_SPEC.md`, and a `.shadow/taskmd` stub first on `PATH` |

The first three get the same model and the same tool list, so any delta between them is
attributable to the skill file. `bare-project` additionally strips taskmd itself, separating
"what the skill teaches" from "what `taskmd init` already documents".

### Hermeticity

`allowed_tools` is passed through as `--allowedTools` and **does not gate built-ins** — an agent
can still reach the `Skill` tool (loading plugin skills installed in your own `~/.claude`) or
`TaskCreate` (Claude Code's task tracker). Either one makes every variant run the same
configuration, and neither is visible in skival's report. Every suite must pin
`disallowed_tools` as well as `allowed_tools`.

skival turns this from a manual audit into a hard assertion: `tool_not_used` fails any sample
that touched a forbidden tool, and the report now carries a per-variant tool census. This
**does** work on the installed build (verified 2026-09-01 against `skival` commit `d896bc2`;
an earlier note here claimed `skival validate` rejects `tool_not_used`, which is no longer
true). Add it to every eval — `list-tasks/suite.yaml` shows the shape:

```yaml
- type: tool_not_used
  tools: ["Skill", "TaskCreate", "TaskUpdate", "TaskList", "TaskGet", "ToolSearch"]
```

Still audit a run's conversations before trusting it — `tool_not_used` only catches the tools
you thought to name. Anything outside the variant's `allowed_tools` is a leak:

```bash
python3 -c "
import json,glob,collections
c=collections.Counter()
for f in glob.glob('results/<run>/evals/*/*/*.conversation.jsonl'):
    for line in open(f):
        for b in ((json.loads(line).get('message') or {}).get('content') or []):
            if isinstance(b,dict) and b.get('type')=='tool_use': c[b['name']]+=1
print(dict(c))"
```

### Verifier types

The `add-task` suite uses only two verifiers, which has repeatedly misled people into thinking
those are the only ones. skival registers ten (`internal/verifier/pipeline.go`). The ones that
matter for these suites:

| Type | Required | Sees | Use for |
|------|----------|------|---------|
| `agent_exits_ok` | — | exit status | every eval, as a floor |
| `check` | `run` | the workspace after the run | skills that mutate task files |
| `check_output` | `run` | agent's final text on **stdin** | read-only skills (`list`, `get`, `next`, `status`) |
| `output_contains` | `values` | agent's final text | cheap presence-only smoke assertions |
| `file_contains` | `path` | one file | narrow single-file assertions |
| `tool_not_used` | `tools` | the conversation | hard hermeticity backstop (see below) |
| `judge` | `criteria` | tool activity + response | avoid — LLM grading is what `benchmark/` did badly |

`check_output` is the key one for read-only skills: same stdlib-only Go grader as `check`, reading
`os.Stdin` instead of the task file. Two rules keep those graders honest:

- **Assert two-sided.** The expected IDs must appear *and* the competing ones must not.
  `output_contains` is presence-only — there is no `output_not_contains` — so an agent that dumps
  every task passes a naive "pending only" assertion. Two-sided logic belongs in `check_output`.
- **Match stable tokens, not prose.** Task IDs and field names are matchable; sentence phrasing is
  not. Matching the plugin skill's pretty-printed layout grades formatting compliance and unfairly
  fails `no-skill`.
- **Separate results from commentary by structure, not by keyword.** Agents routinely end a
  correct filtered answer by naming what they left out — *"The other two tasks (006 export CSV,
  004 update README) are in the Polish phase, not MVP"* — and a two-sided check reads those
  names as leaked results. Blocklisting cue words like "excluded" loses: that sentence has none.
  What works is treating layout as the signal — when an answer contains list items (lines
  *starting* with an ID: table rows, bullets, headings), count only those and treat prose as
  commentary; fall back to proximity matching only when there is no list at all. See
  `list-tasks/workspace/.verify/output.go`.

**Verification** is deterministic. Each eval runs `agent_exits_ok` plus a Go check:

```yaml
- type: check
  run: "cd .verify && GOWORK=off go run . add-basic"
```

`workspace/.verify/` is a self-contained Go module (stdlib only) that shells out to
`taskmd list --format json` and `taskmd validate`, then asserts on the created task:
frontmatter fields, tags, dependencies, absence of template placeholders, and that the
Objective / Tasks / Acceptance Criteria sections carry real content.

`GOWORK=off` is required because the module lives inside a repo that has a `go.work`; the
verifier chdirs from `.verify` up to the project root itself.

### Writing a verifier command

**Start the command with an absolute path.** skival resolves a *relative* first token in `run:`
against the **suite** directory. The bare form above happens to work for `check` steps, but in a
`check_output` step `cd .verify && …` becomes `<suite-dir>/cd` and every sample dies with
`sh: …/cd: No such file or directory`, exit 127 — which reads like a broken grader, not a
broken command. Prefix `/bin/sh -c` and the whole string is passed through untouched:

```yaml
- type: check_output
  run: "/bin/sh -c 'cd .verify && GOWORK=off go run . list-all'"
```

That form is also correct on a skival that does no rewriting, so it is safe either way.
`${SKIVAL_WORK_DIR}` and `${SKIVAL_SUITE_DIR}` expand in `run:` if you need explicit paths.

What a verifier command can rely on, confirmed by probe against `d896bc2`: the working
directory is the **per-sample workdir** (the isolated copy), and for `check_output` the agent's
final text arrives on **stdin**. Arguments after the first token survive intact.

### Adding a check

1. Add a function to `workspace/.verify/checks.go` and register it in the `checks` map.
2. Reference it from `suite.yaml` as `cd .verify && GOWORK=off go run . <name>`.
3. Test both directions by hand before spending tokens on a run:

   ```bash
   cp -R evals/add-task/workspace /tmp/ws && cd /tmp/ws   # NOT in-place — see below
   taskmd add "..." --priority high            # build the state you expect
   (cd .verify && GOWORK=off go run . my-check)  # expect PASS
   ```

   Then break one thing and confirm it fails. A check that can't fail measures nothing.

   **Test from a copy outside this repo.** Run `taskmd list` inside `evals/*/workspace/` while
   it sits in the taskmd checkout and it returns taskmd's own 120+ tasks — project resolution
   walks up to the repo — even with an explicit `-d tasks`. (`taskmd validate` is unaffected,
   which makes it worse: the two disagree and only one is lying.) Real runs are immune because
   `isolate: true` copies to a temp dir, so this only ever bites hand-testing.

   For output graders, prefer a Go test over a shell loop: `list-tasks/workspace/.verify/`
   ships an `output_test.go` that feeds recorded agent answers through every grader in both
   directions via `cd .verify && GOWORK=off go test ./...`, at zero token cost. Both bugs the
   `list-tasks` pilot found are pinned there as regression cases.

## Results

Each suite's committed baseline lives in `<suite>/REPORT.md` — pass rates, cost, latency and
findings.

[`add-task/REPORT.md`](add-task/REPORT.md), 2026-08-15 (`claude-sonnet-5`, 5 samples, 80 total):
`plugin-skill` 100%, `lite-skill` 95%, `no-skill` 75%, `bare-project` 45%.

[`list-tasks/REPORT.md`](list-tasks/REPORT.md), 2026-09-02 (`claude-sonnet-5`, 8 samples, 160
total): `lite-skill` 100%, `no-skill` 80%, `plugin-skill` 77.5%, `bare-project` 65%. Two results
there are worth carrying over to other suites:

- **A skill's value can hide behind prompt phrasing.** `no-skill` scores 8/8 on "show me all my
  tasks" and 2/8 on "which of my tasks are still pending?" — in the *same* workspace. All 14
  failures of that kind ran one command, `cat ~/.claude/projects/<workdir>/memory/MEMORY.md`,
  and never looked at the project. Generic assistant vocabulary ("pending", "high priority")
  resolves against Claude's own memory; project vocabulary ("the mvp phase") does not. An eval
  set that only asks the obvious question measures nothing — which is how the retired
  `benchmark/` harness concluded this skill had 0% effect.
- **A skill can make an agent worse.** `plugin-skill` scores 0/8 on "as JSON": it fetches
  correct JSON and re-renders it as a markdown table. Its pass rate excluding that one eval is
  97% against `no-skill`'s 81%, so a headline number alone would have mis-ranked it in both
  directions.

### One run, one commit

A skill eval measures the skill files *as of some commit*. A pass rate without the revision that
produced it cannot be compared against a later one, which makes "did the fix work?" unanswerable
— and editing a skill and re-running is the entire point of the harness. So use the wrapper
rather than calling `skival run` directly:

```bash
evals/run-eval.sh list-tasks --samples 1     # smoke run
evals/run-eval.sh list-tasks                 # full run
```

It validates the suite, warns if the tree is dirty (a dirty run is recorded against a commit it
did not measure), and writes `snapshot.json` into the run directory with the commit, tool
versions and wall clock.

Layout per suite:

- `REPORT.md` — cross-run index: one row per run, with its commit, and the current headline.
- `reports/<date>-<commit>.md` — the durable per-run write-up. Never overwrite one; a superseded
  run is still the only evidence for what the skill did at that revision.
- `results/` — raw skival output. Gitignored and regenerable.

### Reading a failure

**When a variant fails, read the workdir before believing the number** — graders have bugs too.
Every run's report lists a temp workdir per sample, and the grader can be re-run there:

```bash
cd /var/folders/.../skival-isolate-399124493/.verify
GOWORK=off go run . add-dependency
```
