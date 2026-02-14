export const STATUSES = ["pending", "in-progress", "completed", "blocked", "cancelled"];
export const PRIORITIES = ["critical", "high", "medium", "low"];
export const EFFORTS = ["small", "medium", "large"];

export const STATUS_COLORS: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800 font-medium ring-1 ring-yellow-300 dark:bg-yellow-900/30 dark:text-yellow-300 dark:ring-yellow-700",
  "in-progress": "bg-blue-100 text-blue-800 font-medium ring-1 ring-blue-300 dark:bg-blue-900/30 dark:text-blue-300 dark:ring-blue-700",
  completed: "bg-green-100 text-green-800 font-medium ring-1 ring-green-300 dark:bg-green-900/30 dark:text-green-300 dark:ring-green-700",
  blocked: "bg-red-100 text-red-800 font-medium ring-1 ring-red-300 dark:bg-red-900/30 dark:text-red-300 dark:ring-red-700",
  cancelled: "bg-gray-100 text-gray-600 font-medium ring-1 ring-gray-300 dark:bg-gray-700/50 dark:text-gray-400 dark:ring-gray-600",
};

export const PRIORITY_COLORS: Record<string, string> = {
  critical: "bg-red-100 text-red-600 font-medium ring-1 ring-red-300 dark:bg-red-900/30 dark:text-red-400 dark:ring-red-700",
  high: "bg-orange-100 text-orange-600 font-medium ring-1 ring-orange-300 dark:bg-orange-900/30 dark:text-orange-400 dark:ring-orange-700",
  medium: "bg-gray-100 text-gray-600 font-medium ring-1 ring-gray-300 dark:bg-gray-700/50 dark:text-gray-400 dark:ring-gray-600",
  low: "bg-gray-50 text-gray-400 font-medium ring-1 ring-gray-200 dark:bg-gray-800/50 dark:text-gray-500 dark:ring-gray-700",
};

export const EFFORT_COLORS: Record<string, string> = {
  small: "bg-emerald-100 text-emerald-700 font-medium ring-1 ring-emerald-300 dark:bg-emerald-900/30 dark:text-emerald-400 dark:ring-emerald-700",
  medium: "bg-amber-100 text-amber-700 font-medium ring-1 ring-amber-300 dark:bg-amber-900/30 dark:text-amber-400 dark:ring-amber-700",
  large: "bg-purple-100 text-purple-700 font-medium ring-1 ring-purple-300 dark:bg-purple-900/30 dark:text-purple-400 dark:ring-purple-700",
};
