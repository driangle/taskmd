import type { Recommendation } from "../../api/types.ts";
import { RecommendationCard } from "./RecommendationCard.tsx";

interface NextViewProps {
  recommendations: Recommendation[];
  limit: number;
  onLimitChange: (limit: number) => void;
}

const LIMIT_OPTIONS = [3, 5, 10];

export function NextView({
  recommendations,
  limit,
  onLimitChange,
}: NextViewProps) {
  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold">Recommended Tasks</h2>
        <div className="flex items-center gap-1">
          <span className="text-sm text-gray-500 dark:text-gray-400 mr-2">
            Show:
          </span>
          {LIMIT_OPTIONS.map((n) => (
            <button
              key={n}
              onClick={() => onLimitChange(n)}
              className={`px-3 py-1 text-sm rounded-md transition-colors ${
                limit === n
                  ? "bg-gray-900 text-white dark:bg-white dark:text-gray-900"
                  : "text-gray-600 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-100 dark:hover:bg-gray-700"
              }`}
            >
              {n}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-3">
        {recommendations.map((rec) => (
          <RecommendationCard key={rec.id} rec={rec} />
        ))}
      </div>
    </div>
  );
}
