import { Link } from "react-router-dom";
import type { Recommendation } from "../../api/types.ts";
import { PriorityBadge } from "../tasks/TaskTable/Badges.tsx";

interface RecommendationCardProps {
  rec: Recommendation;
}

export function RecommendationCard({ rec }: RecommendationCardProps) {
  return (
    <div className="flex items-start gap-4 p-4 bg-white border border-gray-200 rounded-lg dark:bg-gray-800 dark:border-gray-700">
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-900 text-white text-sm font-bold shrink-0 dark:bg-white dark:text-gray-900">
        {rec.rank}
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <Link
            to={`/tasks/${rec.id}`}
            className="text-sm font-medium hover:underline truncate"
          >
            {rec.title}
          </Link>
          <span className="text-xs text-gray-400 dark:text-gray-500 shrink-0">
            {rec.id}
          </span>
        </div>

        <div className="flex items-center gap-2 mt-1.5 flex-wrap">
          <PriorityBadge priority={rec.priority} />
          {rec.on_critical_path && (
            <span className="px-2 py-0.5 text-xs font-medium rounded-full bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300">
              critical path
            </span>
          )}
          {rec.downstream_count > 0 && (
            <span className="text-xs text-gray-500 dark:text-gray-400">
              unblocks {rec.downstream_count}{" "}
              {rec.downstream_count === 1 ? "task" : "tasks"}
            </span>
          )}
        </div>

        <div className="flex items-center gap-1.5 mt-2 flex-wrap">
          {rec.reasons.map((reason) => (
            <span
              key={reason}
              className="px-2 py-0.5 text-xs rounded-full bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300"
            >
              {reason}
            </span>
          ))}
        </div>
      </div>

      <div className="text-right shrink-0">
        <span className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {rec.score}
        </span>
        <div className="text-xs text-gray-400 dark:text-gray-500">score</div>
      </div>
    </div>
  );
}
