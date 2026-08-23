import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import type { Task } from "../../../api/types.ts";
import { StatusBadge, PriorityBadge, TypeBadge, PhaseBadge, BlockedStatusBadge, WorktreeBadge } from "./Badges.tsx";
import { compareBlocked, comparePriority } from "./sorting.ts";
import { effectiveStatus } from "./utils.ts";

export function createTaskColumns(
  selectedTags: Set<string>,
  toggleTag: (tag: string) => void,
  taskStatusMap?: Map<string, string>,
  showPhase = true,
) {
  const columnHelper = createColumnHelper<Task>();

  return [
    columnHelper.accessor("id", {
      header: "ID",
      cell: (info) => (
        <Link
          to={`/tasks/${info.getValue()}`}
          className="font-mono text-xs text-blue-600 hover:underline dark:text-blue-400"
        >
          {info.getValue()}
        </Link>
      ),
    }),
    columnHelper.accessor("title", {
      header: "Title",
      cell: (info) => (
        <Link
          to={`/tasks/${info.row.original.id}`}
          className="font-medium text-blue-600 hover:underline dark:text-blue-400"
        >
          {info.getValue()}
        </Link>
      ),
    }),
    columnHelper.accessor("status", {
      header: "Status",
      cell: (info) => {
        const task = info.row.original;
        const remote = Boolean(task.worktree) || Boolean(task.remote_only);
        return (
          <div className="flex items-center gap-1.5">
            <StatusBadge status={effectiveStatus(task)} />
            {remote && (
              <WorktreeBadge
                worktree={task.worktree}
                branch={task.branch}
                remoteOnly={task.remote_only}
              />
            )}
          </div>
        );
      },
    }),
    columnHelper.accessor("dependencies", {
      id: "blocked",
      header: "Blocked",
      meta: { className: "hidden md:table-cell" },
      cell: (info) => <BlockedStatusBadge dependencies={info.getValue()} taskStatusMap={taskStatusMap} />,
      sortingFn: (rowA, rowB) =>
        compareBlocked(rowA.original.dependencies, rowB.original.dependencies, taskStatusMap),
    }),
    columnHelper.accessor("priority", {
      header: "Priority",
      meta: { className: "hidden sm:table-cell" },
      cell: (info) => {
        const v = info.getValue();
        return v ? <PriorityBadge priority={v} /> : null;
      },
      sortingFn: (rowA, rowB) =>
        comparePriority(rowA.original.priority, rowB.original.priority),
    }),
    columnHelper.accessor("effort", {
      header: "Effort",
      meta: { className: "hidden md:table-cell" },
      cell: (info) => info.getValue() || "-",
    }),
    columnHelper.accessor("type", {
      header: "Type",
      meta: { className: "hidden md:table-cell" },
      cell: (info) => {
        const v = info.getValue();
        return v ? <TypeBadge type={v} /> : "-";
      },
    }),
    columnHelper.accessor("phase", {
      header: "Phase",
      meta: { className: "hidden md:table-cell" },
      cell: (info) => {
        const v = info.getValue();
        if (!v) return "-";
        return showPhase ? <PhaseBadge phase={v} /> : v;
      },
    }),
    columnHelper.accessor("owner", {
      header: "Owner",
      meta: { className: "hidden md:table-cell" },
      cell: (info) => info.getValue() || "-",
    }),
    columnHelper.accessor("tags", {
      header: "Tags",
      meta: { className: "hidden md:table-cell" },
      cell: (info) => {
        const tags = info.getValue();
        if (!tags || tags.length === 0) return "-";
        return (
          <div className="flex gap-1 flex-wrap">
            {tags.map((t) => {
              const isActive = selectedTags.has(t);
              return (
                <button
                  key={t}
                  onClick={() => toggleTag(t)}
                  aria-pressed={isActive}
                  className={`px-1.5 py-0.5 text-xs rounded cursor-pointer transition-colors duration-150 ${
                    isActive
                      ? "bg-blue-100 text-blue-700 ring-1 ring-blue-300 dark:bg-blue-900/30 dark:text-blue-300 dark:ring-blue-700"
                      : "bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
                  }`}
                >
                  {t}
                </button>
              );
            })}
          </div>
        );
      },
      enableSorting: false,
    }),
  ];
}
