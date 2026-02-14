import { useState } from "react";
import type { Task, TaskUpdateRequest } from "../../api/types.ts";
import {
  STATUSES,
  PRIORITIES,
  EFFORTS,
} from "./TaskTable/constants.ts";

interface TaskEditFormProps {
  task: Task;
  onSave: (data: TaskUpdateRequest) => Promise<void>;
  onCancel: () => void;
  error: string | null;
}

export function TaskEditForm({ task, onSave, onCancel, error }: TaskEditFormProps) {
  const [title, setTitle] = useState(task.title);
  const [status, setStatus] = useState(task.status);
  const [priority, setPriority] = useState(task.priority);
  const [effort, setEffort] = useState(task.effort);
  const [owner, setOwner] = useState(task.owner ?? "");
  const [tags, setTags] = useState((task.tags ?? []).join(", "));
  const [body, setBody] = useState(task.body ?? "");
  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);

    const data: TaskUpdateRequest = {};

    if (title !== task.title) data.title = title;
    if (status !== task.status) data.status = status;
    if (priority !== task.priority) data.priority = priority;
    if (effort !== task.effort) data.effort = effort;
    if (owner !== (task.owner ?? "")) data.owner = owner;

    const newTags = tags
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);
    const oldTags = task.tags ?? [];
    if (JSON.stringify(newTags) !== JSON.stringify(oldTags)) {
      data.tags = newTags;
    }

    if (body !== (task.body ?? "")) data.body = body;

    // Only send if something changed
    if (Object.keys(data).length === 0) {
      onCancel();
      return;
    }

    try {
      await onSave(data);
    } finally {
      setSaving(false);
    }
  };

  const inputClasses = "w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:border-gray-600 dark:bg-gray-700 dark:text-gray-200";

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded text-sm dark:bg-red-900/20 dark:border-red-800 dark:text-red-400">
          {error}
        </div>
      )}

      <div>
        <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
          Title
        </label>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className={inputClasses}
        />
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Status
          </label>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className={inputClasses}
          >
            {STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Priority
          </label>
          <select
            value={priority}
            onChange={(e) => setPriority(e.target.value)}
            className={inputClasses}
          >
            <option value="">-</option>
            {PRIORITIES.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Effort
          </label>
          <select
            value={effort}
            onChange={(e) => setEffort(e.target.value)}
            className={inputClasses}
          >
            <option value="">-</option>
            {EFFORTS.map((e) => (
              <option key={e} value={e}>
                {e}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Owner
          </label>
          <input
            type="text"
            value={owner}
            onChange={(e) => setOwner(e.target.value)}
            placeholder="e.g. alice"
            className={inputClasses}
          />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
            Tags (comma-separated)
          </label>
          <input
            type="text"
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="e.g. backend, api, feature"
            className={inputClasses}
          />
        </div>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
          Body (Markdown)
        </label>
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          rows={24}
          className={`${inputClasses} font-mono`}
        />
      </div>

      <div className="flex gap-2 justify-end">
        <button
          type="button"
          onClick={onCancel}
          disabled={saving}
          className="px-4 py-2 text-sm text-gray-700 bg-white border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50 dark:text-gray-300 dark:bg-gray-700 dark:border-gray-600 dark:hover:bg-gray-600"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={saving}
          className="px-4 py-2 text-sm text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {saving ? "Saving..." : "Save"}
        </button>
      </div>
    </form>
  );
}
