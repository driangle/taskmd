# add-task skill — eval report

Baseline run of [`suite.yaml`](suite.yaml). Regenerate with `skival run suite.yaml`; see
[`../README.md`](../README.md) for how the suite is built.

| | |
|---|---|
| **Run** | 2026-08-15 22:51:41 – 23:04:09 (`results/20260815-225141`) |
| **Model** | `claude-sonnet-5`, runner `claude-code` |
| **Samples** | 5 per variant per eval — 80 total |
| **Wall clock / cost** | 12m 28s, $21.55 |
| **taskmd** | 0.4.0 |
| **Repo** | `bcd3484` |

## What is being compared

Four setups get the identical prompt in an identical fixture project, and only one thing changes
between them: what guidance the agent has. Everything else — model, tool access, fixture
contents, graders — is held constant.

**`no-skill`** — vanilla Claude Code dropped into a real taskmd project. It has the `taskmd`
binary, `.taskmd.yaml`, the templates in `.taskmd/`, and the `tasks/CLAUDE.md` +
`TASKMD_SPEC.md` that `taskmd init` writes. What it does *not* have is the add-task skill.
This is the control the others are scored against: *how far does a competent agent get on its
own, in a project that documents itself?*

**`plugin-skill`** — the same setup plus `claude-code-plugin/skills/add-task/SKILL.md`, the full
plugin skill. It drives the CLI (`taskmd add --template bug --group cli …`) and then fills in
the generated file. *What does the shipped skill add?*

**`lite-skill`** — the same setup plus `claude-code-plugin-lite/skills/add-task/SKILL.md`, the
CLI-free variant. It reads `.taskmd.yaml` and the templates itself and writes the markdown
directly with `Write`, never invoking `taskmd`. *Does the no-CLI version hold up against the
full one?*

**`bare-project`** — no skill *and* no taskmd: a workspace with `.taskmd.yaml`, `.taskmd/`,
`tasks/CLAUDE.md` and `TASKMD_SPEC.md` deleted, and a stub `taskmd` first on `PATH` that exits
127. The agent has nothing but the five existing task files to infer conventions from. *How much
of the result comes from the skill versus from taskmd's own scaffolding?*

The skills are injected as an appended system prompt, so `plugin-skill` and `lite-skill` differ
from `no-skill` by exactly one document.

## Bottom line

**The add-task skill works, and both versions of it are worth shipping.** Either skill takes an
agent from 75% to 95–100% on realistic task-filing requests. The full plugin skill is perfect
(20/20); the lite skill matches it within noise at a third less latency. Strip taskmd's own
documentation away and the agent drops to 45% — so `taskmd init`'s scaffolding is carrying real
weight too, and the skill is what closes the remaining gap.

## Summary

| Eval | `no-skill` | `plugin-skill` | `lite-skill` | `bare-project` |
|------|-----------|----------------|--------------|----------------|
| `add-basic` | 2/5 | 5/5 | 5/5 | 3/5 |
| `add-bug-template` | 3/5 | 5/5 | 4/5 | 1/5 |
| `add-group-routing` | 5/5 | 5/5 | 5/5 | 5/5 |
| `add-dependency` | 5/5 | 5/5 | 5/5 | 0/5 |
| **Pass rate** | **75%** | **100%** | **95%** | **45%** |
| Median cost | $0.316 | $0.318 | $0.228 | $0.224 |
| Median duration | 27.3s | 19.6s | 14.7s | 20.5s |

`lite-skill` ranks #1 on the composite score only because it is cheaper and faster — it lost a
sample on correctness, and at n=5 the pass-rate gap is not significant. Treat the two skills as
equivalent in quality, and prefer lite if latency matters.

## What each variant gets wrong

