import type { BoardGroup } from "../../api/types.ts";
import { BoardColumn } from "./BoardColumn.tsx";

const draggableGroupByFields = ["status", "priority", "effort"];

interface BoardViewProps {
  groups: BoardGroup[];
  groupBy: string;
  readonly: boolean;
  onTaskMove?: (taskId: string, sourceGroup: string, targetGroup: string) => void;
}

export function BoardView({ groups, groupBy, readonly, onTaskMove }: BoardViewProps) {
  const canDrag = !readonly && draggableGroupByFields.includes(groupBy);

  return (
    <div
      className="flex flex-col md:flex-row gap-4 md:overflow-x-auto pb-4"
      onDragOver={(e) => e.preventDefault()}
      onDrop={(e) => e.preventDefault()}
    >
      {groups.map((g) => (
        <BoardColumn
          key={g.group}
          group={g}
          canDrag={canDrag}
          onTaskDrop={onTaskMove}
        />
      ))}
    </div>
  );
}
