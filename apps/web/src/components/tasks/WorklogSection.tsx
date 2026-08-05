import { useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import { addWorklogEntry, ApiRequestError } from "../../api/client.ts";
import type { WorklogEntry } from "../../api/types.ts";

interface WorklogSectionProps {
  entries: WorklogEntry[];
  taskId?: string;
  readonly?: boolean;
  /** Called after a successful post so the caller can revalidate. */
  onAdded?: () => void;
}

export function WorklogSection({
  entries,
  taskId,
  readonly,
  onAdded,
}: WorklogSectionProps) {
  const [content, setContent] = useState("");
  const [author, setAuthor] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canPost = !!taskId && !readonly;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!taskId || !content.trim() || submitting) return;
    setSubmitting(true);
    setError(null);
    try {
      await addWorklogEntry(taskId, {
        author: author.trim() || undefined,
        content: content.trim(),
      });
      setContent("");
      onAdded?.();
    } catch (err) {
      setError(
        err instanceof ApiRequestError
          ? err.message
          : "Failed to add worklog entry",
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="border-t border-gray-200 dark:border-gray-700 pt-4 mt-4">
      <h3 className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-3">
        Worklog
        <span className="ml-2 px-1.5 py-0.5 text-xs bg-gray-100 rounded dark:bg-gray-700">
          {entries.length}
        </span>
      </h3>
      <div className="space-y-4">
        {entries.map((entry, i) => (
          <div
            key={i}
            className="border-l-2 border-gray-200 dark:border-gray-600 pl-4"
          >
            <div className="flex items-center gap-2">
              {entry.author && (
                <span className="text-xs font-medium text-gray-700 dark:text-gray-200">
                  {entry.author}
                </span>
              )}
              <time className="text-xs text-gray-400 font-mono">
                {new Date(entry.timestamp).toLocaleString()}
              </time>
            </div>
            <div className="prose prose-sm max-w-none dark:prose-invert mt-1">
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                rehypePlugins={[rehypeRaw]}
              >
                {entry.content}
              </ReactMarkdown>
            </div>
          </div>
        ))}
        {entries.length === 0 && (
          <p className="text-xs text-gray-400">No worklog entries yet.</p>
        )}
      </div>

      {canPost && (
        <form onSubmit={handleSubmit} className="mt-4 space-y-2">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Add a comment… (markdown supported)"
            rows={3}
            className="w-full rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 p-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={author}
              onChange={(e) => setAuthor(e.target.value)}
              placeholder="Author (optional)"
              className="flex-1 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
            <button
              type="submit"
              disabled={!content.trim() || submitting}
              className="rounded bg-blue-600 px-3 py-1 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {submitting ? "Adding…" : "Comment"}
            </button>
          </div>
          {error && (
            <p className="text-xs text-red-500" role="alert">
              {error}
            </p>
          )}
        </form>
      )}
    </div>
  );
}
