import { useMemo } from "react";
import dagre from "dagre";
import type { Node, Edge } from "@xyflow/react";
import type { GraphData } from "../../api/types.ts";

const NODE_WIDTH = 200;
const NODE_HEIGHT = 60;

export function useGraphLayout(data: GraphData | undefined) {
  return useMemo(() => {
    if (!data || data.nodes.length === 0) {
      return { nodes: [] as Node[], edges: [] as Edge[] };
    }

    const g = new dagre.graphlib.Graph();
    g.setDefaultEdgeLabel(() => ({}));
    g.setGraph({ rankdir: "TB", ranksep: 80, nodesep: 40 });

    for (const node of data.nodes) {
      g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
    }

    for (const edge of data.edges) {
      g.setEdge(edge.from, edge.to);
    }

    dagre.layout(g);

    const nodes: Node[] = data.nodes.map((node) => {
      const pos = g.node(node.id);
      return {
        id: node.id,
        type: "task",
        position: { x: pos.x - NODE_WIDTH / 2, y: pos.y - NODE_HEIGHT / 2 },
        data: { label: node.title, status: node.status, priority: node.priority, taskId: node.id },
      };
    });

    const edges: Edge[] = data.edges.map((edge, i) => ({
      id: `e-${i}`,
      source: edge.from,
      target: edge.to,
      type: "smoothstep",
      markerEnd: { type: "arrowclosed" as const },
    }));

    return { nodes, edges };
  }, [data]);
}
