import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { WorklogEntry } from "../api/types.ts";

export function useWorklog(taskId: string | undefined) {
  return useSWR<WorklogEntry[]>(
    taskId ? `/api/tasks/${taskId}/worklog` : null,
    fetcher,
  );
}
