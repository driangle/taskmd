import { useStats } from "../hooks/use-stats.ts";
import { StatsView } from "../components/stats/StatsView.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";

export function StatsPage() {
  const { data, error, isLoading, mutate } = useStats();

  if (isLoading) return <LoadingState variant="cards" />;
  if (error) return <ErrorState error={error} onRetry={() => mutate()} />;
  if (!data) return null;

  if (data.total_tasks === 0) {
    return (
      <p className="text-sm text-gray-500 py-8 text-center">
        No tasks found to show statistics.
      </p>
    );
  }

  return <StatsView stats={data} />;
}
