import { useState } from "react";
import { useBoard } from "../hooks/use-board.ts";
import { BoardView } from "../components/board/BoardView.tsx";

const groupByOptions = ["status", "priority", "effort", "group", "tag"];

export function BoardPage() {
  const [groupBy, setGroupBy] = useState("status");
  const { data, error, isLoading } = useBoard(groupBy);

  return (
    <div>
      <div className="mb-4">
        <label className="text-xs text-gray-500 mr-2">Group by:</label>
        <select
          value={groupBy}
          onChange={(e) => setGroupBy(e.target.value)}
          className="px-2 py-1 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-400"
        >
          {groupByOptions.map((opt) => (
            <option key={opt} value={opt}>
              {opt}
            </option>
          ))}
        </select>
      </div>

      {isLoading && <p className="text-sm text-gray-500">Loading...</p>}
      {error && <p className="text-sm text-red-600">Error: {error.message}</p>}
      {data && <BoardView groups={data} />}
    </div>
  );
}
