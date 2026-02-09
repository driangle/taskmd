import { STATUS_COLORS, PRIORITY_COLORS } from "./constants.ts";

export function StatusBadge({ status }: { status: string }) {
  return (
    <span
      className={`px-2 py-0.5 text-xs font-medium rounded-full ${STATUS_COLORS[status] ?? "bg-gray-100 text-gray-800"}`}
    >
      {status}
    </span>
  );
}

export function PriorityBadge({ priority }: { priority: string }) {
  return (
    <span
      className={`px-2 py-0.5 text-xs font-medium rounded-full ${PRIORITY_COLORS[priority] ?? "bg-gray-100 text-gray-500"}`}
    >
      {priority}
    </span>
  );
}

export function BlockedStatusBadge({
  dependencies,
}: {
  dependencies: string[] | null;
}) {
  const blockedByCount = dependencies?.length ?? 0;
  const isBlocked = blockedByCount > 0;

  if (!isBlocked) {
    return (
      <span
        className="px-2 py-0.5 text-xs font-medium rounded-full bg-green-100 text-green-800 inline-flex items-center gap-1"
        aria-label="Task is ready to work on"
      >
        <span aria-hidden="true">✓</span>
        <span className="hidden sm:inline">Ready</span>
      </span>
    );
  }

  const tooltipText = `Blocked by: ${dependencies?.join(", ")}`;

  return (
    <span
      className="px-2 py-0.5 text-xs font-medium rounded-full bg-amber-100 text-amber-800 inline-flex items-center gap-1 cursor-help"
      title={tooltipText}
      aria-label={tooltipText}
    >
      <span aria-hidden="true">⚠</span>
      <span>
        <span className="hidden sm:inline">Blocked </span>({blockedByCount})
      </span>
    </span>
  );
}
