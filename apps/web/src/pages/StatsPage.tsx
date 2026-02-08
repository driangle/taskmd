import { useStats } from "../hooks/use-stats.ts";
import { StatsView } from "../components/stats/StatsView.tsx";

export function StatsPage() {
  const { data, error, isLoading } = useStats();

  if (isLoading) return <p className="text-sm text-gray-500">Loading...</p>;
  if (error)
    return <p className="text-sm text-red-600">Error: {error.message}</p>;
  if (!data) return null;

  return <StatsView stats={data} />;
}
