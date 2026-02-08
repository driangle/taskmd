import { useTasks } from "../hooks/use-tasks.ts";
import { TaskTable } from "../components/tasks/TaskTable.tsx";

export function TasksPage() {
  const { data, error, isLoading } = useTasks();

  if (isLoading) return <p className="text-sm text-gray-500">Loading...</p>;
  if (error)
    return <p className="text-sm text-red-600">Error: {error.message}</p>;
  if (!data) return null;

  return <TaskTable tasks={data} />;
}
