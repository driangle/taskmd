import useSWR from "swr";
import { fetcher } from "../api/client.ts";
import { DEFAULT_EFFORTS } from "../components/tasks/TaskTable/constants.ts";

export interface PhaseInfo {
  id: string;
  name: string;
  description: string;
}

export interface WorktreeOverlayInfo {
  /** Local worktree's directory name; may be absent. */
  name?: string;
  /** Number of sibling worktrees merged into the served view. */
  siblings: number;
}

interface AppConfig {
  readonly: boolean;
  version: string;
  phases: PhaseInfo[];
  /** Project's effort vocabulary. Absent on servers older than this key. */
  efforts?: string[];
  /** Active worktree overlay; absent in single-worktree repos. */
  worktree?: WorktreeOverlayInfo;
}

declare global {
  interface Window {
    __TASKMD_CONFIG__?: AppConfig;
  }
}

// Read config injected by the Go server into the HTML at build time.
// Available immediately on first render — no async fetch needed for the default config.
const injectedConfig: AppConfig | undefined = window.__TASKMD_CONFIG__;

export interface UseConfigResult {
  readonly: boolean;
  version: string;
  phases: PhaseInfo[];
  efforts: string[];
  /** Active worktree overlay; absent in single-worktree repos. */
  worktree?: WorktreeOverlayInfo;
}

export function useConfig(project?: string | null): UseConfigResult {
  const params = new URLSearchParams();
  if (project) params.set("project", project);
  const qs = params.toString();
  const isDefault = !project;
  const { data } = useSWR<AppConfig>(`/api/config${qs ? `?${qs}` : ""}`, fetcher, {
    revalidateOnFocus: false,
    fallbackData: isDefault ? injectedConfig : undefined,
  });
  return {
    readonly: data?.readonly ?? false,
    version: data?.version ?? "",
    phases: data?.phases ?? [],
    efforts: data?.efforts?.length ? data.efforts : DEFAULT_EFFORTS,
    worktree: data?.worktree,
  };
}
