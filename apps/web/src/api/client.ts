import type { ApiError, Task, TaskUpdateRequest } from "./types.ts";

export async function fetcher<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
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
