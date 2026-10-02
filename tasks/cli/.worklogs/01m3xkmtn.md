## 2026-10-02T09:13:19Z

Wired phase order into every next surface.

**MCP:** NewServer now takes an mcp.Config (Version, Efforts, PhaseOrder, Worktrees), since a fourth positional arg would break the object-args rule. Phase order is resolved once at startup, same as efforts, and has the same documented limitation for a per-call task_dir in another project.
**Web:** handleNext takes phases and uses effectivePhases, so project-scoped requests use the resolved project's phases. Added a phaseIDs helper and reused it in handleBoard. ExportConfig gained PhaseOrder.
**All-projects:** recommendForProject loads the project's own phases via loadProjectPhases(entry.Path).

**Finding:** passing PhaseOrder alone was not enough. runNextAllProjects re-sorted the merged list by raw score, which discarded any strict ordering within a project (both --strict-phases and --strict-priority). I replaced it with mergeProjectRecs, a k-way merge that keeps each project's order and takes the highest-scoring head. Without strict flags the result matches the old sort.

**Verified:** I temporarily removed the all-projects wiring and confirmed the new tests fail, then restored it. go test ./..., make e2e and make lint are all green. A smoke test of next --all-projects --strict-phases on real projects shows each project's own phase reasons.

**Loose end:** strict-priority tiers apply within a project but not across projects (the merge compares score, not priority tier).

## 2026-10-02T09:17:15Z

Made the reconciler's two fix-now items.
1) next --all-projects now loads each project's own effort scale (loadProjectEffortScale). Before, it used the CWD project's scale for every project, the same bug class as phases.
2) mergeProjectRecs compares heads by priority tier first when --strict-priority is set, so strict priority holds across projects. Phases stay project-local.
Tests for both fail against the old code. Full go test, e2e and lint are green. The reconciler also amended 01m3xf0rs to cover the export and all-projects.
