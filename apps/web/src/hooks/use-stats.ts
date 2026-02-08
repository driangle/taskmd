import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { Stats } from "../api/types.ts";

export function useStats() {
  return useSWR<Stats>("/api/stats", fetcher);
}
