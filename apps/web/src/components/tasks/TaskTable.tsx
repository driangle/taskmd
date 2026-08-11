import {
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useReactTable,
  type SortingState,
} from "@tanstack/react-table";
import { useState, useMemo, useCallback } from "react";
import { useSearchParams } from "react-router-dom";
import type { Task } from "../../api/types.ts";
import { usePhase } from "../../hooks/use-phase.tsx";
import { useConfig } from "../../hooks/use-config.ts";
import { STATUSES, PRIORITIES, TYPES } from "./TaskTable/constants.ts";
import { FilterBar } from "./TaskTable/FilterBar.tsx";
import { createTaskColumns } from "./TaskTable/columns.tsx";
import { toggleInSet } from "./TaskTable/utils.ts";
import { applyFilters, hasActiveFilters as checkActiveFilters } from "./TaskTable/filters.ts";
import { MobileCardList } from "./TaskTable/MobileCardList.tsx";
import { DesktopTable } from "./TaskTable/DesktopTable.tsx";

interface TaskTableProps {
  tasks: Task[];
  initialTags?: string[];
  initialStatuses?: string[];
  initialPriorities?: string[];
  initialEffort?: string[];
  initialTypes?: string[];
  initialPhases?: string[];
  project?: string | null;
}

export function TaskTable({ tasks, initialTags, initialStatuses, initialPriorities, initialEffort, initialTypes, initialPhases, project }: TaskTableProps) {
  const [, setSearchParams] = useSearchParams();
  const { efforts } = useConfig(project);
  const [sorting, setSorting] = useState<SortingState>([]);
  const [globalFilter, setGlobalFilter] = useState("");
  const [selectedStatuses, setSelectedStatuses] = useState<Set<string>>(
    () => initialStatuses && initialStatuses.length > 0 ? new Set(initialStatuses) : new Set(STATUSES),
  );
  const [selectedPriorities, setSelectedPriorities] = useState<Set<string>>(
    () => initialPriorities && initialPriorities.length > 0 ? new Set(initialPriorities) : new Set(PRIORITIES),
  );
  const [selectedTypes, setSelectedTypes] = useState<Set<string>>(
    () => initialTypes && initialTypes.length > 0 ? new Set(initialTypes) : new Set(TYPES),
  );
  const [selectedTags, setSelectedTags] = useState<Set<string>>(
    () => new Set(initialTags),
  );
  // `null` means "all efforts", resolved against the live vocabulary below. The
  // config can arrive after first render (static exports fetch it), so freezing
  // a Set of the fallback vocabulary here would strand the wrong values.
  const [selectedEffortOrAll, setSelectedEffort] = useState<Set<string> | null>(
    () => initialEffort && initialEffort.length > 0 ? new Set(initialEffort) : null,
  );
  const selectedEffort = useMemo(
    () => selectedEffortOrAll ?? new Set(efforts),
    [selectedEffortOrAll, efforts],
  );
  const [selectedPhases, setSelectedPhases] = useState<Set<string>>(
    () => initialPhases && initialPhases.length > 0 ? new Set(initialPhases) : new Set<string>(),
  );

  const { phase: globalPhase } = usePhase();

  const availablePhases = useMemo(() => {
    const phases = new Set<string>();
    for (const task of tasks) {
      if (task.phase) phases.add(task.phase);
    }
    return [...phases].sort();
  }, [tasks]);

  // Hide phase badges when filtered to a single phase (global selector or local filter)
  const showPhase = !globalPhase && !(selectedPhases.size === 1);

  const filterState = { selectedStatuses, selectedPriorities, selectedTypes, selectedTags, selectedEffort, selectedPhases, globalFilter };
  const hasActiveFilters = checkActiveFilters(filterState, efforts);

  const syncFiltersToUrl = useCallback(
    (updates: { tag?: Set<string>; status?: Set<string>; priority?: Set<string>; effort?: Set<string>; type?: Set<string>; phase?: Set<string> }) => {
      setSearchParams(
        (prev) => {
          for (const [param, values] of Object.entries(updates)) {
            prev.delete(param);
            if (values) {
              values.forEach((v: string) => prev.append(param, v));
            }
          }
          return prev;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  function clearFilters() {
    setSelectedStatuses(new Set(STATUSES));
    setSelectedPriorities(new Set(PRIORITIES));
    setSelectedTypes(new Set(TYPES));
    setSelectedTags(new Set());
    setSelectedEffort(null);
    setSelectedPhases(new Set());
    syncFiltersToUrl({ tag: new Set(), status: new Set(), priority: new Set(), effort: new Set(), type: new Set(), phase: new Set() });
    setGlobalFilter("");
  }

  const toggleTag = useCallback((tag: string) => {
    setSelectedTags((prev) => {
      const next = toggleInSet(prev, tag);
      syncFiltersToUrl({ tag: next });
      return next;
    });
  }, [syncFiltersToUrl]);

  const filteredTasks = useMemo(
    () => applyFilters(tasks, filterState, efforts),
    // eslint-disable-next-line react-hooks/exhaustive-deps -- filterState is derived from these individual deps
    [tasks, selectedStatuses, selectedPriorities, selectedTypes, selectedTags, selectedEffort, selectedPhases, globalFilter, efforts],
  );

  const taskStatusMap = useMemo(
    () => new Map(tasks.map((t) => [t.id, t.status])),
    [tasks],
  );

  const columns = useMemo(
    () => createTaskColumns(selectedTags, toggleTag, taskStatusMap, showPhase),
    [selectedTags, toggleTag, taskStatusMap, showPhase],
  );

  const table = useReactTable({
    data: filteredTasks,
    columns,
    state: { sorting, globalFilter },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  const rows = table.getRowModel().rows;
  const visibleCount = rows.length;

  return (
    <div>
      <FilterBar
        globalFilter={globalFilter}
        onGlobalFilterChange={setGlobalFilter}
        selectedStatuses={selectedStatuses}
        onToggleStatus={(s) =>
          setSelectedStatuses((prev) => {
            const next = toggleInSet(prev, s);
            syncFiltersToUrl({ status: next.size === STATUSES.length ? new Set() : next });
            return next;
          })
        }
        onSelectAllStatuses={() => {
          setSelectedStatuses(new Set(STATUSES));
          syncFiltersToUrl({ status: new Set() });
        }}
        selectedPriorities={selectedPriorities}
        onTogglePriority={(p) =>
          setSelectedPriorities((prev) => {
            const next = toggleInSet(prev, p);
            syncFiltersToUrl({ priority: next.size === PRIORITIES.length ? new Set() : next });
            return next;
          })
        }
        onSelectAllPriorities={() => {
          setSelectedPriorities(new Set(PRIORITIES));
          syncFiltersToUrl({ priority: new Set() });
        }}
        efforts={efforts}
        selectedEffort={selectedEffort}
        onToggleEffort={(e) => {
          const next = toggleInSet(selectedEffort, e);
          const isAll = next.size === efforts.length;
          setSelectedEffort(isAll ? null : next);
          syncFiltersToUrl({ effort: isAll ? new Set() : next });
        }}
        onSelectAllEffort={() => {
          setSelectedEffort(null);
          syncFiltersToUrl({ effort: new Set() });
        }}
        selectedTypes={selectedTypes}
        onToggleType={(ty) =>
          setSelectedTypes((prev) => {
            const next = toggleInSet(prev, ty);
            syncFiltersToUrl({ type: next.size === TYPES.length ? new Set() : next });
            return next;
          })
        }
        onSelectAllTypes={() => {
          setSelectedTypes(new Set(TYPES));
          syncFiltersToUrl({ type: new Set() });
        }}
        selectedTags={selectedTags}
        onRemoveTag={toggleTag}
        selectedPhases={selectedPhases}
        availablePhases={availablePhases}
        onTogglePhase={(m) =>
          setSelectedPhases((prev) => {
            const next = toggleInSet(prev, m);
            syncFiltersToUrl({ phase: next });
            return next;
          })
        }
        onClearFilters={clearFilters}
        hasActiveFilters={hasActiveFilters}
      />
      <MobileCardList rows={rows} onClearFilters={clearFilters} showPhase={showPhase} />
      <DesktopTable table={table} rows={rows} columns={columns} clearFilters={clearFilters} />
      <p className="mt-2 text-xs text-gray-400">
        Showing {visibleCount} of {tasks.length} tasks
      </p>
    </div>
  );
}
