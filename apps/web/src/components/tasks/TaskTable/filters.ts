import type { Task } from "../../../api/types.ts";
import { STATUSES, PRIORITIES, DEFAULT_EFFORTS, TYPES } from "./constants.ts";
import { effectiveStatus } from "./utils.ts";

export interface FilterState {
  selectedStatuses: Set<string>;
  selectedPriorities: Set<string>;
  selectedTypes: Set<string>;
  selectedTags: Set<string>;
  selectedEffort: Set<string>;
  selectedPhases: Set<string>;
  globalFilter: string;
}

/**
 * `efforts` is the project's configured effort vocabulary; it decides when the
 * effort filter counts as "all selected" and therefore inactive.
 */
export function applyFilters(
  tasks: Task[],
  filters: FilterState,
  efforts: string[] = DEFAULT_EFFORTS,
): Task[] {
  return tasks.filter((task) => {
    if (!filters.selectedStatuses.has(effectiveStatus(task))) return false;
    if (task.priority && !filters.selectedPriorities.has(task.priority))
      return false;
    if (task.type && !filters.selectedTypes.has(task.type)) return false;
    if (filters.selectedTags.size > 0) {
      if (!task.tags || !task.tags.some((t) => filters.selectedTags.has(t)))
        return false;
    }
    if (filters.selectedEffort.size < efforts.length) {
      if (!task.effort || !filters.selectedEffort.has(task.effort)) return false;
    }
    if (filters.selectedPhases.size > 0) {
      if (!task.phase || !filters.selectedPhases.has(task.phase))
        return false;
    }
    return true;
  });
}

export function hasActiveFilters(
  filters: FilterState,
  efforts: string[] = DEFAULT_EFFORTS,
): boolean {
  return (
    filters.selectedStatuses.size !== STATUSES.length ||
    filters.selectedPriorities.size !== PRIORITIES.length ||
    filters.selectedTypes.size !== TYPES.length ||
    filters.selectedTags.size > 0 ||
    filters.selectedEffort.size !== efforts.length ||
    filters.selectedPhases.size > 0 ||
    filters.globalFilter !== ""
  );
}

export function defaultFilterState(
  efforts: string[] = DEFAULT_EFFORTS,
): FilterState {
  return {
    selectedStatuses: new Set(STATUSES),
    selectedPriorities: new Set(PRIORITIES),
    selectedTypes: new Set(TYPES),
    selectedTags: new Set<string>(),
    selectedEffort: new Set(efforts),
    selectedPhases: new Set<string>(),
    globalFilter: "",
  };
}
