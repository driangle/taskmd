import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import type { WorklogEntry } from "../../api/types.ts";

interface WorklogSectionProps {
  entries: WorklogEntry[];
}

export function WorklogSection({ entries }: WorklogSectionProps) {
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
          <div key={i} className="border-l-2 border-gray-200 dark:border-gray-600 pl-4">
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
      </div>
    </div>
  );
}
