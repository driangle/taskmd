import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { GraphData } from "../api/types.ts";

export function useGraph() {
  return useSWR<GraphData>("/api/graph", fetcher);
}

export function useGraphMermaid() {
  return useSWR<string>("/api/graph/mermaid", async (url: string) => {
    const res = await fetch(url);
    if (!res.ok) throw new Error(`API error: ${res.status}`);
    return res.text();
  });
}
