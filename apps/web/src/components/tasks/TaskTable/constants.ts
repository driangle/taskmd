export const STATUSES = ["pending", "in-progress", "completed", "blocked", "cancelled"];
export const PRIORITIES = ["critical", "high", "medium", "low"];
export const EFFORTS = ["small", "medium", "large"];

export const STATUS_COLORS: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300",
  "in-progress": "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300",
  completed: "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300",
  blocked: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300",
  cancelled: "bg-gray-100 text-gray-600 dark:bg-gray-700/50 dark:text-gray-400",
};

export const PRIORITY_COLORS: Record<string, string> = {
  critical: "bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400",
  high: "bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400",
  medium: "bg-gray-100 text-gray-600 dark:bg-gray-700/50 dark:text-gray-400",
  low: "bg-gray-50 text-gray-400 dark:bg-gray-800/50 dark:text-gray-500",
};
