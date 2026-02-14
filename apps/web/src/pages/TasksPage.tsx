import { useSearchParams } from "react-router-dom";
import { useTasks } from "../hooks/use-tasks.ts";
import { TaskTable } from "../components/tasks/TaskTable.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";

export function TasksPage() {
  const [searchParams] = useSearchParams();
  const { data, error, isLoading, mutate } = useTasks();

  if (isLoading) return <LoadingState variant="table" />;
  if (error) return <ErrorState error={error} onRetry={() => mutate()} />;
  if (!data) return null;

  if (data.length === 0) {
    return (
      <p className="text-sm text-gray-500 py-8 text-center">
        No tasks found.
      </p>
    );
  }

  const initialTags = searchParams.getAll("tag");

  return <TaskTable tasks={data} initialTags={initialTags} />;
}
