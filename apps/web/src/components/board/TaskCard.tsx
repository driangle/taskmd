import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import type { BoardTask } from "../../api/types.ts";

interface TaskCardProps {
  task: BoardTask;
  sourceGroup: string;
  canDrag: boolean;
}

// Prevent browser default drop behavior (navigation) anywhere on the page
function preventNavigation(e: Event) {
  e.preventDefault();
}

export function TaskCard({ task, sourceGroup, canDrag }: TaskCardProps) {
  const guardActive = useRef(false);
  const [dragging, setDragging] = useState(false);

  useEffect(() => {
    return () => {
      if (guardActive.current) {
        document.removeEventListener("dragover", preventNavigation);
        document.removeEventListener("drop", preventNavigation);
      }
    };
  }, []);

  function handleDragStart(e: React.DragEvent<HTMLDivElement>) {
    e.dataTransfer.setData("text/plain", task.id);
    e.dataTransfer.setData("application/x-source-group", sourceGroup);
    e.dataTransfer.effectAllowed = "move";
    setDragging(true);

    guardActive.current = true;
    document.addEventListener("dragover", preventNavigation);
    document.addEventListener("drop", preventNavigation);
  }

  function handleDragEnd() {
    setDragging(false);
    guardActive.current = false;
    document.removeEventListener("dragover", preventNavigation);
    document.removeEventListener("drop", preventNavigation);
  }

  return (
    <div
      draggable={canDrag}
      onDragStart={canDrag ? handleDragStart : undefined}
      onDragEnd={canDrag ? handleDragEnd : undefined}
      className={`p-3 bg-white rounded border border-gray-100 shadow-sm dark:bg-gray-800/50 dark:border-gray-700 group transition-opacity ${dragging ? "opacity-50" : ""}`}
    >
      <div className="flex items-start justify-between gap-2">
        {canDrag && (
          <span
            className="cursor-grab text-gray-300 dark:text-gray-600 opacity-0 group-hover:opacity-100 transition-opacity shrink-0 mt-0.5"
            style={{ touchAction: "none" }}
            aria-hidden="true"
          >
            <svg width="12" height="12" viewBox="0 0 12 12" fill="currentColor">
              <circle cx="3.5" cy="2" r="1.2" />
              <circle cx="8.5" cy="2" r="1.2" />
              <circle cx="3.5" cy="6" r="1.2" />
              <circle cx="8.5" cy="6" r="1.2" />
              <circle cx="3.5" cy="10" r="1.2" />
              <circle cx="8.5" cy="10" r="1.2" />
            </svg>
          </span>
        )}
        <Link
          to={`/tasks/${task.id}`}
          draggable={false}
          className="text-sm font-medium leading-snug text-blue-600 hover:underline dark:text-blue-400 flex-1"
        >
          {task.title}
        </Link>
        <span className="text-xs text-gray-400 dark:text-gray-500 font-mono shrink-0">
          {task.id}
        </span>
      </div>
      {task.priority && (
        <span className={`mt-1.5 inline-block text-xs text-gray-500 dark:text-gray-400 ${canDrag ? "ml-5" : ""}`}>
          {task.priority}
        </span>
      )}
    </div>
  );
}
