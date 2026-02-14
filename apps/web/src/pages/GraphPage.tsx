import { useState, useMemo, useCallback } from "react";
import { ReactFlowProvider } from "@xyflow/react";
import { useGraph } from "../hooks/use-graph.ts";
import { GraphView } from "../components/graph/GraphView.tsx";
import { GraphFilters } from "../components/graph/GraphFilters.tsx";
import { GraphStats } from "../components/graph/GraphStats.tsx";
import { GraphSearch } from "../components/graph/GraphSearch.tsx";
import { GraphLegend } from "../components/graph/GraphLegend.tsx";
import { useGraphLayout } from "../components/graph/useGraphLayout.ts";
import { LoadingState } from "../components/shared/LoadingState.tsx";
import { ErrorState } from "../components/shared/ErrorState.tsx";
import type { GraphData } from "../api/types.ts";
import type { Viewport } from "@xyflow/react";

// Persist graph state across navigations (module-level, survives unmount)
const savedState = {
  statuses: new Set<string>(),
  viewport: undefined as Viewport | undefined,
};

export function GraphPage() {
  const { data, error, isLoading, mutate } = useGraph();
  const [selectedStatuses, setSelectedStatuses] = useState<Set<string>>(savedState.statuses);
  const [searchQuery, setSearchQuery] = useState("");

  const matchedNodeIds = useMemo(() => {
    if (!data || searchQuery === "") return new Set<string>();
    const q = searchQuery.toLowerCase();
    return new Set(
      data.nodes
        .filter((n) => n.id.toLowerCase().includes(q) || n.title.toLowerCase().includes(q))
        .map((n) => n.id),
    );
  }, [data, searchQuery]);

  const filteredData = useMemo((): GraphData | undefined => {
    if (!data) return undefined;
    if (selectedStatuses.size === 0) return data;

    const visibleNodes = data.nodes.filter((n) => selectedStatuses.has(n.status));
    const visibleIds = new Set(visibleNodes.map((n) => n.id));
    const visibleEdges = data.edges.filter((e) => visibleIds.has(e.from) && visibleIds.has(e.to));

    return { nodes: visibleNodes, edges: visibleEdges, cycles: data.cycles };
  }, [data, selectedStatuses]);

  const { nodes, edges } = useGraphLayout(filteredData);

  const toggleStatus = useCallback((status: string) => {
    setSelectedStatuses((prev) => {
      const next = new Set(prev);
      if (next.has(status)) {
        next.delete(status);
      } else {
        next.add(status);
      }
      savedState.statuses = next;
      return next;
    });
  }, []);

  const clearFilters = useCallback(() => {
    const empty = new Set<string>();
    savedState.statuses = empty;
    setSelectedStatuses(empty);
  }, []);

  const onViewportChange = useCallback((viewport: Viewport) => {
    savedState.viewport = viewport;
  }, []);

  if (isLoading) return <LoadingState variant="graph" />;
  if (error) return <ErrorState error={error} onRetry={() => mutate()} />;
  if (!data) return null;

  if (data.nodes.length === 0) {
    return (
      <p className="text-sm text-gray-500 dark:text-gray-400 py-8 text-center">
        No dependencies to display.
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
        <div className="flex items-center gap-4 flex-wrap">
          <GraphFilters
            selectedStatuses={selectedStatuses}
            onToggleStatus={toggleStatus}
            onClearFilters={clearFilters}
          />
          <GraphStats data={data} visibleCount={nodes.length} />
        </div>
      </div>
      <ReactFlowProvider>
        <div className="relative bg-white rounded-lg border border-gray-200 h-[calc(100vh-200px)] min-h-[500px] dark:bg-gray-800 dark:border-gray-700">
          <div className="absolute top-3 left-3 z-10">
            <GraphSearch
              query={searchQuery}
              onQueryChange={setSearchQuery}
              matchedNodeIds={matchedNodeIds}
            />
          </div>
          <GraphLegend />
          <GraphView
            nodes={nodes}
            edges={edges}
            defaultViewport={savedState.viewport}
            onViewportChange={onViewportChange}
            matchedNodeIds={matchedNodeIds}
            searchActive={searchQuery !== ""}
          />
        </div>
      </ReactFlowProvider>
    </div>
  );
}
