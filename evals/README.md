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

## Running

```bash
cd evals/add-task
skival validate suite.yaml                       # structure check, no cost
skival run suite.yaml                            # full suite (4 evals x 3 variants x 5 samples)
skival run suite.yaml --samples 1                # cheap smoke run
skival run suite.yaml --evals add-basic --variants plugin-skill --samples 1
skival run suite.yaml --results-dir ./results    # save for `skival report` / `skival compare`
```

A full run is 60 samples. Samples are set to 5 rather than 3 because at least one variant
behaves bimodally — 3 samples land anywhere between 0/3 and 3/3.

Requires `taskmd` and `go` on `PATH`. The `taskmd` used is whatever is installed, *not*
`taskmd-dev`, so evals measure the skill against a released CLI.

## How a suite is put together

**`workspace/`** is the fixture project every sample starts from: `taskmd init` output plus
five baseline tasks (`001`–`005`, mixed statuses/priorities/types) and `src/app.go` with TODO
comments. Task content is carried over from `benchmark/fixtures/` so results stay comparable.

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

Newer skival turns this from a manual audit into a hard assertion: `tool_not_used` fails any
sample that touched a forbidden tool, and upstream has since added default-deny tool enforcement
and a per-variant tool census in the report. **The `skival` on this machine predates all of it** —
`skival validate` rejects `tool_not_used` with `field tools not found`. Run `make install` in the
skival checkout before relying on it, and until then keep auditing by hand.

Audit a run's conversations before trusting it; anything outside the variant's `allowed_tools`
is a leak:

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

### Adding a check

1. Add a function to `workspace/.verify/checks.go` and register it in the `checks` map.
2. Reference it from `suite.yaml` as `cd .verify && GOWORK=off go run . <name>`.
3. Test both directions by hand before spending tokens on a run:

   ```bash
   cd evals/add-task/workspace
   taskmd add "..." --priority high            # build the state you expect
   (cd .verify && GOWORK=off go run . my-check)  # expect PASS
   ```

   Then break one thing and confirm it fails. A check that can't fail measures nothing.

## Results

The current baseline lives in [`add-task/REPORT.md`](add-task/REPORT.md) — pass rates, cost,
latency, findings, and skival's verbatim output. Headline from the 2026-08-15 run
(`claude-sonnet-5`, 5 samples, 80 total): `plugin-skill` 100%, `lite-skill` 95%,
`no-skill` 75%, `bare-project` 45%.

Store each suite's committed baseline as `<suite>/REPORT.md`; raw run output under
`<suite>/results/` is gitignored and regenerable.

### Reading a failure

**When a variant fails, read the workdir before believing the number** — graders have bugs too.
Every run's report lists a temp workdir per sample, and the grader can be re-run there:

```bash
cd /var/folders/.../skival-isolate-399124493/.verify
GOWORK=off go run . add-dependency
```
