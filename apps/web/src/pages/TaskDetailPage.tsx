import { useState } from "react";
import { useParams, Link } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import { useTaskDetail } from "../hooks/use-task-detail.ts";
import { useConfig } from "../hooks/use-config.ts";
import { updateTask, ApiRequestError } from "../api/client.ts";
import type { TaskUpdateRequest } from "../api/types.ts";
import { TaskEditForm } from "../components/tasks/TaskEditForm.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";
import { StatusBadge } from "../components/tasks/TaskTable/Badges.tsx";

export function TaskDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data: task, error, isLoading, mutate } = useTaskDetail(id);
  const { readonly } = useConfig();
  const [isEditing, setIsEditing] = useState(false);
  const [editError, setEditError] = useState<string | null>(null);

  if (isLoading) return <LoadingState variant="detail" />;
  if (error) return <ErrorState error={error} onRetry={() => mutate()} />;

  if (!task) {
    return (
      <div>
        <p className="text-sm text-gray-500 dark:text-gray-400">Task not found: {id}</p>
        <Link to="/tasks" className="text-sm text-blue-600 hover:underline dark:text-blue-400">
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
      <div className="bg-white border border-gray-200 rounded-lg p-6 dark:bg-gray-800 dark:border-gray-700">
        {isEditing && !readonly ? (
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
                {!readonly && (
                  <button
                    onClick={() => setIsEditing(true)}
                    className="px-3 py-1 text-xs font-medium text-gray-600 bg-gray-100 rounded hover:bg-gray-200 dark:text-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600"
                  >
                    Edit
                  </button>
                )}
              </div>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6 text-sm">
              {task.priority && (
                <Field label="Priority" value={task.priority} />
              )}
              {task.effort && <Field label="Effort" value={task.effort} />}
              {task.owner && <Field label="Owner" value={task.owner} />}
              {task.group && <Field label="Group" value={task.group} />}
              {task.created && <Field label="Created" value={task.created} />}
            </div>

            {task.dependencies && task.dependencies.length > 0 && (
              <div className="mb-6">
                <h3 className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-2">
                  Dependencies
                </h3>
                <div className="flex gap-2 flex-wrap">
                  {task.dependencies.map((dep) => (
                    <Link
                      key={dep}
                      to={`/tasks/${dep}`}
                      className="px-2 py-1 text-xs font-mono bg-gray-100 rounded hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600"
                    >
                      {dep}
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {task.tags && task.tags.length > 0 && (
              <div className="mb-6">
                <h3 className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-2">
                  Tags
                </h3>
                <div className="flex gap-1 flex-wrap">
                  {task.tags.map((t) => (
                    <Link
                      key={t}
                      to={`/tasks?tag=${encodeURIComponent(t)}`}
                      className="px-1.5 py-0.5 text-xs bg-gray-100 rounded cursor-pointer hover:bg-gray-200 transition-colors duration-150 dark:bg-gray-700 dark:hover:bg-gray-600"
                    >
                      {t}
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {task.body && (
              <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
                <div className="prose prose-sm max-w-none dark:prose-invert">
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
      <dt className="text-xs text-gray-500 dark:text-gray-400">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  );
}
