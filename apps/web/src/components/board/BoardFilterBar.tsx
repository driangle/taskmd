import { useState } from "react";
import type { CSSProperties } from "react";
import {
  STATUSES,
  PRIORITIES,
  TYPES,
  STATUS_COLORS,
  PRIORITY_COLORS,
  TYPE_COLORS,
} from "../tasks/TaskTable/constants.ts";
import { useEffortBadges } from "../tasks/TaskTable/effort-colors.ts";
import { toggleInSet } from "../tasks/TaskTable/utils.ts";
import { TagAutocomplete } from "./TagAutocomplete.tsx";

const INACTIVE_PILL =
  "bg-gray-50 border border-gray-200 text-gray-400 hover:bg-gray-100 hover:text-gray-500 dark:bg-gray-800/50 dark:border-gray-700 dark:text-gray-500 dark:hover:bg-gray-700 dark:hover:text-gray-400";


interface PillRowProps {
  label: string;
  items: string[];
  selected: Set<string>;
  colors: Record<string, string>;
  /** Inline styles for values Tailwind cannot class ahead of time (custom effort vocabularies). */
  styles?: Record<string, CSSProperties>;
  onToggle: (value: string) => void;
  onSelectAll: () => void;
}

function PillRow({ label, items, selected, colors, styles, onToggle, onSelectAll }: PillRowProps) {
  const allSelected = selected.size === items.length;
  return (
    <div className="flex items-center gap-2 flex-wrap">
      <span className="text-xs text-gray-500 dark:text-gray-400 font-medium">
        {label}:
      </span>
      <button
        onClick={onSelectAll}
        className={`min-h-[44px] sm:min-h-0 inline-flex items-center px-2.5 py-1 text-xs rounded-full transition-colors duration-150 ${
          allSelected
            ? "bg-gray-200 text-gray-700 font-medium ring-1 ring-gray-300 dark:bg-gray-600 dark:text-gray-200 dark:ring-gray-500"
            : INACTIVE_PILL
        }`}
      >
        all
      </button>
      {items.map((item) => {
        const active = selected.has(item);
        return (
          <button
            key={item}
            onClick={() => onToggle(item)}
            style={active ? styles?.[item] : undefined}
            className={`min-h-[44px] sm:min-h-0 inline-flex items-center px-2.5 py-1 text-xs rounded-full transition-colors duration-150 ${
              active ? (colors[item] ?? "font-medium") : INACTIVE_PILL
            }`}
          >
            {item}
          </button>
        );
      })}
    </div>
  );
}

export interface BoardFilterBarProps {
  groupBy: string;
  selectedStatuses: Set<string>;
  onStatusesChange: (next: Set<string>) => void;
  selectedPriorities: Set<string>;
  onPrioritiesChange: (next: Set<string>) => void;
  /** The project's configured effort vocabulary. */
  efforts: string[];
  selectedEfforts: Set<string>;
  onEffortsChange: (next: Set<string>) => void;
  selectedTypes: Set<string>;
  onTypesChange: (next: Set<string>) => void;
  selectedTags: Set<string>;
  onTagsChange: (next: Set<string>) => void;
  availableTags: string[];
}

export function BoardFilterBar({
  groupBy,
  selectedStatuses,
  onStatusesChange,
  selectedPriorities,
  onPrioritiesChange,
  efforts,
  selectedEfforts,
  onEffortsChange,
  selectedTypes,
  onTypesChange,
  selectedTags,
  onTagsChange,
  availableTags,
}: BoardFilterBarProps) {
  const [filtersOpen, setFiltersOpen] = useState(false);
  const effortBadges = useEffortBadges(efforts);

  const hasActiveFilters =
    (groupBy !== "status" && selectedStatuses.size !== STATUSES.length) ||
    (groupBy !== "priority" && selectedPriorities.size !== PRIORITIES.length) ||
    (groupBy !== "effort" && selectedEfforts.size !== efforts.length) ||
    (groupBy !== "type" && selectedTypes.size !== TYPES.length) ||
    selectedTags.size > 0;

  function handleClearFilters() {
    onStatusesChange(new Set(STATUSES));
    onPrioritiesChange(new Set(PRIORITIES));
    onEffortsChange(new Set(efforts));
    onTypesChange(new Set(TYPES));
    onTagsChange(new Set());
  }

  return (
    <div className="mb-4 space-y-2">
      <div className="flex items-center gap-3 flex-wrap">
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
            <span className="w-1.5 h-1.5 rounded-full bg-blue-500" />
          )}
        </button>
        {hasActiveFilters && (
          <button
            onClick={handleClearFilters}
            className="min-h-[44px] sm:min-h-0 inline-flex items-center text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 underline"
          >
            Clear filters
          </button>
        )}
      </div>

      {filtersOpen && (
        <div className="space-y-2">
          {groupBy !== "status" && (
            <PillRow
              label="Status"
              items={STATUSES}
              selected={selectedStatuses}
              colors={STATUS_COLORS}
              onToggle={(s) => onStatusesChange(toggleInSet(selectedStatuses, s))}
              onSelectAll={() => onStatusesChange(new Set(STATUSES))}
            />
          )}
          {groupBy !== "priority" && (
            <PillRow
              label="Priority"
              items={PRIORITIES}
              selected={selectedPriorities}
              colors={PRIORITY_COLORS}
              onToggle={(p) => onPrioritiesChange(toggleInSet(selectedPriorities, p))}
              onSelectAll={() => onPrioritiesChange(new Set(PRIORITIES))}
            />
          )}
          {groupBy !== "effort" && (
            <PillRow
              label="Effort"
              items={efforts}
              selected={selectedEfforts}
              colors={effortBadges.colors}
              styles={effortBadges.styles}
              onToggle={(e) => onEffortsChange(toggleInSet(selectedEfforts, e))}
              onSelectAll={() => onEffortsChange(new Set(efforts))}
            />
          )}
          {groupBy !== "type" && (
            <PillRow
              label="Type"
              items={TYPES}
              selected={selectedTypes}
              colors={TYPE_COLORS}
              onToggle={(ty) => onTypesChange(toggleInSet(selectedTypes, ty))}
              onSelectAll={() => onTypesChange(new Set(TYPES))}
            />
          )}
          {groupBy !== "tag" && availableTags.length > 0 && (
            <TagAutocomplete
              availableTags={availableTags}
              selectedTags={selectedTags}
              onTagsChange={onTagsChange}
            />
          )}
        </div>
      )}
    </div>
  );
}
