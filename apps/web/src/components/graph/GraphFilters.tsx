import { STATUSES, STATUS_COLORS } from "../tasks/TaskTable/constants.ts";

interface GraphFiltersProps {
  selectedStatuses: Set<string>;
  onToggleStatus: (status: string) => void;
  onClearFilters: () => void;
}

export function GraphFilters({ selectedStatuses, onToggleStatus, onClearFilters }: GraphFiltersProps) {
  const hasFilters = selectedStatuses.size > 0;

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <span className="text-xs text-gray-500 dark:text-gray-400 font-medium">Status:</span>
      {STATUSES.map((s) => {
        const active = selectedStatuses.has(s);
        return (
          <button
            key={s}
            onClick={() => onToggleStatus(s)}
            className={`px-2.5 py-1 text-xs rounded-full transition-colors duration-150 ${
              active
                ? STATUS_COLORS[s]
                : "bg-gray-50 border border-gray-200 text-gray-400 hover:bg-gray-100 hover:text-gray-500 dark:bg-gray-800/50 dark:border-gray-700 dark:text-gray-500 dark:hover:bg-gray-700 dark:hover:text-gray-400"
            }`}
          >
            {s}
          </button>
        );
      })}
      {hasFilters && (
        <button
          onClick={onClearFilters}
          className="text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 underline ml-1"
        >
          Clear
        </button>
      )}
    </div>
  );
}
