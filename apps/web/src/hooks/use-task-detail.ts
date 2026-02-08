import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { Task } from "../api/types.ts";

export function useTaskDetail(taskId: string | undefined) {
  return useSWR<Task>(taskId ? `/api/tasks/${taskId}` : null, fetcher);
}
