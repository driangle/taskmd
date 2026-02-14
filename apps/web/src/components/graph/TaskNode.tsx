import { memo } from "react";
import { Handle, Position } from "@xyflow/react";

const STATUS_BG: Record<string, string> = {
  pending: "bg-yellow-50 border-yellow-300 dark:bg-yellow-900/20 dark:border-yellow-700",
  "in-progress": "bg-blue-50 border-blue-300 dark:bg-blue-900/20 dark:border-blue-700",
  completed: "bg-green-50 border-green-300 dark:bg-green-900/20 dark:border-green-700",
  blocked: "bg-red-50 border-red-300 dark:bg-red-900/20 dark:border-red-700",
  cancelled: "bg-gray-50 border-gray-300 dark:bg-gray-800 dark:border-gray-600",
};

const PRIORITY_BORDER: Record<string, string> = {
  critical: "border-l-4 border-l-red-500",
  high: "border-l-4 border-l-orange-400",
  medium: "",
  low: "",
};

interface TaskNodeData {
  label: string;
  status: string;
  priority?: string;
  taskId: string;
  highlighted?: boolean;
  dimmed?: boolean;
}

export const TaskNode = memo(function TaskNode({ data }: { data: TaskNodeData }) {
  const bg = STATUS_BG[data.status] ?? "bg-gray-50 border-gray-300 dark:bg-gray-800 dark:border-gray-600";
  const priorityBorder = data.priority ? (PRIORITY_BORDER[data.priority] ?? "") : "";
  const highlight = data.highlighted ? "ring-2 ring-blue-500" : "";
  const dim = data.dimmed ? "opacity-40" : "";

  return (
    <>
      <Handle type="target" position={Position.Top} className="!bg-gray-400 dark:!bg-gray-500 !w-2 !h-2" />
      <div
        className={`w-[200px] rounded-md border shadow-sm px-3 py-2 cursor-pointer transition-opacity duration-200 ${bg} ${priorityBorder} ${highlight} ${dim}`}
      >
        <div className="text-[10px] text-gray-500 dark:text-gray-400 font-mono">{data.taskId}</div>
        <div className="text-xs font-medium text-gray-800 dark:text-gray-200 truncate" title={data.label}>
          {data.label}
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} className="!bg-gray-400 dark:!bg-gray-500 !w-2 !h-2" />
    </>
  );
});
