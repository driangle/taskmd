import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useReactTable,
  type SortingState,
} from "@tanstack/react-table";
import { useState } from "react";
import { Link } from "react-router-dom";
import type { Task } from "../../api/types.ts";

const columnHelper = createColumnHelper<Task>();

const columns = [
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
          {tags.map((t) => (
            <span
              key={t}
              className="px-1.5 py-0.5 text-xs bg-gray-100 rounded"
            >
              {t}
            </span>
          ))}
        </div>
      );
    },
    enableSorting: false,
  }),
];

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    pending: "bg-yellow-100 text-yellow-800",
    "in-progress": "bg-blue-100 text-blue-800",
    completed: "bg-green-100 text-green-800",
    blocked: "bg-red-100 text-red-800",
  };
  return (
    <span
      className={`px-2 py-0.5 text-xs font-medium rounded-full ${colors[status] ?? "bg-gray-100 text-gray-800"}`}
    >
      {status}
    </span>
  );
}

function PriorityBadge({ priority }: { priority: string }) {
  const colors: Record<string, string> = {
    critical: "text-red-600 font-semibold",
    high: "text-orange-600",
    medium: "text-gray-600",
    low: "text-gray-400",
  };
  return (
    <span className={`text-xs ${colors[priority] ?? "text-gray-500"}`}>
      {priority}
    </span>
  );
}

interface TaskTableProps {
  tasks: Task[];
}

export function TaskTable({ tasks }: TaskTableProps) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [globalFilter, setGlobalFilter] = useState("");

  const table = useReactTable({
    data: tasks,
    columns,
    state: { sorting, globalFilter },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  return (
    <div>
      <input
        type="text"
        value={globalFilter}
        onChange={(e) => setGlobalFilter(e.target.value)}
        placeholder="Filter tasks..."
        className="mb-4 px-3 py-2 border border-gray-300 rounded-md text-sm w-full max-w-xs focus:outline-none focus:ring-2 focus:ring-gray-400"
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
            {table.getRowModel().rows.map((row) => (
              <tr key={row.id} className="hover:bg-gray-50">
                {row.getVisibleCells().map((cell) => (
                  <td key={cell.id} className="px-4 py-3 text-sm">
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <p className="mt-2 text-xs text-gray-400">
        {table.getRowModel().rows.length} tasks
      </p>
    </div>
  );
}
