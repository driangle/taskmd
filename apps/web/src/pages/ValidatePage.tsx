import { useValidate } from "../hooks/use-validate.ts";
import { ValidateView } from "../components/validate/ValidateView.tsx";

export function ValidatePage() {
  const { data, error, isLoading } = useValidate();

  if (isLoading) return <p className="text-sm text-gray-500">Loading...</p>;
  if (error)
    return <p className="text-sm text-red-600">Error: {error.message}</p>;
  if (!data) return null;

  return <ValidateView result={data} />;
}
