import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import type { ValidationResult } from "../api/types.ts";

export function useValidate() {
  return useSWR<ValidationResult>("/api/validate", fetcher);
}
