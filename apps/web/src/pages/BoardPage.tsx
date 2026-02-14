import { useSearchParams } from "react-router-dom";
import { useBoard } from "../hooks/use-board.ts";
import { BoardView } from "../components/board/BoardView.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";

const groupByOptions = ["status", "priority", "effort", "group", "tag"];

export function BoardPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const groupBy = searchParams.get("groupBy") ?? "status";
  const { data, error, isLoading, mutate } = useBoard(groupBy);

  function handleGroupByChange(value: string) {
    setSearchParams(value === "status" ? {} : { groupBy: value }, {
      replace: true,
    });
  }

  return (
    <div>
      <div className="mb-4">
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

      {isLoading && <LoadingState variant="board" />}
      {error && <ErrorState error={error} onRetry={() => mutate()} />}
      {data && data.length === 0 && (
        <p className="text-sm text-gray-500 py-8 text-center">
          No tasks to display.
        </p>
      )}
      {data && data.length > 0 && <BoardView groups={data} />}
    </div>
  );
}
