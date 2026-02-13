import { useSearchParams } from "react-router-dom";
import { useTasks } from "../hooks/use-tasks.ts";
import { TaskTable } from "../components/tasks/TaskTable.tsx";

export function TasksPage() {
  const [searchParams] = useSearchParams();
  const { data, error, isLoading } = useTasks();

  if (isLoading) return <p className="text-sm text-gray-500">Loading...</p>;
  if (error)
    return <p className="text-sm text-red-600">Error: {error.message}</p>;
  if (!data) return null;

  const initialTags = searchParams.getAll("tag");

  return <TaskTable tasks={data} initialTags={initialTags} />;
}
