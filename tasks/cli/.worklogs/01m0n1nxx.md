## 2026-08-23T06:19:05Z

Starting task. Plan: read spec §7 + ADR 0005 §1, then wire gitmeta repo identity into the projects registry (register dedupe, cwd matching, --project worktree scoping, --all-projects dedupe).

## 2026-08-23T06:22:52Z

Context gathered (spec §7, ADR 0005 §1, gitmeta, registry code). Design: new internal/cli/repo_identity.go with a resolveRepoIdentity seam over gitmeta.Resolve plus primaryWorktreeRoot/sameRepo/rel-path mapping helpers. register canonicalizes target into the primary worktree via rel mapping (preserves monorepo subproject paths; non-git and bare repos pass through) and no-ops with 'already registered as <id>' when the repo identity matches an existing entry. resolveProjectDir remaps the registered path into the cwd's worktree when identities match, so --project scopes to the local checkout (overlay applies via scanDir). scanAllProjects dedupes entries by CommonDir and retargets to the current worktree when inside that repo. Also fixing the stale ~/.taskmd/config.yaml hint in projects.go.

## 2026-08-23T06:29:09Z

Implementation complete. New internal/cli/repo_identity.go holds the identity helpers behind a resolveRepoIdentity seam (stubbed to nil in resetCLIState so command tests stay hermetic). register canonicalizes the target through canonicalRegisterPath — a rel-path remap from the current worktree root onto the primary root derived from the common dir, which preserves monorepo subproject paths and passes through non-git/bare targets — and registerProject no-ops with 'Already registered as <id>' when repo identity matches an existing entry (non-git keeps the old ID-only duplicate handling). resolveProjectDir remaps the registered path into the cwd's worktree via localWorktreePathFor; scanAllProjects dedupes entries by CommonDir and retargets to the current worktree. Fixed the stale ~/.taskmd/config.yaml hint. Tests: 8 unit tests in repo_identity_test.go plus 3 e2e tests in projects_worktree_test.go using real git worktree add. Full unit + e2e suites and lint green; smoke-tested against this repo's own 3 worktrees (list --project taskmd from taskmd-2 scans taskmd-2's files). Registry schema untouched ({id,name,path}).
