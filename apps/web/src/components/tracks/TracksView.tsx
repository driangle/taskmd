import type { TracksResult } from "../../api/types.ts";
import { TrackColumn } from "./TrackColumn.tsx";
import { FlexibleSection } from "./FlexibleSection.tsx";

interface TracksViewProps {
  data: TracksResult;
  limit: number;
  onLimitChange: (limit: number) => void;
}

const LIMIT_OPTIONS = [0, 2, 3, 5];

export function TracksView({ data, limit, onLimitChange }: TracksViewProps) {
  const isEmpty = data.tracks.length === 0 && data.flexible.length === 0;

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold">Parallel Tracks</h2>
        <div className="flex items-center gap-1">
          <span className="text-sm text-gray-500 dark:text-gray-400 mr-2">
            Max tracks:
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
              {n === 0 ? "All" : n}
            </button>
          ))}
        </div>
      </div>

      {data.warnings && data.warnings.length > 0 && (
        <div className="mb-4 rounded-lg border border-amber-200 bg-amber-50 p-3 dark:border-amber-800 dark:bg-amber-900/20">
          <p className="text-sm font-medium text-amber-800 dark:text-amber-300 mb-1">
            Warnings
          </p>
          <ul className="text-xs text-amber-700 dark:text-amber-400 space-y-0.5">
            {data.warnings.map((w, i) => (
              <li key={i}>{w}</li>
            ))}
          </ul>
        </div>
      )}

      {isEmpty && (
        <p className="text-sm text-gray-500 py-8 text-center">
          No actionable tasks found. All tasks are either completed or blocked.
        </p>
      )}

      {data.tracks.length > 0 && (
        <div className="flex flex-col md:flex-row gap-4 md:overflow-x-auto pb-4">
          {data.tracks.map((track) => (
            <TrackColumn key={track.id} track={track} />
          ))}
        </div>
      )}

      <FlexibleSection tasks={data.flexible} />
    </div>
  );
}
