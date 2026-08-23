import type { WorktreeCopy } from "../../api/types.ts";
import { StatusBadge } from "./TaskTable/Badges.tsx";

/**
 * Per-worktree copies of a task, shown on the detail page when the copies
 * diverge (the API sends `worktrees` only in that case), mirroring `taskmd get`.
 */
export function WorktreeCopiesSection({ copies }: { copies: WorktreeCopy[] }) {
  return (
    <div className="mb-6">
      <h3 className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-2">
        Worktree Copies
      </h3>
      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-left text-xs text-gray-500 dark:text-gray-400">
              <th className="pr-4 pb-1 font-medium">Worktree</th>
              <th className="pr-4 pb-1 font-medium">Branch</th>
              <th className="pr-4 pb-1 font-medium">Status</th>
              <th className="pb-1 font-medium">Owner</th>
            </tr>
          </thead>
          <tbody>
            {copies.map((copy, i) => (
              <tr key={`${copy.worktree ?? "local"}-${i}`} className="border-t border-gray-100 dark:border-gray-700">
                <td className="pr-4 py-1.5 font-mono text-xs">
                  {copy.local ? (
                    <span>
                      {copy.worktree || "this worktree"}{" "}
                      <span className="text-gray-400 dark:text-gray-500">(local)</span>
                    </span>
                  ) : (
                    copy.worktree || "—"
                  )}
                </td>
                <td className="pr-4 py-1.5 font-mono text-xs">{copy.branch || "—"}</td>
                <td className="pr-4 py-1.5">
                  <StatusBadge status={copy.status} />
                </td>
                <td className="py-1.5">{copy.owner || "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
