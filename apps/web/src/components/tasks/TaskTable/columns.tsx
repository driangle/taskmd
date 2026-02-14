import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";
import type { Task } from "../../../api/types.ts";
import { StatusBadge, PriorityBadge, BlockedStatusBadge } from "./Badges.tsx";

export function createTaskColumns(
  selectedTags: Set<string>,
  toggleTag: (tag: string) => void,
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
      cell: (info) => <StatusBadge status={info.getValue()} />,
    }),
    columnHelper.accessor("dependencies", {
      id: "blocked",
      header: "Blocked",
      cell: (info) => <BlockedStatusBadge dependencies={info.getValue()} />,
      sortingFn: (rowA, rowB) => {
        const aCount = rowA.original.dependencies?.length ?? 0;
        const bCount = rowB.original.dependencies?.length ?? 0;
        return aCount - bCount;
      },
    }),
    columnHelper.accessor("priority", {
      header: "Priority",
      cell: (info) => {
        const v = info.getValue();
        return v ? <PriorityBadge priority={v} /> : null;
      },
    }),
    columnHelper.accessor("effort", {
      header: "Effort",
      cell: (info) => info.getValue() || "-",
    }),
    columnHelper.accessor("tags", {
      header: "Tags",
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
