import { Link } from "react-router-dom";
import type { Task } from "../../api/types.ts";
import { PhaseBadge } from "./TaskTable/Badges.tsx";

/** The detail page's metadata grid: priority, effort, phase, owner, etc. */
export function TaskMetaGrid({ task }: { task: Task }) {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6 text-sm">
      {task.priority && <Field label="Priority" value={task.priority} />}
      {task.effort && <Field label="Effort" value={task.effort} />}
      {task.phase && (
        <div>
          <dt className="text-xs text-gray-500 dark:text-gray-400">Phase</dt>
          <dd className="mt-0.5"><PhaseBadge phase={task.phase} /></dd>
        </div>
      )}
      {task.owner && <Field label="Owner" value={task.owner} />}
      {task.group && <Field label="Group" value={task.group} />}
      {task.parent && (
        <div>
          <dt className="text-xs text-gray-500 dark:text-gray-400">Parent</dt>
          <dd className="font-medium">
            <Link
              to={`/tasks/${task.parent}`}
              className="text-blue-600 hover:underline dark:text-blue-400 font-mono"
            >
              {task.parent}
            </Link>
          </dd>
        </div>
      )}
      {task.created && <Field label="Created" value={task.created} />}
    </div>
  );
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs text-gray-500 dark:text-gray-400">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  );
}
