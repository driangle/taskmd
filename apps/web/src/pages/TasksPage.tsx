import { useSearchParams } from "react-router-dom";
import { useTasks } from "../hooks/use-tasks.ts";
import { usePhase } from "../hooks/use-phase.tsx";
import { useProject } from "../hooks/use-project.ts";
import { TaskTable } from "../components/tasks/TaskTable.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";

export function TasksPage() {
  const [searchParams] = useSearchParams();
  const { phase } = usePhase();
  const { project } = useProject();
  const { data, error, isLoading, mutate } = useTasks(phase, project);

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
  const initialStatuses = searchParams.getAll("status");
  const initialPriorities = searchParams.getAll("priority");
  const initialEffort = searchParams.getAll("effort");
  const initialTypes = searchParams.getAll("type");
  const initialPhases = searchParams.getAll("phase");

  return (
    <TaskTable
      tasks={data}
      initialTags={initialTags}
      initialStatuses={initialStatuses}
      initialPriorities={initialPriorities}
      initialEffort={initialEffort}
      initialTypes={initialTypes}
      initialPhases={initialPhases}
      project={project}
    />
  );
}
