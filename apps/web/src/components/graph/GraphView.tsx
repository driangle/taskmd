import { useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import {
  ReactFlow,
  Controls,
  Background,
  BackgroundVariant,
  type NodeTypes,
  type NodeMouseHandler,
  type Node,
  type Edge,
  type Viewport,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { TaskNode } from "./TaskNode.tsx";

const nodeTypes: NodeTypes = { task: TaskNode };
const fitViewOptions = { maxZoom: 0.85, padding: 0.15 };

interface GraphViewProps {
  nodes: Node[];
  edges: Edge[];
  defaultViewport?: Viewport;
  onViewportChange?: (viewport: Viewport) => void;
  matchedNodeIds?: Set<string>;
  searchActive?: boolean;
}

export function GraphView({
  nodes,
  edges,
  defaultViewport,
  onViewportChange,
  matchedNodeIds,
  searchActive,
}: GraphViewProps) {
  const navigate = useNavigate();

  const onNodeClick: NodeMouseHandler = useCallback(
    (_event, node) => {
      navigate(`/tasks/${node.data.taskId}`);
    },
    [navigate],
  );

  const decoratedNodes = useMemo(() => {
    if (!searchActive || !matchedNodeIds) return nodes;
    return nodes.map((node) => ({
      ...node,
      data: {
        ...node.data,
        highlighted: matchedNodeIds.has(node.id),
        dimmed: !matchedNodeIds.has(node.id),
      },
    }));
  }, [nodes, matchedNodeIds, searchActive]);

  if (nodes.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-sm text-gray-500">
        No tasks to display
      </div>
    );
  }

  const hasRestoredViewport = defaultViewport !== undefined;

  return (
    <ReactFlow
      nodes={decoratedNodes}
      edges={edges}
      nodeTypes={nodeTypes}
      onNodeClick={onNodeClick}
      fitView={!hasRestoredViewport}
      fitViewOptions={fitViewOptions}
      defaultViewport={hasRestoredViewport ? defaultViewport : undefined}
      onViewportChange={onViewportChange}
      minZoom={0.1}
      maxZoom={2}
      proOptions={{ hideAttribution: true }}
    >
      <Controls position="bottom-right" />
      <Background variant={BackgroundVariant.Dots} gap={16} size={1} color="#e5e7eb" />
    </ReactFlow>
  );
}
