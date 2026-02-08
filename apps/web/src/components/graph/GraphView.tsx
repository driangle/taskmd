import { useEffect, useRef } from "react";
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

  useEffect(() => {
    if (!containerRef.current || !mermaidSyntax) return;

    const id = `mermaid-${Date.now()}`;
    mermaid.render(id, mermaidSyntax).then(({ svg }) => {
      if (containerRef.current) {
        containerRef.current.innerHTML = svg;
      }
    });
  }, [mermaidSyntax]);

  return (
    <div
      ref={containerRef}
      className="bg-white rounded-lg border border-gray-200 p-4 overflow-auto"
    />
  );
}
