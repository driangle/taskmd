import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { TracksResult } from "../api/types.ts";

export function useTracks(limit: number = 0) {
  const params = limit > 0 ? `?limit=${limit}` : "";
  return useSWR<TracksResult>(`/api/tracks${params}`, fetcher);
}
