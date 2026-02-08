import { useGraphMermaid } from "../hooks/use-graph.ts";
import { GraphView } from "../components/graph/GraphView.tsx";

export function GraphPage() {
  const { data, error, isLoading } = useGraphMermaid();

  if (isLoading) return <p className="text-sm text-gray-500">Loading...</p>;
  if (error)
    return <p className="text-sm text-red-600">Error: {error.message}</p>;
  if (!data) return null;

  return <GraphView mermaidSyntax={data} />;
}
