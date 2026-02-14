import { useValidate } from "../hooks/use-validate.ts";
import { ValidateView } from "../components/validate/ValidateView.tsx";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";

export function ValidatePage() {
  const { data, error, isLoading, mutate } = useValidate();

  if (isLoading) return <LoadingState />;
  if (error) return <ErrorState error={error} onRetry={() => mutate()} />;
  if (!data) return null;

  return <ValidateView result={data} />;
}
