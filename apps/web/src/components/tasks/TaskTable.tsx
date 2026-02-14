import {
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useReactTable,
  type SortingState,
} from "@tanstack/react-table";
import { useState, useMemo, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import type { Task } from "../../api/types.ts";
import { STATUSES, PRIORITIES } from "./TaskTable/constants.ts";
import { FilterBar } from "./TaskTable/FilterBar.tsx";
import { createTaskColumns } from "./TaskTable/columns.tsx";
import { toggleInSet } from "./TaskTable/utils.ts";

interface TaskTableProps {
  tasks: Task[];
  initialTags?: string[];
}

export function TaskTable({ tasks, initialTags }: TaskTableProps) {
  const [, setSearchParams] = useSearchParams();
  const [sorting, setSorting] = useState<SortingState>([]);
  const [globalFilter, setGlobalFilter] = useState("");
  const [selectedStatuses, setSelectedStatuses] = useState<Set<string>>(
    new Set(STATUSES),
  );
  const [selectedPriorities, setSelectedPriorities] = useState<Set<string>>(
    new Set(PRIORITIES),
  );
  const [selectedTags, setSelectedTags] = useState<Set<string>>(
    () => new Set(initialTags),
  );

  const hasActiveFilters =
    selectedStatuses.size !== STATUSES.length ||
    selectedPriorities.size !== PRIORITIES.length ||
    selectedTags.size > 0 ||
    globalFilter !== "";

  const syncTagsToUrl = useCallback(
    (tags: Set<string>) => {
      setSearchParams(
        (prev) => {
          prev.delete("tag");
          tags.forEach((t) => prev.append("tag", t));
          return prev;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  function clearFilters() {
    setSelectedStatuses(new Set(STATUSES));
    setSelectedPriorities(new Set(PRIORITIES));
    setSelectedTags(new Set());
    syncTagsToUrl(new Set());
    setGlobalFilter("");
  }

  function toggleTag(tag: string) {
    setSelectedTags((prev) => {
      const next = toggleInSet(prev, tag);
      syncTagsToUrl(next);
      return next;
    });
  }

  const filteredTasks = useMemo(() => {
    return tasks.filter((task) => {
      if (!selectedStatuses.has(task.status)) return false;
      if (task.priority && !selectedPriorities.has(task.priority)) return false;
      if (selectedTags.size > 0) {
        if (!task.tags || !task.tags.some((t) => selectedTags.has(t)))
          return false;
      }
      return true;
    });
  }, [tasks, selectedStatuses, selectedPriorities, selectedTags]);

  const columns = useMemo(
    () => createTaskColumns(selectedTags, toggleTag),
    [selectedTags],
  );

  const table = useReactTable({
    data: filteredTasks,
    columns,
    state: { sorting, globalFilter },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  const visibleCount = table.getRowModel().rows.length;

  return (
    <div>
      <FilterBar
        globalFilter={globalFilter}
        onGlobalFilterChange={setGlobalFilter}
        selectedStatuses={selectedStatuses}
        onToggleStatus={(s) =>
          setSelectedStatuses((prev) => toggleInSet(prev, s))
        }
        selectedPriorities={selectedPriorities}
        onTogglePriority={(p) =>
          setSelectedPriorities((prev) => toggleInSet(prev, p))
        }
        selectedTags={selectedTags}
        onRemoveTag={toggleTag}
        onClearFilters={clearFilters}
        hasActiveFilters={hasActiveFilters}
      />
      <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-800">
            {table.getHeaderGroups().map((hg) => (
              <tr key={hg.id}>
                {hg.headers.map((header) => (
                  <th
                    key={header.id}
                    onClick={header.column.getToggleSortingHandler()}
                    className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer select-none"
                  >
                    <div className="flex items-center gap-1">
                      {flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                      {{ asc: " ^", desc: " v" }[
                        header.column.getIsSorted() as string
                      ] ?? ""}
                    </div>
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody className="bg-white divide-y divide-gray-200 dark:bg-gray-800 dark:divide-gray-700">
            {visibleCount === 0 ? (
              <tr>
                <td
                  colSpan={columns.length}
                  className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400"
                >
                  No tasks match your filters.{" "}
                  <button
                    onClick={clearFilters}
                    className="text-blue-600 hover:underline dark:text-blue-400"
                  >
                    Clear filters
                  </button>
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="px-4 py-3 text-sm">
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      <p className="mt-2 text-xs text-gray-400">
        Showing {visibleCount} of {tasks.length} tasks
      </p>
    </div>
  );
}
