import { useEffect } from "react";
import { useSWRConfig } from "swr";

export function useLiveReload() {
  const { mutate } = useSWRConfig();

  useEffect(() => {
    const es = new EventSource("/api/events");

    es.addEventListener("reload", () => {
      // Revalidate all SWR caches
      mutate(() => true, undefined, { revalidate: true });
    });

    es.addEventListener("error", () => {
      // EventSource auto-reconnects
    });

    return () => es.close();
  }, [mutate]);
}
