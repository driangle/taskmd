import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { Task } from "../api/types.ts";

export function useTasks() {
  return useSWR<Task[]>("/api/tasks", fetcher);
}