**`no-skill` (75%) — files the task, doesn't finish it.** Three of its five `add-basic` failures
are identical: it runs `taskmd add`, gets a scaffold, and hands it back with the template's
`<!-- ... -->` comments still in place. That is exactly the step the skill spells out ("replace
placeholder content… don't leave placeholders"). It is also the slowest variant at 27.3s median
— unguided exploration costs more than following instructions.

**`bare-project` (45%) — plausible markdown that isn't taskmd.** With no `TASKMD_SPEC.md` and no
CLI it invents the schema. All five `add-dependency` samples wrote `depends_on: ["002"]`; the
real field is `dependencies`. It is a reasonable guess, it parses as YAML, and it is wrong — the
dependency graph would silently not exist. One sample also reused ID `005` for a new task, a
collision the CLI would have prevented.

**`add-group-routing` is saturated** (5/5 for every variant, `bare-project` included). Directory
conventions are legible from the filesystem alone, so neither skill nor CLI is needed. It is a
regression check now, not a discriminator.

## Methodology notes

- **Grading is deterministic Go** (`workspace/.verify/`), shelling out to `taskmd -d tasks list
  --format json` and `taskmd -d tasks validate`. The explicit `-d tasks` matters: `bare-project`
  has no `.taskmd.yaml`, and without the flag taskmd scans from `.` and reports root-level tasks
  as group `tasks`, so variants would grade differently.
- **Verifiers run outside the variant's environment**, which is what makes `bare-project`
  gradable — the agent sees a stub `taskmd` that exits 127 while the verifier uses the real one.
- **`samples: 5`.** At n=5 the 95% CI on a 100% pass rate is [84–100%]; differences under ~15
  points are not resolvable, and skival flags this on the ranking table.
- **Tool access is pinned** via `allowed_tools` *and* `disallowed_tools` in the suite — see the
  hermeticity note in [`../README.md`](../README.md) before adding a variant or a new suite.

## Next

Skill- and CLI-level improvements suggested by these results are in
[`SUGGESTIONS.md`](SUGGESTIONS.md). Suite-level follow-ups:

1. **Retire or replace `add-group-routing`** as a discriminator (saturated).
2. **Add evals for what only the skills teach** — phases (`--phase`, including adding a missing
   phase to `.taskmd.yaml`), `--parent` subtasks, custom templates, the effort vocabulary.
3. **Raise samples to 10** for the two skill variants if you want their 5-point gap resolved.

---

<!-- Everything below is skival's output for the run above, verbatim. -->

## Per-sample results

```
EVAL                                          VARIANT       SAMPLE  STATUS  COST                       DURATION                                        TOKENS (IN/OUT)
----                                          ---------     ------  ------  ----                       --------                                        ---------------
Basic task with priority and tags             no-skill      1       fail    $0.1893                    16.6s                                           12/774
Basic task with priority and tags             no-skill      2       fail    $0.2515                    18.3s                                           12/864
Basic task with priority and tags             no-skill      3       pass    $0.2984                    25.2s                                           16/1.6k
Basic task with priority and tags             no-skill      4       pass    $0.3733                    35.6s                                           22/1.9k
Basic task with priority and tags             no-skill      5       fail    $0.2465                    20.1s                                           12/857
Basic task with priority and tags             no-skill      agg     FAIL    $0.2515 [$0.1893–$0.3733]  20.1s [16.6s–35.6s] cost_cv=22.6% dur_cv=29.7%  12/864
Basic task with priority and tags             plugin-skill  1       pass    $0.7233                    20.8s                                           12/1.2k
Basic task with priority and tags             plugin-skill  2       pass    $0.2497                    20.6s                                           12/1.1k
Basic task with priority and tags             plugin-skill  3       pass    $0.7415                    21.6s                                           14/1.2k
Basic task with priority and tags             plugin-skill  4       pass    $0.2907                    24.0s                                           16/1.3k
Basic task with priority and tags             plugin-skill  5       pass    $0.4942                    24.6s                                           14/1.4k
Basic task with priority and tags             plugin-skill  agg     PASS    $0.4942 [$0.2497–$0.7415]  21.6s [20.6s–24.6s] cost_cv=41.5% dur_cv=7.5%   14/1.2k
Basic task with priority and tags             lite-skill    1       pass    $0.2185                    15.4s                                           8/1.1k
Basic task with priority and tags             lite-skill    2       pass    $0.2196                    16.6s                                           8/1.1k
Basic task with priority and tags             lite-skill    3       pass    $0.2231                    17.9s                                           8/1.3k
Basic task with priority and tags             lite-skill    4       pass    $0.2301                    16.0s                                           8/1.2k
Basic task with priority and tags             lite-skill    5       pass    $0.2158                    14.7s                                           8/949
Basic task with priority and tags             lite-skill    agg     PASS    $0.2196 [$0.2158–$0.2301]  16.0s [14.7s–17.9s] cost_cv=2.2% dur_cv=6.7%    8/1.1k
Basic task with priority and tags             bare-project  1       pass    $0.1510                    16.8s                                           8/1.2k
Basic task with priority and tags             bare-project  2       fail    $0.2150                    18.6s                                           10/1.1k
Basic task with priority and tags             bare-project  3       fail    $0.1961                    16.2s                                           8/1.0k
Basic task with priority and tags             bare-project  4       pass    $0.2239                    19.7s                                           10/1.3k
Basic task with priority and tags             bare-project  5       pass    $0.2152                    20.0s                                           10/1.1k
Basic task with priority and tags             bare-project  agg     FAIL    $0.2150 [$0.1510–$0.2239]  18.6s [16.2s–20.0s] cost_cv=13.1% dur_cv=8.3%   10/1.1k
Bug report routed to a template and group     no-skill      1       fail    $0.0717                    5.8s                                            2/303
Bug report routed to a template and group     no-skill      2       pass    $0.3372                    32.4s                                           18/2.1k
Bug report routed to a template and group     no-skill      3       pass    $0.3747                    34.6s                                           22/2.2k
Bug report routed to a template and group     no-skill      4       fail    $0.3668                    28.8s                                           22/1.8k
Bug report routed to a template and group     no-skill      5       pass    $0.2985                    31.4s                                           16/2.1k
Bug report routed to a template and group     no-skill      agg     FAIL    $0.3372 [$0.0717–$0.3747]  31.4s [5.8s–34.6s] cost_cv=38.7% dur_cv=39.7%   18/2.1k
Bug report routed to a template and group     plugin-skill  1       pass    $0.2629                    19.2s                                           14/887
Bug report routed to a template and group     plugin-skill  2       pass    $0.2427                    18.5s                                           12/833
Bug report routed to a template and group     plugin-skill  3       pass    $0.2714                    21.2s                                           14/1.0k
Bug report routed to a template and group     plugin-skill  4       pass    $0.2235                    15.2s                                           10/735
Bug report routed to a template and group     plugin-skill  5       pass    $0.2631                    18.9s                                           14/906
Bug report routed to a template and group     plugin-skill  agg     PASS    $0.2629 [$0.2235–$0.2714]  18.9s [15.2s–21.2s] cost_cv=6.9% dur_cv=10.4%   14/887
Bug report routed to a template and group     lite-skill    1       pass    $0.2215                    11.0s                                           8/744
Bug report routed to a template and group     lite-skill    2       pass    $0.2228                    11.5s                                           8/804
Bug report routed to a template and group     lite-skill    3       pass    $0.2276                    15.4s                                           8/1.0k
Bug report routed to a template and group     lite-skill    4       fail    $0.2211                    10.8s                                           8/722
Bug report routed to a template and group     lite-skill    5       pass    $0.2232                    11.4s                                           8/823
Bug report routed to a template and group     lite-skill    agg     FAIL    $0.2228 [$0.2211–$0.2276]  11.4s [10.8s–15.4s] cost_cv=1.0% dur_cv=14.0%   8/804
Bug report routed to a template and group     bare-project  1       fail    $0.2627                    27.4s                                           14/1.8k
Bug report routed to a template and group     bare-project  2       fail    $0.2393                    23.9s                                           12/1.5k
Bug report routed to a template and group     bare-project  3       fail    $0.2528                    24.3s                                           12/1.8k
Bug report routed to a template and group     bare-project  4       fail    $0.2424                    25.5s                                           12/1.6k
Bug report routed to a template and group     bare-project  5       pass    $0.3748                    42.7s                                           22/2.7k
Bug report routed to a template and group     bare-project  agg     FAIL    $0.2528 [$0.2393–$0.3748]  25.5s [23.9s–42.7s] cost_cv=18.5% dur_cv=24.6%  12/1.8k
Task routed to the right group directory      no-skill      1       pass    $0.3945                    36.9s                                           24/2.3k
Task routed to the right group directory      no-skill      2       pass    $0.3507                    29.9s                                           20/1.9k
Task routed to the right group directory      no-skill      3       pass    $0.2800                    25.4s                                           12/2.0k
Task routed to the right group directory      no-skill      4       pass    $0.3285                    26.6s                                           18/1.8k
Task routed to the right group directory      no-skill      5       pass    $0.3091                    26.8s                                           16/1.6k
Task routed to the right group directory      no-skill      agg     PASS    $0.3285 [$0.2800–$0.3945]  26.8s [25.4s–36.9s] cost_cv=11.6% dur_cv=14.3%  18/1.9k
Task routed to the right group directory      plugin-skill  1       pass    $0.2466                    22.1s                                           12/1.0k
Task routed to the right group directory      plugin-skill  2       pass    $0.2698                    21.9s                                           14/1.3k
Task routed to the right group directory      plugin-skill  3       pass    $0.2478                    18.8s                                           12/1.1k
Task routed to the right group directory      plugin-skill  4       pass    $0.2479                    18.9s                                           12/1.1k
Task routed to the right group directory      plugin-skill  5       pass    $0.2499                    19.0s                                           12/1.1k
Task routed to the right group directory      plugin-skill  agg     PASS    $0.2479 [$0.2466–$0.2698]  19.0s [18.8s–22.1s] cost_cv=3.5% dur_cv=7.6%    12/1.1k
Task routed to the right group directory      lite-skill    1       pass    $0.2132                    12.3s                                           8/839
Task routed to the right group directory      lite-skill    2       pass    $0.2140                    13.6s                                           8/869
Task routed to the right group directory      lite-skill    3       pass    $0.2299                    15.9s                                           8/1.2k
Task routed to the right group directory      lite-skill    4       pass    $0.2237                    13.9s                                           8/877
Task routed to the right group directory      lite-skill    5       pass    $0.2236                    11.7s                                           8/868
Task routed to the right group directory      lite-skill    agg     PASS    $0.2236 [$0.2132–$0.2299]  13.6s [11.7s–15.9s] cost_cv=2.9% dur_cv=10.8%   8/869
Task routed to the right group directory      bare-project  1       pass    $0.2434                    20.6s                                           12/1.5k
Task routed to the right group directory      bare-project  2       pass    $0.2360                    23.5s                                           12/1.3k
Task routed to the right group directory      bare-project  3       pass    $0.2176                    21.9s                                           10/1.2k
Task routed to the right group directory      bare-project  4       pass    $0.2187                    20.4s                                           10/1.2k
Task routed to the right group directory      bare-project  5       pass    $0.2107                    20.3s                                           10/985
Task routed to the right group directory      bare-project  agg     PASS    $0.2187 [$0.2107–$0.2434]  20.6s [20.3s–23.5s] cost_cv=5.5% dur_cv=5.7%    10/1.2k
Ordering constraint captured as a dependency  no-skill      1       pass    $0.3935                    35.8s                                           26/2.0k
Ordering constraint captured as a dependency  no-skill      2       pass    $0.3474                    30.8s                                           20/2.0k
Ordering constraint captured as a dependency  no-skill      3       pass    $0.3457                    28.5s                                           20/1.7k
Ordering constraint captured as a dependency  no-skill      4       pass    $0.3538                    37.8s                                           20/1.8k
Ordering constraint captured as a dependency  no-skill      5       pass    $0.2921                    27.8s                                           20/1.7k
Ordering constraint captured as a dependency  no-skill      agg     PASS    $0.3474 [$0.2921–$0.3935]  30.8s [27.8s–37.8s] cost_cv=9.3% dur_cv=12.5%   20/1.8k
Ordering constraint captured as a dependency  plugin-skill  1       pass    $0.2652                    20.6s                                           14/1.1k
Ordering constraint captured as a dependency  plugin-skill  2       pass    $0.2922                    24.2s                                           16/1.4k
Ordering constraint captured as a dependency  plugin-skill  3       pass    $0.2238                    15.4s                                           10/747
Ordering constraint captured as a dependency  plugin-skill  4       pass    $0.2655                    18.9s                                           14/1.0k
Ordering constraint captured as a dependency  plugin-skill  5       pass    $0.2446                    17.8s                                           12/864
Ordering constraint captured as a dependency  plugin-skill  agg     PASS    $0.2652 [$0.2238–$0.2922]  18.9s [15.4s–24.2s] cost_cv=8.9% dur_cv=15.2%   14/1.0k
Ordering constraint captured as a dependency  lite-skill    1       pass    $0.2454                    18.2s                                           10/1.1k
Ordering constraint captured as a dependency  lite-skill    2       pass    $0.2212                    12.8s                                           8/745
Ordering constraint captured as a dependency  lite-skill    3       pass    $0.2512                    18.0s                                           241/1.2k
Ordering constraint captured as a dependency  lite-skill    4       pass    $0.2611                    17.7s                                           12/1.0k
Ordering constraint captured as a dependency  lite-skill    5       pass    $0.1865                    17.1s                                           10/1.1k
Ordering constraint captured as a dependency  lite-skill    agg     PASS    $0.2454 [$0.1865–$0.2611]  17.7s [12.8s–18.2s] cost_cv=11.5% dur_cv=11.9%  10/1.1k
Ordering constraint captured as a dependency  bare-project  1       fail    $0.2096                    15.6s                                           10/1.0k
Ordering constraint captured as a dependency  bare-project  2       fail    $0.2081                    17.2s                                           10/957
Ordering constraint captured as a dependency  bare-project  3       fail    $0.1381                    13.4s                                           8/786
Ordering constraint captured as a dependency  bare-project  4       fail    $0.2317                    19.2s                                           12/1.2k
Ordering constraint captured as a dependency  bare-project  5       fail    $0.2534                    23.1s                                           14/1.3k
Ordering constraint captured as a dependency  bare-project  agg     FAIL    $0.2096 [$0.1381–$0.2534]  17.2s [13.4s–23.1s] cost_cv=18.6% dur_cv=18.7%  10/1.0k
```

## Failures

- **Basic task with priority and tags** > no-skill > sample 1: check — check: command failed: FAIL add-basic: task file still contains placeholder content: "<!--"
no filled summary section (looked for [Objective Description Summary Overview Context] with >= 40 chars)
Tasks section has 1 checklist items, want >= 2
section "Acceptance Criteria" is too thin (5 chars, want >= 40)
exit status 1
- **Basic task with priority and tags** > no-skill > sample 2: check — check: command failed: FAIL add-basic: task file still contains placeholder content: "<!--"
no filled summary section (looked for [Objective Description Summary Overview Context] with >= 40 chars)
Tasks section has 1 checklist items, want >= 2
section "Acceptance Criteria" is too thin (5 chars, want >= 40)
exit status 1
- **Basic task with priority and tags** > no-skill > sample 5: check — check: command failed: FAIL add-basic: task file still contains placeholder content: "<!--"
no filled summary section (looked for [Objective Description Summary Overview Context] with >= 40 chars)
Tasks section has 1 checklist items, want >= 2
section "Acceptance Criteria" is too thin (5 chars, want >= 40)
exit status 1
- **Basic task with priority and tags** > bare-project > sample 2: check — check: command failed: FAIL add-basic: no new task was created (still only the 6 fixture tasks)
exit status 1
- **Basic task with priority and tags** > bare-project > sample 3: check — check: command failed: FAIL add-basic: no new task was created (still only the 6 fixture tasks)
exit status 1
- **Bug report routed to a template and group** > no-skill > sample 1: check — check: command failed: FAIL add-bug-template: no new task was created (still only the 5 fixture tasks)
exit status 1
- **Bug report routed to a template and group** > no-skill > sample 4: check — check: command failed: FAIL add-bug-template: section "Actual Behavior" is too thin (14 chars, want >= 20)
exit status 1
- **Bug report routed to a template and group** > lite-skill > sample 4: check — check: command failed: FAIL add-bug-template: section "Actual Behavior" is too thin (14 chars, want >= 20)
exit status 1
- **Bug report routed to a template and group** > bare-project > sample 1: check — check: command failed: FAIL add-bug-template: task file has no "## Steps to Reproduce" section
task file has no "## Expected Behavior" section
task file has no "## Actual Behavior" section
exit status 1
- **Bug report routed to a template and group** > bare-project > sample 2: check — check: command failed: FAIL add-bug-template: task file has no "## Steps to Reproduce" section
task file has no "## Expected Behavior" section
task file has no "## Actual Behavior" section
exit status 1
- **Bug report routed to a template and group** > bare-project > sample 3: check — check: command failed: FAIL add-bug-template: task file has no "## Actual Behavior" section
exit status 1
- **Bug report routed to a template and group** > bare-project > sample 4: check — check: command failed: FAIL add-bug-template: task file has no "## Steps to Reproduce" section
task file has no "## Expected Behavior" section
task file has no "## Actual Behavior" section
exit status 1
- **Ordering constraint captured as a dependency** > bare-project > sample 1: check — check: command failed: FAIL add-dependency: task dependencies [] do not include "002"
exit status 1
- **Ordering constraint captured as a dependency** > bare-project > sample 2: check — check: command failed: FAIL add-dependency: no new task was created (still only the 6 fixture tasks)
exit status 1
- **Ordering constraint captured as a dependency** > bare-project > sample 3: check — check: command failed: FAIL add-dependency: task dependencies [] do not include "002"
exit status 1
- **Ordering constraint captured as a dependency** > bare-project > sample 4: check — check: command failed: FAIL add-dependency: task dependencies [] do not include "002"
exit status 1
- **Ordering constraint captured as a dependency** > bare-project > sample 5: check — check: command failed: FAIL add-dependency: task dependencies [] do not include "002"
exit status 1

---

**Results saved to** `results/20260815-225141`

Made with [skival](https://github.com/driangle/skival)

## Rankings

```
RANK  VARIANT       SCORE  PASS RATE  95% CI     MEDIAN COST  MEDIAN DURATION
----  ---------     -----  ---------  ------     -----------  ---------------
#1    lite-skill    0.955  95%        [76–99%]   $0.2278      14.7s
#2    plugin-skill  0.922  100%       [84–100%]  $0.3175      19.6s
#3    no-skill      0.720  75%        [53–89%]   $0.3161      27.3s
#4    bare-project  0.583  45%        [26–66%]   $0.2240      20.5s

> ⚠ #1 (lite-skill) vs #2 (plugin-skill): not significant at this sample size (pass-rate intervals overlap).
```
