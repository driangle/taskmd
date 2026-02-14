import { useState } from "react";
import { useNext } from "../hooks/use-next.ts";
import { NextView } from "../components/next/NextView.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";

export function NextPage() {
  const [limit, setLimit] = useState(5);
  const { data, error, isLoading, mutate } = useNext(limit);

  if (isLoading) return <LoadingState variant="cards" />;
  if (error) return <ErrorState error={error} onRetry={() => mutate()} />;
  if (!data) return null;

  if (data.length === 0) {
    return (
      <p className="text-sm text-gray-500 py-8 text-center">
        No actionable tasks found. All tasks are either completed or blocked.
      </p>
    );
  }

  return (
    <NextView recommendations={data} limit={limit} onLimitChange={setLimit} />
  );
}
