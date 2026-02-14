import type { GraphData } from "../../api/types.ts";

interface GraphStatsProps {
  data: GraphData;
  visibleCount: number;
}

export function GraphStats({ data, visibleCount }: GraphStatsProps) {
  const totalCount = data.nodes.length;
  const blockedCount = data.nodes.filter((n) => n.status === "blocked").length;
  const hasCycles = data.cycles && data.cycles.length > 0;

  return (
    <div className="flex items-center gap-4 text-xs text-gray-500 flex-wrap">
      <span>
        Showing <span className="font-medium text-gray-700">{visibleCount}</span> of{" "}
        <span className="font-medium text-gray-700">{totalCount}</span> tasks
      </span>
      {blockedCount > 0 && (
        <span className="text-red-600">
          {blockedCount} blocked
        </span>
      )}
      {hasCycles && (
        <span className="text-amber-600 font-medium">
          Circular dependencies detected
        </span>
      )}
    </div>
  );
}
