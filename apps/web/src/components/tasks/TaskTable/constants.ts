export const STATUSES = ["pending", "in-progress", "in-review", "completed", "blocked", "cancelled"];
export const PRIORITIES = ["critical", "high", "medium", "low"];
/**
 * Fallback effort vocabulary, used when the server does not report one.
 * The live vocabulary comes from `/api/config` via `useConfig()` — see
 * `effort-colors.ts` for how custom vocabularies are colored.
 */
export const DEFAULT_EFFORTS = ["small", "medium", "large"];
export const TYPES = ["feature", "bug", "improvement", "chore", "docs"];

export const STATUS_COLORS: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800 font-medium ring-1 ring-yellow-300 dark:bg-yellow-900/30 dark:text-yellow-300 dark:ring-yellow-700",
  "in-progress": "bg-blue-100 text-blue-800 font-medium ring-1 ring-blue-300 dark:bg-blue-900/30 dark:text-blue-300 dark:ring-blue-700",
  "in-review": "bg-purple-100 text-purple-800 font-medium ring-1 ring-purple-300 dark:bg-purple-900/30 dark:text-purple-300 dark:ring-purple-700",
  completed: "bg-green-100 text-green-800 font-medium ring-1 ring-green-300 dark:bg-green-900/30 dark:text-green-300 dark:ring-green-700",
  blocked: "bg-red-100 text-red-800 font-medium ring-1 ring-red-300 dark:bg-red-900/30 dark:text-red-300 dark:ring-red-700",
  cancelled: "bg-gray-100 text-gray-600 font-medium ring-1 ring-gray-300 dark:bg-gray-700/50 dark:text-gray-400 dark:ring-gray-600",
};

export const PRIORITY_COLORS: Record<string, string> = {
  critical: "bg-red-100 text-red-600 font-medium ring-1 ring-red-300 dark:bg-red-900/30 dark:text-red-400 dark:ring-red-700",
  high: "bg-orange-100 text-orange-600 font-medium ring-1 ring-orange-300 dark:bg-orange-900/30 dark:text-orange-400 dark:ring-orange-700",
  medium: "bg-indigo-100 text-indigo-600 font-medium ring-1 ring-indigo-300 dark:bg-indigo-900/30 dark:text-indigo-400 dark:ring-indigo-700",
  low: "bg-sky-50 text-sky-500 font-medium ring-1 ring-sky-200 dark:bg-sky-900/30 dark:text-sky-400 dark:ring-sky-700",
};

export const EFFORT_COLORS: Record<string, string> = {
  small: "bg-emerald-100 text-emerald-700 font-medium ring-1 ring-emerald-300 dark:bg-emerald-900/30 dark:text-emerald-400 dark:ring-emerald-700",
  medium: "bg-amber-100 text-amber-700 font-medium ring-1 ring-amber-300 dark:bg-amber-900/30 dark:text-amber-400 dark:ring-amber-700",
  large: "bg-purple-100 text-purple-700 font-medium ring-1 ring-purple-300 dark:bg-purple-900/30 dark:text-purple-400 dark:ring-purple-700",
};

export const TYPE_COLORS: Record<string, string> = {
  feature: "bg-violet-100 text-violet-700 font-medium ring-1 ring-violet-300 dark:bg-violet-900/30 dark:text-violet-400 dark:ring-violet-700",
  bug: "bg-red-100 text-red-700 font-medium ring-1 ring-red-300 dark:bg-red-900/30 dark:text-red-400 dark:ring-red-700",
  improvement: "bg-cyan-100 text-cyan-700 font-medium ring-1 ring-cyan-300 dark:bg-cyan-900/30 dark:text-cyan-400 dark:ring-cyan-700",
  chore: "bg-slate-100 text-slate-600 font-medium ring-1 ring-slate-300 dark:bg-slate-900/30 dark:text-slate-400 dark:ring-slate-600",
  docs: "bg-indigo-100 text-indigo-700 font-medium ring-1 ring-indigo-300 dark:bg-indigo-900/30 dark:text-indigo-400 dark:ring-indigo-700",
};

/**
 * Color palette for phase badges. Each entry provides light/dark mode classes.
 * A phase gets a consistent color by hashing its name to an index.
 */
const PHASE_PALETTE = [
  "bg-teal-100 text-teal-700 ring-1 ring-teal-300 dark:bg-teal-900/30 dark:text-teal-300 dark:ring-teal-700",
  "bg-fuchsia-100 text-fuchsia-700 ring-1 ring-fuchsia-300 dark:bg-fuchsia-900/30 dark:text-fuchsia-300 dark:ring-fuchsia-700",
  "bg-lime-100 text-lime-700 ring-1 ring-lime-300 dark:bg-lime-900/30 dark:text-lime-300 dark:ring-lime-700",
  "bg-rose-100 text-rose-700 ring-1 ring-rose-300 dark:bg-rose-900/30 dark:text-rose-300 dark:ring-rose-700",
  "bg-sky-100 text-sky-700 ring-1 ring-sky-300 dark:bg-sky-900/30 dark:text-sky-300 dark:ring-sky-700",
  "bg-amber-100 text-amber-700 ring-1 ring-amber-300 dark:bg-amber-900/30 dark:text-amber-300 dark:ring-amber-700",
  "bg-violet-100 text-violet-700 ring-1 ring-violet-300 dark:bg-violet-900/30 dark:text-violet-300 dark:ring-violet-700",
  "bg-emerald-100 text-emerald-700 ring-1 ring-emerald-300 dark:bg-emerald-900/30 dark:text-emerald-300 dark:ring-emerald-700",
];

function hashString(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash + str.charCodeAt(i)) | 0;
  }
  return Math.abs(hash);
}

export function getPhaseColor(phase: string): string {
  return PHASE_PALETTE[hashString(phase) % PHASE_PALETTE.length];
}
