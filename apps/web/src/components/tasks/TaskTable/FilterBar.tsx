import { STATUSES, PRIORITIES, STATUS_COLORS, PRIORITY_COLORS } from "./constants.ts";

export interface FilterBarProps {
  globalFilter: string;
  onGlobalFilterChange: (value: string) => void;
  selectedStatuses: Set<string>;
  onToggleStatus: (status: string) => void;
  selectedPriorities: Set<string>;
  onTogglePriority: (priority: string) => void;
  selectedTags: Set<string>;
  onRemoveTag: (tag: string) => void;
  onClearFilters: () => void;
  hasActiveFilters: boolean;
}

export function FilterBar({
  globalFilter,
  onGlobalFilterChange,
  selectedStatuses,
  onToggleStatus,
  selectedPriorities,
  onTogglePriority,
  selectedTags,
  onRemoveTag,
  onClearFilters,
  hasActiveFilters,
}: FilterBarProps) {
  return (
    <div className="mb-4 space-y-3">
      <div className="flex items-center gap-3 flex-wrap">
        <input
          type="text"
          value={globalFilter}
          onChange={(e) => onGlobalFilterChange(e.target.value)}
          placeholder="Filter tasks..."
          className="px-3 py-2 border border-gray-300 rounded-md text-sm w-full max-w-xs focus:outline-none focus:ring-2 focus:ring-gray-400"
        />
        {hasActiveFilters && (
          <button
            onClick={onClearFilters}
            className="text-xs text-gray-500 hover:text-gray-700 underline"
          >
            Clear filters
          </button>
        )}
      </div>

      <div className="flex items-center gap-2 flex-wrap">
        <span className="text-xs text-gray-500 font-medium">Status:</span>
        {STATUSES.map((s) => {
          const active = selectedStatuses.has(s);
          return (
            <button
              key={s}
              onClick={() => onToggleStatus(s)}
              className={`px-2.5 py-1 text-xs rounded-full transition-colors duration-150 ${
                active
                  ? STATUS_COLORS[s]
                  : "bg-white border border-gray-200 text-gray-400"
              }`}
            >
              {s}
            </button>
          );
        })}
      </div>

      <div className="flex items-center gap-2 flex-wrap">
        <span className="text-xs text-gray-500 font-medium">Priority:</span>
        {PRIORITIES.map((p) => {
          const active = selectedPriorities.has(p);
          return (
            <button
              key={p}
              onClick={() => onTogglePriority(p)}
              className={`px-2.5 py-1 text-xs rounded-full transition-colors duration-150 ${
                active
                  ? PRIORITY_COLORS[p]
                  : "bg-white border border-gray-200 text-gray-400"
              }`}
            >
              {p}
            </button>
          );
        })}
      </div>

      {selectedTags.size > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs text-gray-500 font-medium">Tags:</span>
          {[...selectedTags].map((tag) => (
            <button
              key={tag}
              onClick={() => onRemoveTag(tag)}
              className="px-2 py-0.5 text-xs bg-blue-100 text-blue-700 rounded-full ring-1 ring-blue-300 flex items-center gap-1 transition-colors duration-150 hover:bg-blue-200"
            >
              {tag}
              <span className="text-blue-400">&times;</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
