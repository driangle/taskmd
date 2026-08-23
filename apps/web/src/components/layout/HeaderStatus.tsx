import type { WorktreeOverlayInfo } from "../../hooks/use-config.ts";

/**
 * Status pills next to the header branding: version, read-only mode, and the
 * worktree overlay indicator ("worktree agent-b — 3 siblings").
 */
export function HeaderStatus({
  version,
  readonly,
  worktree,
}: {
  version: string;
  readonly: boolean;
  worktree?: WorktreeOverlayInfo;
}) {
  return (
    <>
      {version && (
        <span className="text-xs text-gray-400 dark:text-gray-500">
          {version}
        </span>
      )}
      {readonly && (
        <span className="px-2 py-0.5 text-xs font-medium rounded-full bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300">
          Read Only
        </span>
      )}
      {worktree && (
        <span
          className="px-2 py-0.5 text-xs font-medium rounded-full bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300 whitespace-nowrap"
          title="Tasks are merged across this repository's git worktrees"
        >
          <span aria-hidden="true">⎇ </span>
          {worktree.name ? `worktree ${worktree.name}` : "worktrees"} —{" "}
          {worktree.siblings} sibling{worktree.siblings === 1 ? "" : "s"}
        </span>
      )}
    </>
  );
}
