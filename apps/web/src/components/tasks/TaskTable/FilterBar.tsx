import { useState } from "react";
import { STATUSES, PRIORITIES, TYPES, STATUS_COLORS, PRIORITY_COLORS, TYPE_COLORS, getPhaseColor } from "./constants.ts";
import { useEffortBadges } from "./effort-colors.ts";
import { FilterRow, INACTIVE_STYLE } from "./FilterRow.tsx";

export interface FilterBarProps {
  globalFilter: string;
  onGlobalFilterChange: (value: string) => void;
  selectedStatuses: Set<string>;
  onToggleStatus: (status: string) => void;
  onSelectAllStatuses: () => void;
  selectedPriorities: Set<string>;
  onTogglePriority: (priority: string) => void;
  onSelectAllPriorities: () => void;
  /** The project's configured effort vocabulary. */
  efforts: string[];
  selectedEffort: Set<string>;
  onToggleEffort: (effort: string) => void;
  onSelectAllEffort: () => void;
  selectedTypes: Set<string>;
  onToggleType: (type: string) => void;
  onSelectAllTypes: () => void;
  selectedTags: Set<string>;
  onRemoveTag: (tag: string) => void;
  selectedPhases: Set<string>;
  availablePhases: string[];
  onTogglePhase: (phase: string) => void;
  onClearFilters: () => void;
  hasActiveFilters: boolean;
}

export function FilterBar({
  globalFilter,
  onGlobalFilterChange,
  selectedStatuses,
  onToggleStatus,
  onSelectAllStatuses,
  selectedPriorities,
  onTogglePriority,
  onSelectAllPriorities,
  efforts,
  selectedEffort,
  onToggleEffort,
  onSelectAllEffort,
  selectedTypes,
  onToggleType,
  onSelectAllTypes,
  selectedTags,
  onRemoveTag,
  selectedPhases,
  availablePhases,
  onTogglePhase,
  onClearFilters,
  hasActiveFilters,
}: FilterBarProps) {
  const [filtersOpen, setFiltersOpen] = useState(false);
  const effortBadges = useEffortBadges(efforts);

  return (
    <div className="mb-4 space-y-3">
      <div className="flex items-center gap-3 flex-wrap">
        <input
          type="text"
          value={globalFilter}
          onChange={(e) => onGlobalFilterChange(e.target.value)}
          placeholder="Filter tasks..."
          className="px-3 py-2 border border-gray-300 rounded-md text-sm w-full max-w-xs focus:outline-none focus:ring-2 focus:ring-gray-400 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200"
        />
        <button
          onClick={() => setFiltersOpen((o) => !o)}
          className="cursor-pointer inline-flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-medium text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-200 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        >
          <svg
            className={`w-3.5 h-3.5 transition-transform duration-150 ${filtersOpen ? "rotate-90" : ""}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
          </svg>
          Filters
          {hasActiveFilters && (
            <span
              role="img"
              aria-label="Active filters"
              className="w-1.5 h-1.5 rounded-full bg-blue-500"
            />
          )}
        </button>
        {hasActiveFilters && (
          <button
            onClick={onClearFilters}
            className="min-h-[44px] sm:min-h-0 inline-flex items-center text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 underline"
          >
            Clear filters
          </button>
        )}
      </div>

      {filtersOpen && (
        <div className="space-y-3">
          <FilterRow
            label="Status"
            items={STATUSES}
            selected={selectedStatuses}
            colors={STATUS_COLORS}
            onToggle={onToggleStatus}
            onSelectAll={onSelectAllStatuses}
          />
          <FilterRow
            label="Priority"
            items={PRIORITIES}
            selected={selectedPriorities}
            colors={PRIORITY_COLORS}
            onToggle={onTogglePriority}
            onSelectAll={onSelectAllPriorities}
          />
          <FilterRow
            label="Effort"
            items={efforts}
            selected={selectedEffort}
            colors={effortBadges.colors}
            styles={effortBadges.styles}
            onToggle={onToggleEffort}
            onSelectAll={onSelectAllEffort}
          />
          <FilterRow
            label="Type"
            items={TYPES}
            selected={selectedTypes}
            colors={TYPE_COLORS}
            onToggle={onToggleType}
            onSelectAll={onSelectAllTypes}
          />

          {availablePhases.length > 0 && (
            <div className="flex items-center gap-2 flex-wrap" data-arrow-nav>
              <span className="text-xs text-gray-500 dark:text-gray-400 font-medium">Phase:</span>
              {availablePhases.map((m) => (
                <button
                  key={m}
                  onClick={() => onTogglePhase(m)}
                  className={`min-h-[44px] sm:min-h-0 inline-flex items-center px-2.5 py-1 text-xs rounded-full transition-colors duration-150 ${
                    selectedPhases.has(m) ? `${getPhaseColor(m)} font-medium` : INACTIVE_STYLE
                  }`}
                >
                  {m}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {selectedTags.size > 0 && (
        <div className="flex items-center gap-2 flex-wrap" data-arrow-nav>
          <span className="text-xs text-gray-500 dark:text-gray-400 font-medium">Tags:</span>
          {[...selectedTags].map((tag) => (
            <button
              key={tag}
              onClick={() => onRemoveTag(tag)}
              className="min-h-[44px] sm:min-h-0 px-2 py-0.5 text-xs bg-blue-100 text-blue-700 rounded-full ring-1 ring-blue-300 inline-flex items-center gap-1 transition-colors duration-150 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-300 dark:ring-blue-700 dark:hover:bg-blue-900/50"
            >
              {tag}
              <span className="text-blue-400 dark:text-blue-500">&times;</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
