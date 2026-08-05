import type {
  ApiError,
  Task,
  TaskUpdateRequest,
  WorklogEntry,
} from "./types.ts";

export async function fetcher<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? `API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

export class ApiRequestError extends Error {
  details: string[];

  constructor(message: string, details: string[] = []) {
    super(message);
    this.name = "ApiRequestError";
    this.details = details;
  }
}

export async function updateTask(
  id: string,
  data: TaskUpdateRequest,
): Promise<Task> {
  const res = await fetch(`/api/tasks/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });

  if (!res.ok) {
    const body: ApiError = await res.json().catch(() => ({
      error: `HTTP ${res.status}`,
    }));
    throw new ApiRequestError(body.error, body.details);
  }

  return res.json();
}

export async function addWorklogEntry(
  id: string,
  data: { author?: string; content: string },
): Promise<WorklogEntry[]> {
  const res = await fetch(`/api/tasks/${id}/worklog`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });

  if (!res.ok) {
    const body: ApiError = await res.json().catch(() => ({
      error: `HTTP ${res.status}`,
    }));
    throw new ApiRequestError(body.error, body.details);
  }

  return res.json();
}
