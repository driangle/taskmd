import { Link } from "react-router-dom";
import type { BoardGroup } from "../../api/types.ts";

const statusColors: Record<string, string> = {
  pending: "border-yellow-300 bg-yellow-50",
  "in-progress": "border-blue-300 bg-blue-50",
  completed: "border-green-300 bg-green-50",
  blocked: "border-red-300 bg-red-50",
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
          className={`flex-shrink-0 w-72 rounded-lg border-t-4 bg-white shadow-sm ${
            statusColors[g.group] ?? "border-gray-300 bg-gray-50"
          }`}
        >
          <div className="px-4 py-3 border-b border-gray-100">
            <h3 className="text-sm font-semibold text-gray-700">
              {g.group}{" "}
              <span className="text-gray-400 font-normal">({g.count})</span>
            </h3>
          </div>
          <div className="p-2 space-y-2">
            {g.tasks.map((task) => (
              <div
                key={task.id}
                className="p-3 bg-white rounded border border-gray-100 shadow-sm"
              >
                <div className="flex items-start justify-between gap-2">
                  <Link
                    to={`/tasks/${task.id}`}
                    className="text-sm font-medium leading-snug text-blue-600 hover:underline"
                  >
                    {task.title}
                  </Link>
                  <span className="text-xs text-gray-400 font-mono shrink-0">
                    {task.id}
                  </span>
                </div>
                {task.priority && (
                  <span className="mt-1.5 inline-block text-xs text-gray-500">
                    {task.priority}
                  </span>
                )}
              </div>
            ))}
            {g.tasks.length === 0 && (
              <p className="text-xs text-gray-400 p-2">No tasks</p>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
