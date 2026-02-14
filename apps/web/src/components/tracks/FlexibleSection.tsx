import type { TrackTask } from "../../api/types.ts";
import { TrackCard } from "./TrackCard.tsx";

interface FlexibleSectionProps {
  tasks: TrackTask[];
}

export function FlexibleSection({ tasks }: FlexibleSectionProps) {
  if (tasks.length === 0) return null;

  return (
    <div className="mt-6">
      <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">
        Flexible{" "}
        <span className="text-gray-400 dark:text-gray-500 font-normal">
          ({tasks.length})
        </span>
      </h3>
      <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
        These tasks have no <code className="px-1 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-xs">touches</code> scope defined and can be worked on alongside any track.
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
        {tasks.map((task) => (
          <TrackCard key={task.id} task={task} />
        ))}
      </div>
    </div>
  );
}
