import type { Task } from "../../../api/types.ts";

/**
 * The status the merged worktree view reports for a task, falling back to the
 * local copy's status when the overlay is inactive. Mirrors the CLI's list
 * rendering: display and filtering use the effective status.
 */
export function effectiveStatus(task: Task): string {
  return task.effective_status || task.status;
}

export function toggleInSet<T>(set: Set<T>, value: T): Set<T> {
  const next = new Set(set);
  if (next.has(value)) {
    next.delete(value);
  } else {
    next.add(value);
  }
  return next;
}
