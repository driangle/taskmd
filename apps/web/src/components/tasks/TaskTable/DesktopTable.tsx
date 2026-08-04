import { flexRender, type Table, type Row } from "@tanstack/react-table";
import { useNavigate } from "react-router-dom";
import { KeyboardList } from "../../shared/KeyboardList.tsx";
import type { Task } from "../../../api/types.ts";

interface DesktopTableProps {
  table: Table<Task>;
  rows: Row<Task>[];
  columns: { length: number };
  clearFilters: () => void;
}

export function DesktopTable({ table, rows, columns, clearFilters }: DesktopTableProps) {
  const navigate = useNavigate();
  const visibleCount = rows.length;

  return (
    <KeyboardList
      className="hidden sm:block overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700"
      role="grid"
      aria-label="Task list"
      itemCount={visibleCount}
      onActivate={(index) => {
        const task = rows[index]?.original;
        if (task) navigate(`/tasks/${task.id}`);
      }}
    >
      {(focusedRowIndex) => (
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-800">
            {table.getHeaderGroups().map((hg) => (
              <tr key={hg.id}>
                {hg.headers.map((header) => {
                  const sorted = header.column.getIsSorted();
                  return (
                    <th
                      key={header.id}
                      onClick={header.column.getToggleSortingHandler()}
                      aria-sort={
                        sorted === "asc"
                          ? "ascending"
                          : sorted === "desc"
                            ? "descending"
                            : "none"
                      }
                      className={`px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer select-none ${(header.column.columnDef.meta as Record<string, string>)?.className ?? ""}`}
                    >
                      <div className="flex items-center gap-1">
                        {flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
                        {{ asc: " ^", desc: " v" }[sorted as string] ?? ""}
                      </div>
                    </th>
                  );
                })}
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
              rows.map((row, idx) => (
                <tr
                  key={row.id}
                  id={`task-row-${idx}`}
                  aria-selected={idx === focusedRowIndex}
                  className={`hover:bg-gray-50 dark:hover:bg-gray-700 ${idx === focusedRowIndex ? "ring-2 ring-inset ring-blue-500" : ""}`}
                >
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className={`px-4 py-3 text-sm ${(cell.column.columnDef.meta as Record<string, string>)?.className ?? ""}`}>
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
      )}
    </KeyboardList>
  );
}
