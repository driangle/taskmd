import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { TracksResult } from "../api/types.ts";

export function useTracks() {
  return useSWR<TracksResult>("/api/tracks", fetcher);
}
