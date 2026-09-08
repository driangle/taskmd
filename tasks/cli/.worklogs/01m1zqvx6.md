## 2026-09-08T07:12:54Z

Investigated a user report that `validate` returned zero tasks inside a git worktree. Reproduced against gg/pinspot, where worktrees are nested under the primary checkout at .claude/worktrees/. Bisected with --worktree-scope isolated (128 tasks) vs unified (0), pinning it to the overlay.

## 2026-09-08T07:13:01Z

Root cause: AttributeNestedSiblingCopies dropped local tasks whose path fell under any sibling root, without checking nesting direction. The primary checkout is an ancestor of a nested worktree, so it swallowed the whole local list. Fixed by filtering to siblings whose root is strictly inside scanDir; threaded scanDir through Builder.Overlay. Added attribution_test.go covering ancestor, nested (spec §8), and beside-the-repo layouts. Verified against pinspot: 128 tasks from both the worktree and the primary. go test ./..., make lint (0 issues), and make e2e all pass. Marking completed retroactively.
