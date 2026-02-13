export const STATUSES = ["pending", "in-progress", "completed", "blocked", "cancelled"];
export const PRIORITIES = ["critical", "high", "medium", "low"];
export const EFFORTS = ["small", "medium", "large"];

export const STATUS_COLORS: Record<string, string> = {
  pending: "bg-yellow-100 text-yellow-800",
  "in-progress": "bg-blue-100 text-blue-800",
  completed: "bg-green-100 text-green-800",
  blocked: "bg-red-100 text-red-800",
  cancelled: "bg-gray-100 text-gray-600",
};

export const PRIORITY_COLORS: Record<string, string> = {
  critical: "bg-red-100 text-red-600",
  high: "bg-orange-100 text-orange-600",
  medium: "bg-gray-100 text-gray-600",
  low: "bg-gray-50 text-gray-400",
};
