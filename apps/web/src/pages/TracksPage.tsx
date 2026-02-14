import { useState } from "react";
import { useTracks } from "../hooks/use-tracks.ts";
import { TracksView } from "../components/tracks/TracksView.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";

export function TracksPage() {
  const [limit, setLimit] = useState(0);
  const { data, error, isLoading, mutate } = useTracks(limit);

  if (isLoading) return <LoadingState variant="board" />;
  if (error) return <ErrorState error={error} onRetry={() => mutate()} />;
  if (!data) return null;

  return <TracksView data={data} limit={limit} onLimitChange={setLimit} />;
}
