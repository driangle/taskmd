import { useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import mermaid from "mermaid";

mermaid.initialize({
  startOnLoad: false,
  theme: "default",
  securityLevel: "loose",
});

interface GraphViewProps {
  mermaidSyntax: string;
}

export function GraphView({ mermaidSyntax }: GraphViewProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (!containerRef.current || !mermaidSyntax) return;

    const id = `mermaid-${Date.now()}`;
    mermaid.render(id, mermaidSyntax).then(({ svg }) => {
      if (containerRef.current) {
        containerRef.current.innerHTML = svg;

        // Add click handlers to graph nodes
        const nodes = containerRef.current.querySelectorAll(".node");
        nodes.forEach((node) => {
          // Extract task ID from the node's id attribute
          // Mermaid generates ids like "flowchart-task-001-123"
          const nodeId = node.getAttribute("id");
          const match = nodeId?.match(/task-(\d+)/);
          if (match) {
            const taskId = match[1];

            // Make node clickable
            const nodeElement = node as HTMLElement;
            nodeElement.style.cursor = "pointer";

            // Add click handler
            nodeElement.addEventListener("click", () => {
              navigate(`/tasks/${taskId}`);
            });

            // Add hover effect
            nodeElement.addEventListener("mouseenter", () => {
              const rect = nodeElement.querySelector("rect, polygon");
              if (rect) {
                (rect as SVGElement).style.opacity = "0.8";
              }
            });

            nodeElement.addEventListener("mouseleave", () => {
              const rect = nodeElement.querySelector("rect, polygon");
              if (rect) {
                (rect as SVGElement).style.opacity = "1";
              }
            });
          }
        });
      }
    });
  }, [mermaidSyntax, navigate]);

  return (
    <div
      ref={containerRef}
      className="bg-white rounded-lg border border-gray-200 p-4 overflow-auto"
    />
  );
}
