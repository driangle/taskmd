import { useRef, useEffect } from "react";
import { useReactFlow, type Viewport } from "@xyflow/react";

interface GraphSearchProps {
  query: string;
  onQueryChange: (query: string) => void;
  matchedNodeIds: Set<string>;
}

export function GraphSearch({ query, onQueryChange, matchedNodeIds }: GraphSearchProps) {
  const { fitView, getViewport, setViewport } = useReactFlow();
  const savedViewport = useRef<Viewport | null>(null);
  const prevQueryEmpty = useRef(true);

  useEffect(() => {
    const wasEmpty = prevQueryEmpty.current;
    const isEmpty = query === "";
    prevQueryEmpty.current = isEmpty;

    if (isEmpty) {
      if (!wasEmpty && savedViewport.current) {
        setViewport(savedViewport.current, { duration: 300 });
        savedViewport.current = null;
      }
      return;
    }

    if (wasEmpty) {
      savedViewport.current = getViewport();
    }

    if (matchedNodeIds.size > 0) {
      fitView({
        nodes: Array.from(matchedNodeIds).map((id) => ({ id })),
        padding: 0.3,
        maxZoom: 0.85,
        duration: 300,
      });
    }
  }, [query, matchedNodeIds, fitView, getViewport, setViewport]);

  return (
    <div className="relative flex items-center gap-2">
      <input
        type="text"
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
        placeholder="Search tasks..."
        className="w-48 px-2.5 py-1 text-xs border border-gray-200 rounded-full bg-white focus:outline-none focus:border-blue-400 focus:ring-1 focus:ring-blue-400"
      />
      {query && (
        <>
          <span className="text-[10px] text-gray-500">
            {matchedNodeIds.size} found
          </span>
          <button
            onClick={() => onQueryChange("")}
            className="text-xs text-gray-400 hover:text-gray-600"
            aria-label="Clear search"
          >
            &times;
          </button>
        </>
      )}
    </div>
  );
}
