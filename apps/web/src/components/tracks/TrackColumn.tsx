import type { Track } from "../../api/types.ts";
import { TrackCard } from "./TrackCard.tsx";

const trackColors = [
  "border-blue-300 dark:border-blue-700",
  "border-emerald-300 dark:border-emerald-700",
  "border-purple-300 dark:border-purple-700",
  "border-amber-300 dark:border-amber-700",
  "border-rose-300 dark:border-rose-700",
  "border-cyan-300 dark:border-cyan-700",
];

interface TrackColumnProps {
  track: Track;
}

export function TrackColumn({ track }: TrackColumnProps) {
  const colorClass = trackColors[(track.id - 1) % trackColors.length];

  return (
    <div
      className={`flex-shrink-0 w-full md:w-72 rounded-lg border-t-4 bg-white shadow-sm dark:bg-gray-800 ${colorClass}`}
    >
      <div className="px-4 py-3 border-b border-gray-100 dark:border-gray-700">
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200">
          Track {track.id}{" "}
          <span className="text-gray-400 dark:text-gray-500 font-normal">
            ({track.tasks.length})
          </span>
        </h3>
        {track.scopes.length > 0 && (
          <div className="flex items-center gap-1 mt-1.5 flex-wrap">
            {track.scopes.map((scope) => (
              <span
                key={scope}
                className="px-1.5 py-0.5 text-xs rounded bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400"
              >
                {scope}
              </span>
            ))}
          </div>
        )}
      </div>
      <div className="p-2 space-y-2">
        {track.tasks.map((task) => (
          <TrackCard key={task.id} task={task} />
        ))}
      </div>
    </div>
  );
}
