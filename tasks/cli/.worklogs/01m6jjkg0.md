## 2026-11-04T08:37:41Z

Started. Plan: add --group shortcut on list that folds into filters as group=<value>, same path as --status/--priority/--phase. Task file was untracked in main checkout; copied into worktree to commit with the change.

## 2026-11-04T08:40:43Z

Done. --group added to FilterShortcuts and folded in as group=<value> alongside status/priority/phase, so wildcards and none/any come from sdk/go/filter. Help text says 'filter by group'. No explicit flag-var resets exist in list tests: resetCLIState/resetFlagTree restores every registered flag to its default, so listGroup is covered automatically. Tests: TestListCommand_GroupFlag (exact, frontmatter override, wildcard, none, any, +status, +filter, no matches, unchanged without flag), parity with --filter group=, table no-match, applyShortcutFilters Group; e2e TestListGroup. Docs: apps/docs/guide/cli.md list flags table + examples. make test, make e2e, make lint all pass; manual check on repo tasks/: --group cli and --filter group=cli both return 147.
