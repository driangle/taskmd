import { Link } from "react-router-dom";
import type { TrackTask } from "../../api/types.ts";
import { PriorityBadge } from "../tasks/TaskTable/Badges.tsx";
import { EFFORT_COLORS } from "../tasks/TaskTable/constants.ts";

interface TrackCardProps {
  task: TrackTask;
}

export function TrackCard({ task }: TrackCardProps) {
  return (
    <div className="p-3 bg-white rounded border border-gray-100 shadow-sm dark:bg-gray-800/50 dark:border-gray-700">
      <div className="flex items-start justify-between gap-2">
        <Link
          to={`/tasks/${task.id}`}
          className="text-sm font-medium leading-snug text-blue-600 hover:underline dark:text-blue-400 flex-1"
        >
          {task.title}
        </Link>
        <span className="text-xs text-gray-400 dark:text-gray-500 font-mono shrink-0">
          {task.id}
        </span>
      </div>

      <div className="flex items-center gap-1.5 mt-1.5 flex-wrap">
        {task.priority && <PriorityBadge priority={task.priority} />}
        {task.effort && (
          <span
            className={`px-2 py-0.5 text-xs font-medium rounded-full ${EFFORT_COLORS[task.effort] ?? "bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400"}`}
          >
            {task.effort}
          </span>
        )}
        <span className="text-xs text-gray-400 dark:text-gray-500 ml-auto">
          {task.score} pts
        </span>
      </div>

      {task.touches && task.touches.length > 0 && (
        <div className="flex items-center gap-1 mt-2 flex-wrap">
          {task.touches.map((scope) => (
            <span
              key={scope}
              className="px-1.5 py-0.5 text-xs rounded bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300"
            >
              {scope}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
