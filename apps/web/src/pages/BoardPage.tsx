import { useState, useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { useBoard } from "../hooks/use-board.ts";
import { usePhase } from "../hooks/use-phase.tsx";
import { useProject } from "../hooks/use-project.ts";
import { useConfig } from "../hooks/use-config.ts";
import { updateTask } from "../api/client.ts";
import { BoardView } from "../components/board/BoardView.tsx";
import { BoardFilterBar } from "../components/board/BoardFilterBar.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";
import type { BoardGroup } from "../api/types.ts";
import { STATUSES, PRIORITIES, TYPES } from "../components/tasks/TaskTable/constants.ts";

const baseGroupByOptions = ["status", "priority", "effort", "type", "group", "tag"];

const groupByToField: Record<string, string> = {
  status: "status",
  priority: "priority",
  effort: "effort",
  type: "type",
};

export function BoardPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const { phase } = usePhase();
  const { project } = useProject();
  const { readonly, phases, efforts } = useConfig(project);
  const groupByOptions = useMemo(
    () => phases.length > 0 ? [...baseGroupByOptions, "phase"] : baseGroupByOptions,
    [phases],
  );
  const rawGroupBy = searchParams.get("groupBy") ?? "status";
  const groupBy = groupByOptions.includes(rawGroupBy) ? rawGroupBy : "status";
  const { data, error, isLoading, mutate } = useBoard(groupBy, phase, project);
  const [moveError, setMoveError] = useState<string | null>(null);
  const [moving, setMoving] = useState(false);

  const [selectedStatuses, setSelectedStatuses] = useState(() => new Set(STATUSES));
  const [selectedPriorities, setSelectedPriorities] = useState(() => new Set(PRIORITIES));
  // `null` means "all efforts"; the vocabulary can arrive after first render
  // (static exports fetch the config), so it is resolved at render time.
  const [selectedEffortsOrAll, setSelectedEfforts] = useState<Set<string> | null>(null);
  const selectedEfforts = useMemo(
    () => selectedEffortsOrAll ?? new Set(efforts),
    [selectedEffortsOrAll, efforts],
  );
  const [selectedTypes, setSelectedTypes] = useState(() => new Set(TYPES));
  const [selectedTags, setSelectedTags] = useState<Set<string>>(() => new Set());

  const availableTags = useMemo(() => {
    if (!data) return [];
    const tags = new Set<string>();
    for (const group of data) {
      for (const task of group.tasks) {
        for (const tag of task.tags ?? []) {
          tags.add(tag);
        }
      }
    }
    return [...tags].sort();
  }, [data]);

  const filteredGroups = useMemo((): BoardGroup[] | undefined => {
    if (!data) return undefined;
    return data.map((group) => {
      const filtered = group.tasks.filter((task) => {
        if (groupBy !== "status" && !selectedStatuses.has(task.status)) return false;
        if (groupBy !== "priority" && task.priority && !selectedPriorities.has(task.priority)) return false;
        if (groupBy !== "effort" && task.effort && !selectedEfforts.has(task.effort)) return false;
        if (groupBy !== "type" && task.type && !selectedTypes.has(task.type)) return false;
        if (groupBy !== "tag" && selectedTags.size > 0) {
          const taskTags = task.tags ?? [];
          if (!taskTags.some((t) => selectedTags.has(t))) return false;
        }
        return true;
      });
      return { ...group, tasks: filtered, count: filtered.length };
    });
  }, [data, groupBy, selectedStatuses, selectedPriorities, selectedEfforts, selectedTypes, selectedTags]);

  function handleGroupByChange(value: string) {
    setSearchParams(value === "status" ? {} : { groupBy: value }, {
      replace: true,
    });
  }

  async function handleTaskMove(taskId: string, _sourceGroup: string, targetGroup: string) {
    const field = groupByToField[groupBy];
    if (!field) return;

    setMoveError(null);
    setMoving(true);

    try {
      await updateTask(taskId, { [field]: targetGroup });
      await new Promise((r) => setTimeout(r, 500));
      await mutate();
    } catch (err) {
      setMoveError(
        `Failed to move task ${taskId}: ${err instanceof Error ? err.message : "Unknown error"}`,
      );
    } finally {
      setMoving(false);
    }
  }

  return (
    <div>
      <div className="mb-4 flex items-center gap-4">
        <div>
          <label className="text-xs text-gray-500 mr-2">Group by:</label>
          <select
            value={groupBy}
            onChange={(e) => handleGroupByChange(e.target.value)}
            className="px-2 py-1 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-400"
          >
            {groupByOptions.map((opt) => (
              <option key={opt} value={opt}>
                {opt}
              </option>
            ))}
          </select>
        </div>
        {moveError && (
          <p className="text-sm text-red-600 dark:text-red-400">{moveError}</p>
        )}
      </div>

      {data && (
        <BoardFilterBar
          groupBy={groupBy}
          selectedStatuses={selectedStatuses}
          onStatusesChange={setSelectedStatuses}
          selectedPriorities={selectedPriorities}
          onPrioritiesChange={setSelectedPriorities}
          efforts={efforts}
          selectedEfforts={selectedEfforts}
          onEffortsChange={setSelectedEfforts}
          selectedTypes={selectedTypes}
          onTypesChange={setSelectedTypes}
          selectedTags={selectedTags}
          onTagsChange={setSelectedTags}
          availableTags={availableTags}
        />
      )}

      {isLoading && <LoadingState variant="board" />}
      {error && <ErrorState error={error} onRetry={() => mutate()} />}
      {filteredGroups && filteredGroups.length === 0 && (
        <p className="text-sm text-gray-500 py-8 text-center">
          No tasks to display.
        </p>
      )}
      {filteredGroups && filteredGroups.length > 0 && (
        <div className="relative">
          <BoardView
            groups={filteredGroups}
            groupBy={groupBy}
            readonly={readonly}
            onTaskMove={handleTaskMove}
            showPhase={groupBy !== "phase" && !phase}
          />
          {moving && (
            <div className="absolute inset-0 bg-white/60 dark:bg-gray-900/60 flex items-center justify-center rounded-lg">
              <p className="text-sm text-gray-500 dark:text-gray-400">Updating...</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
