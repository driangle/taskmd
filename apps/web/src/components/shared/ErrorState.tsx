interface ErrorStateProps {
  error: Error;
  onRetry?: () => void;
}

function isConnectionError(error: Error): boolean {
  const msg = error.message.toLowerCase();
  return (
    error instanceof TypeError &&
    (msg.includes("fetch") || msg.includes("network"))
  ) || msg.includes("failed to fetch");
}

export function ErrorState({ error, onRetry }: ErrorStateProps) {
  const connectionError = isConnectionError(error);

  return (
    <div className="rounded-lg border border-red-200 bg-red-50 p-4 max-w-lg dark:border-red-800 dark:bg-red-900/20">
      <div className="flex items-start gap-3">
        <div className="text-red-500 dark:text-red-400 mt-0.5 shrink-0">
          <svg
            className="h-5 w-5"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z"
              clipRule="evenodd"
            />
          </svg>
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-medium text-red-800 dark:text-red-300">
            {connectionError
              ? "Cannot connect to server"
              : "Something went wrong"}
          </h3>
          <p className="mt-1 text-sm text-red-700 dark:text-red-400">
            {connectionError
              ? "The taskmd server is not reachable. Make sure it's running and try again."
              : error.message}
          </p>
          {onRetry && (
            <button
              onClick={onRetry}
              className="mt-3 px-3 py-1.5 text-sm font-medium text-red-700 bg-white border border-red-300 rounded-md hover:bg-red-50 transition-colors dark:text-red-300 dark:bg-gray-800 dark:border-red-700 dark:hover:bg-gray-700"
            >
              Retry
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
