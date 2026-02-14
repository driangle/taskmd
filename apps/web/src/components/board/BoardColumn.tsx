import { useRef, useState } from "react";
import type { BoardGroup } from "../../api/types.ts";
import { TaskCard } from "./TaskCard.tsx";

const statusColors: Record<string, string> = {
  pending: "border-yellow-300 bg-yellow-50 dark:bg-yellow-900/20",
  "in-progress": "border-blue-300 bg-blue-50 dark:bg-blue-900/20",
  completed: "border-green-300 bg-green-50 dark:bg-green-900/20",
  blocked: "border-red-300 bg-red-50 dark:bg-red-900/20",
};

interface BoardColumnProps {
  group: BoardGroup;
  canDrag: boolean;
  onTaskDrop?: (taskId: string, sourceGroup: string, targetGroup: string) => void;
}

export function BoardColumn({ group, canDrag, onTaskDrop }: BoardColumnProps) {
  const [dragOver, setDragOver] = useState(false);
  const dragOverRef = useRef(false);

  function handleDragOver(e: React.DragEvent) {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    if (!dragOverRef.current) {
      dragOverRef.current = true;
      setDragOver(true);
    }
  }

  function handleDragLeave(e: React.DragEvent) {
    if (e.currentTarget.contains(e.relatedTarget as Node)) return;
    dragOverRef.current = false;
    setDragOver(false);
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault();
    dragOverRef.current = false;
    setDragOver(false);
    const taskId = e.dataTransfer.getData("text/plain");
    const sourceGroup = e.dataTransfer.getData("application/x-source-group");
    if (!taskId || sourceGroup === group.group) return;
    onTaskDrop?.(taskId, sourceGroup, group.group);
  }

  return (
    <div
      onDragOver={canDrag ? handleDragOver : undefined}
      onDragLeave={canDrag ? handleDragLeave : undefined}
      onDrop={canDrag ? handleDrop : undefined}
      className={`flex-shrink-0 w-full md:w-72 rounded-lg border-t-4 bg-white shadow-sm dark:bg-gray-800 transition-shadow ${
        statusColors[group.group] ?? "border-gray-300 bg-gray-50 dark:bg-gray-800"
      } ${dragOver ? "ring-2 ring-blue-400 dark:ring-blue-500" : ""}`}
    >
      <div className="px-4 py-3 border-b border-gray-100 dark:border-gray-700">
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200">
          {group.group}{" "}
          <span className="text-gray-400 dark:text-gray-500 font-normal">
            ({group.count})
          </span>
        </h3>
      </div>
      <div className="p-2 space-y-2">
        {group.tasks.map((task) => (
          <TaskCard
            key={task.id}
            task={task}
            sourceGroup={group.group}
            canDrag={canDrag}
          />
        ))}
        {group.tasks.length === 0 && (
          <p className="text-xs text-gray-400 dark:text-gray-500 p-2">No tasks</p>
        )}
      </div>
    </div>
  );
}
