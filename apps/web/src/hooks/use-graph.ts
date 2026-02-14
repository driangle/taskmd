import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { GraphData } from "../api/types.ts";

export function useGraph() {
  return useSWR<GraphData>("/api/graph", fetcher);
}
