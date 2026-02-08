import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { BoardGroup } from "../api/types.ts";

export function useBoard(groupBy: string = "status") {
  return useSWR<BoardGroup[]>(`/api/board?groupBy=${groupBy}`, fetcher);
}
