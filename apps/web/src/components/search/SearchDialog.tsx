import { useState, useEffect, useRef, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useSearch } from "../../hooks/use-search.ts";
import { STATUS_COLORS } from "../tasks/TaskTable/constants.ts";

interface SearchDialogProps {
  open: boolean;
  onClose: () => void;
}

export function SearchDialog({ open, onClose }: SearchDialogProps) {
  const [query, setQuery] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();
  const { data: results, isLoading } = useSearch(query);

  useEffect(() => {
    if (open) {
      setQuery("");
      // Small delay to ensure the dialog is rendered before focusing
      requestAnimationFrame(() => inputRef.current?.focus());
    }
  }, [open]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Escape") {
        onClose();
      }
    },
    [onClose],
  );

  const handleSelect = useCallback(
    (id: string) => {
      onClose();
      navigate(`/tasks/${id}`);
    },
    [onClose, navigate],
  );

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]"
      onClick={onClose}
      onKeyDown={handleKeyDown}
    >
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50" />

      {/* Dialog */}
      <div
        className="relative w-full max-w-lg mx-4 bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Search input */}
        <div className="flex items-center gap-3 px-4 border-b border-gray-200 dark:border-gray-700">
          <svg
            className="w-4 h-4 text-gray-400 shrink-0"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search tasks..."
            className="flex-1 py-3 bg-transparent text-sm text-gray-900 dark:text-gray-100 placeholder-gray-400 outline-none"
          />
          <kbd className="hidden sm:inline-block px-1.5 py-0.5 text-[10px] font-medium text-gray-400 bg-gray-100 dark:bg-gray-700 rounded">
            ESC
          </kbd>
        </div>

        {/* Results */}
        <div className="max-h-80 overflow-y-auto">
          {!query && (
            <p className="px-4 py-8 text-center text-sm text-gray-400">
              Start typing to search tasks...
            </p>
          )}

          {query && isLoading && (
            <p className="px-4 py-8 text-center text-sm text-gray-400">
              Searching...
            </p>
          )}

          {query && !isLoading && results?.length === 0 && (
            <p className="px-4 py-8 text-center text-sm text-gray-400">
              No results found for &ldquo;{query}&rdquo;
            </p>
          )}

          {results && results.length > 0 && (
            <ul className="py-2">
              {results.map((result) => (
                <li key={result.id}>
                  <button
                    type="button"
                    className="w-full px-4 py-2.5 text-left hover:bg-gray-100 dark:hover:bg-gray-700/50 transition-colors"
                    onClick={() => handleSelect(result.id)}
                  >
                    <div className="flex items-center gap-2 mb-0.5">
                      <span className="text-xs font-mono text-gray-400">
                        #{result.id}
                      </span>
                      <span className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                        <Highlight text={result.title} query={query} />
                      </span>
                      <span
                        className={`ml-auto shrink-0 px-2 py-0.5 text-xs font-medium rounded-full ${STATUS_COLORS[result.status] ?? "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300"}`}
                      >
                        {result.status}
                      </span>
                    </div>
                    {result.snippet && (
                      <p className="text-xs text-gray-500 dark:text-gray-400 truncate">
                        <span className="text-gray-400 dark:text-gray-500 mr-1">
                          {result.match_location}:
                        </span>
                        <Highlight text={result.snippet} query={query} />
                      </p>
                    )}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}

function Highlight({ text, query }: { text: string; query: string }) {
  if (!query) return <>{text}</>;

  const lowerText = text.toLowerCase();
  const lowerQuery = query.toLowerCase();
  const idx = lowerText.indexOf(lowerQuery);

  if (idx < 0) return <>{text}</>;

  const before = text.slice(0, idx);
  const match = text.slice(idx, idx + query.length);
  const after = text.slice(idx + query.length);

  return (
    <>
      {before}
      <mark className="bg-yellow-200 dark:bg-yellow-900/60 text-inherit rounded-sm px-0.5">
        {match}
      </mark>
      {after}
    </>
  );
}
