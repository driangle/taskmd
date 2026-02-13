import { useState } from "react";
import { useParams, Link } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import { useTaskDetail } from "../hooks/use-task-detail.ts";
import { updateTask, ApiRequestError } from "../api/client.ts";
import type { TaskUpdateRequest } from "../api/types.ts";
import { TaskEditForm } from "../components/tasks/TaskEditForm.tsx";

export function TaskDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data: task, error, isLoading, mutate } = useTaskDetail(id);
  const [isEditing, setIsEditing] = useState(false);
  const [editError, setEditError] = useState<string | null>(null);

  if (isLoading) return <p className="text-sm text-gray-500">Loading...</p>;
  if (error)
    return <p className="text-sm text-red-600">Error: {error.message}</p>;

  if (!task) {
    return (
      <div>
        <p className="text-sm text-gray-500">Task not found: {id}</p>
        <Link to="/tasks" className="text-sm text-blue-600 hover:underline">
          Back to tasks
        </Link>
      </div>
    );
  }

  const handleSave = async (data: TaskUpdateRequest) => {
    setEditError(null);
    try {
      const updated = await updateTask(task.id, data);
      await mutate(updated, false);
      setIsEditing(false);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        const msg = err.details?.length
          ? `${err.message}: ${err.details.join(", ")}`
          : err.message;
        setEditError(msg);
      } else {
        setEditError("An unexpected error occurred.");
      }
    }
  };

  return (
    <div>
      <Link
        to="/tasks"
        className="text-sm text-gray-500 hover:text-gray-700 mb-4 inline-block"
      >
        &larr; All tasks
      </Link>

      <div className="bg-white border border-gray-200 rounded-lg p-6">
        {isEditing ? (
          <TaskEditForm
            task={task}
            onSave={handleSave}
            onCancel={() => {
              setIsEditing(false);
              setEditError(null);
            }}
            error={editError}
          />
        ) : (
          <>
            <div className="flex items-start justify-between mb-4">
              <div>
                <span className="font-mono text-xs text-gray-400">
                  {task.id}
                </span>
                <h2 className="text-xl font-semibold mt-1">{task.title}</h2>
              </div>
              <div className="flex items-center gap-2">
                <StatusBadge status={task.status} />
                <button
                  onClick={() => setIsEditing(true)}
                  className="px-3 py-1 text-xs font-medium text-gray-600 bg-gray-100 rounded hover:bg-gray-200"
                >
                  Edit
                </button>
              </div>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6 text-sm">
              {task.priority && (
                <Field label="Priority" value={task.priority} />
              )}
              {task.effort && <Field label="Effort" value={task.effort} />}
              {task.group && <Field label="Group" value={task.group} />}
              {task.created && <Field label="Created" value={task.created} />}
            </div>

            {task.dependencies && task.dependencies.length > 0 && (
              <div className="mb-6">
                <h3 className="text-xs font-medium text-gray-500 uppercase mb-2">
                  Dependencies
                </h3>
                <div className="flex gap-2 flex-wrap">
                  {task.dependencies.map((dep) => (
                    <Link
                      key={dep}
                      to={`/tasks/${dep}`}
                      className="px-2 py-1 text-xs font-mono bg-gray-100 rounded hover:bg-gray-200"
                    >
                      {dep}
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {task.tags && task.tags.length > 0 && (
              <div className="mb-6">
                <h3 className="text-xs font-medium text-gray-500 uppercase mb-2">
                  Tags
                </h3>
                <div className="flex gap-1 flex-wrap">
                  {task.tags.map((t) => (
                    <span
                      key={t}
                      className="px-1.5 py-0.5 text-xs bg-gray-100 rounded"
                    >
                      {t}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {task.body && (
              <div className="border-t border-gray-200 pt-4">
                <div className="prose prose-sm max-w-none">
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    rehypePlugins={[rehypeRaw]}
                  >
                    {task.body}
                  </ReactMarkdown>
                </div>
              </div>
            )}

            {task.file_path && (
              <div className="mt-4 text-xs text-gray-400 font-mono">
                {task.file_path}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs text-gray-500">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    pending: "bg-yellow-100 text-yellow-800",
    "in-progress": "bg-blue-100 text-blue-800",
    completed: "bg-green-100 text-green-800",
    blocked: "bg-red-100 text-red-800",
  };
  return (
    <span
      className={`px-2 py-0.5 text-xs font-medium rounded-full ${colors[status] ?? "bg-gray-100 text-gray-800"}`}
    >
      {status}
    </span>
  );
}
