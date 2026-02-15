import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { SearchResult } from "../api/types.ts";

export function useSearch(query: string) {
  return useSWR<SearchResult[]>(
    query ? `/api/search?q=${encodeURIComponent(query)}` : null,
    fetcher,
  );
}
