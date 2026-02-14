import { Link } from "react-router-dom";
import type { BoardGroup } from "../../api/types.ts";

const statusColors: Record<string, string> = {
  pending: "border-yellow-300 bg-yellow-50 dark:bg-yellow-900/20",
  "in-progress": "border-blue-300 bg-blue-50 dark:bg-blue-900/20",
  completed: "border-green-300 bg-green-50 dark:bg-green-900/20",
  blocked: "border-red-300 bg-red-50 dark:bg-red-900/20",
};

interface BoardViewProps {
  groups: BoardGroup[];
}

export function BoardView({ groups }: BoardViewProps) {
  return (
    <div className="flex gap-4 overflow-x-auto pb-4">
      {groups.map((g) => (
        <div
          key={g.group}
          className={`flex-shrink-0 w-72 rounded-lg border-t-4 bg-white shadow-sm dark:bg-gray-800 ${
            statusColors[g.group] ?? "border-gray-300 bg-gray-50 dark:bg-gray-800"
          }`}
        >
          <div className="px-4 py-3 border-b border-gray-100 dark:border-gray-700">
            <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200">
              {g.group}{" "}
              <span className="text-gray-400 dark:text-gray-500 font-normal">({g.count})</span>
            </h3>
          </div>
          <div className="p-2 space-y-2">
            {g.tasks.map((task) => (
              <div
                key={task.id}
                className="p-3 bg-white rounded border border-gray-100 shadow-sm dark:bg-gray-800/50 dark:border-gray-700"
              >
                <div className="flex items-start justify-between gap-2">
                  <Link
                    to={`/tasks/${task.id}`}
                    className="text-sm font-medium leading-snug text-blue-600 hover:underline dark:text-blue-400"
                  >
                    {task.title}
                  </Link>
                  <span className="text-xs text-gray-400 dark:text-gray-500 font-mono shrink-0">
                    {task.id}
                  </span>
                </div>
                {task.priority && (
                  <span className="mt-1.5 inline-block text-xs text-gray-500 dark:text-gray-400">
                    {task.priority}
                  </span>
                )}
              </div>
            ))}
            {g.tasks.length === 0 && (
              <p className="text-xs text-gray-400 dark:text-gray-500 p-2">No tasks</p>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
