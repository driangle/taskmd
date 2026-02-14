import useSWR from "swr";
import { fetcher } from "../api/client.ts";

interface AppConfig {
  readonly: boolean;
}

export function useConfig() {
  const { data } = useSWR<AppConfig>("/api/config", fetcher, {
    revalidateOnFocus: false,
  });
  return { readonly: data?.readonly ?? false };
}
