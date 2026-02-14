import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { Recommendation } from "../api/types.ts";

export function useNext(limit: number = 5) {
  return useSWR<Recommendation[]>(`/api/next?limit=${limit}`, fetcher);
}
