import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useReactTable,
  type SortingState,
} from "@tanstack/react-table";
import { useState, useMemo } from "react";
import { Link } from "react-router-dom";
import type { Task } from "../../api/types.ts";

const STATUSES = ["pending", "in-progress", "completed", "blocked"];
const PRIORITIES = ["critical", "high", "medium", "low"];

const STATUS_COLORS: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800",
  "in-progress": "bg-blue-100 text-blue-800",
  completed: "bg-green-100 text-green-800",
  blocked: "bg-red-100 text-red-800",
};

const PRIORITY_COLORS: Record<string, string> = {
  critical: "bg-red-100 text-red-600",
  high: "bg-orange-100 text-orange-600",
  medium: "bg-gray-100 text-gray-600",
  low: "bg-gray-50 text-gray-400",
};

function StatusBadge({ status }: { status: string }) {
  return (
    <span
      className={`px-2 py-0.5 text-xs font-medium rounded-full ${STATUS_COLORS[status] ?? "bg-gray-100 text-gray-800"}`}
    >
      {status}
    </span>
  );
}

function PriorityBadge({ priority }: { priority: string }) {
  return (
    <span
      className={`px-2 py-0.5 text-xs font-medium rounded-full ${PRIORITY_COLORS[priority] ?? "bg-gray-100 text-gray-500"}`}
    >
      {priority}
    </span>
  );
}

function toggleInSet<T>(set: Set<T>, value: T): Set<T> {
  const next = new Set(set);
  if (next.has(value)) {
    next.delete(value);
  } else {
    next.add(value);
  }
  return next;
}

interface FilterBarProps {
  globalFilter: string;
  onGlobalFilterChange: (value: string) => void;
  selectedStatuses: Set<string>;
  onToggleStatus: (status: string) => void;
  selectedPriorities: Set<string>;
  onTogglePriority: (priority: string) => void;
  selectedTags: Set<string>;
  onRemoveTag: (tag: string) => void;
  onClearFilters: () => void;
  hasActiveFilters: boolean;
}

function FilterBar({
  globalFilter,
  onGlobalFilterChange,
  selectedStatuses,
  onToggleStatus,
  selectedPriorities,
  onTogglePriority,
  selectedTags,
  onRemoveTag,
  onClearFilters,
  hasActiveFilters,
}: FilterBarProps) {
  return (
    <div className="mb-4 space-y-3">
      <div className="flex items-center gap-3 flex-wrap">
        <input
          type="text"
          value={globalFilter}
          onChange={(e) => onGlobalFilterChange(e.target.value)}
          placeholder="Filter tasks..."
          className="px-3 py-2 border border-gray-300 rounded-md text-sm w-full max-w-xs focus:outline-none focus:ring-2 focus:ring-gray-400"
        />
        {hasActiveFilters && (
          <button
            onClick={onClearFilters}
            className="text-xs text-gray-500 hover:text-gray-700 underline"
          >
            Clear filters
          </button>
        )}
      </div>

      <div className="flex items-center gap-2 flex-wrap">
        <span className="text-xs text-gray-500 font-medium">Status:</span>
        {STATUSES.map((s) => {
          const active = selectedStatuses.has(s);
          return (
            <button
              key={s}
              onClick={() => onToggleStatus(s)}
              className={`px-2.5 py-1 text-xs rounded-full transition-colors duration-150 ${
                active
                  ? STATUS_COLORS[s]
                  : "bg-white border border-gray-200 text-gray-400"
              }`}
            >
              {s}
            </button>
          );
        })}
      </div>

      <div className="flex items-center gap-2 flex-wrap">
        <span className="text-xs text-gray-500 font-medium">Priority:</span>
        {PRIORITIES.map((p) => {
          const active = selectedPriorities.has(p);
          return (
            <button
              key={p}
              onClick={() => onTogglePriority(p)}
              className={`px-2.5 py-1 text-xs rounded-full transition-colors duration-150 ${
                active
                  ? PRIORITY_COLORS[p]
                  : "bg-white border border-gray-200 text-gray-400"
              }`}
            >
              {p}
            </button>
          );
        })}
      </div>

      {selectedTags.size > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs text-gray-500 font-medium">Tags:</span>
          {[...selectedTags].map((tag) => (
            <button
              key={tag}
              onClick={() => onRemoveTag(tag)}
              className="px-2 py-0.5 text-xs bg-blue-100 text-blue-700 rounded-full ring-1 ring-blue-300 flex items-center gap-1 transition-colors duration-150 hover:bg-blue-200"
            >
              {tag}
              <span className="text-blue-400">&times;</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

interface TaskTableProps {
  tasks: Task[];
}

export function TaskTable({ tasks }: TaskTableProps) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [globalFilter, setGlobalFilter] = useState("");
  const [selectedStatuses, setSelectedStatuses] = useState<Set<string>>(
    new Set(STATUSES),
  );
  const [selectedPriorities, setSelectedPriorities] = useState<Set<string>>(
    new Set(PRIORITIES),
  );
  const [selectedTags, setSelectedTags] = useState<Set<string>>(new Set());

  const hasActiveFilters =
    selectedStatuses.size !== STATUSES.length ||
    selectedPriorities.size !== PRIORITIES.length ||
    selectedTags.size > 0 ||
    globalFilter !== "";

  function clearFilters() {
    setSelectedStatuses(new Set(STATUSES));
    setSelectedPriorities(new Set(PRIORITIES));
    setSelectedTags(new Set());
    setGlobalFilter("");
  }

  function toggleTag(tag: string) {
    setSelectedTags((prev) => toggleInSet(prev, tag));
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

  const columnHelper = createColumnHelper<Task>();

  const columns = useMemo(
    () => [
      columnHelper.accessor("id", {
        header: "ID",
        cell: (info) => (
          <Link
            to={`/tasks/${info.getValue()}`}
            className="font-mono text-xs text-blue-600 hover:underline"
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
            className="font-medium text-blue-600 hover:underline"
          >
            {info.getValue()}
          </Link>
        ),
      }),
      columnHelper.accessor("status", {
        header: "Status",
        cell: (info) => <StatusBadge status={info.getValue()} />,
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
                        ? "bg-blue-100 text-blue-700 ring-1 ring-blue-300"
                        : "bg-gray-100 text-gray-700 hover:bg-gray-200"
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
    ],
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
      <div className="overflow-x-auto rounded-lg border border-gray-200">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            {table.getHeaderGroups().map((hg) => (
              <tr key={hg.id}>
                {hg.headers.map((header) => (
                  <th
                    key={header.id}
                    onClick={header.column.getToggleSortingHandler()}
                    className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer select-none"
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
          <tbody className="bg-white divide-y divide-gray-200">
            {visibleCount === 0 ? (
              <tr>
                <td
                  colSpan={columns.length}
                  className="px-4 py-8 text-center text-sm text-gray-500"
                >
                  No tasks match your filters.{" "}
                  <button
                    onClick={clearFilters}
                    className="text-blue-600 hover:underline"
                  >
                    Clear filters
                  </button>
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="hover:bg-gray-50">
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
