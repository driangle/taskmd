import { useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useBoard } from "../hooks/use-board.ts";
import { useConfig } from "../hooks/use-config.ts";
import { updateTask } from "../api/client.ts";
import { BoardView } from "../components/board/BoardView.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";

const groupByOptions = ["status", "priority", "effort", "group", "tag"];

const groupByToField: Record<string, string> = {
  status: "status",
  priority: "priority",
  effort: "effort",
};

export function BoardPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const groupBy = searchParams.get("groupBy") ?? "status";
  const { data, error, isLoading, mutate } = useBoard(groupBy);
  const { readonly } = useConfig();
  const [moveError, setMoveError] = useState<string | null>(null);
  const [moving, setMoving] = useState(false);

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

      {isLoading && <LoadingState variant="board" />}
      {error && <ErrorState error={error} onRetry={() => mutate()} />}
      {data && data.length === 0 && (
        <p className="text-sm text-gray-500 py-8 text-center">
          No tasks to display.
        </p>
      )}
      {data && data.length > 0 && (
        <div className="relative">
          <BoardView
            groups={data}
            groupBy={groupBy}
            readonly={readonly}
            onTaskMove={handleTaskMove}
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
